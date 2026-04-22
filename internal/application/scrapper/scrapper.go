package scrapper

import (
	"context"
	"fmt"
	"slices"
	"sync"
	"time"

	scrapperinterfaces "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/interfaces"
	models "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/models"
	ports "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

type Scrapper struct {
	repo      scrapperinterfaces.Repository
	logger    ports.Logger
	checkers  []scrapperinterfaces.Checker
	botClient scrapperinterfaces.BotClient
	batchSize int
	workers   int
}

func NewScrapper(logger ports.Logger,
	repo scrapperinterfaces.Repository,
	botClient scrapperinterfaces.BotClient,
	checkers []scrapperinterfaces.Checker,
	batchSize int,
	workers int) *Scrapper {
	if batchSize <= 0 {
		batchSize = 100
	}
	if workers <= 0 {
		workers = 4
	}

	return &Scrapper{
		repo:      repo,
		logger:    logger,
		botClient: botClient,
		batchSize: batchSize,
		workers:   workers,
		checkers:  append([]scrapperinterfaces.Checker(nil), checkers...),
	}
}

func (s *Scrapper) RunCron(ctx context.Context, interval time.Duration) {
	s.logger.Info("scrapper cron scheduled", "interval", interval)
	ticker := time.NewTicker(interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				s.logger.Info("scrapper cron stopped", "error", ctx.Err())
				return
			case <-ticker.C:
				s.logger.Info("scrapper cron tick")
				s.CheckLinks(ctx)
			}

		}
	}()
}

func (s *Scrapper) CheckLinks(ctx context.Context) {
	var (
		afterLinkID        int64
		totalLinks         int
		totalSubscriptions int
	)

	s.logger.Info("scrapper links scan started", "batchSize", s.batchSize, "workerCount", s.workers)

	for {
		batch, err := s.repo.ListAllLinksBatch(ctx, afterLinkID, s.batchSize)
		if err != nil {
			s.logger.Error("failed to load tracked links batch", "afterLinkID", afterLinkID, "batchSize", s.batchSize, "error", err)
			return
		}

		if len(batch) == 0 {
			break
		}

		s.logger.Info("scrapper links batch loaded", "afterLinkID", afterLinkID, "batchSize", s.batchSize, "uniqueLinks", len(batch))

		for _, link := range batch {
			totalLinks++
			totalSubscriptions += len(link.ChatIDs)
		}
		failedLinksByChat := s.processBatch(ctx, batch)
		s.sendFailedLinksReports(ctx, failedLinksByChat)

		afterLinkID = batch[len(batch)-1].LinkID
	}

	s.logger.Info("scrapper links scan finished", "uniqueLinks", totalLinks, "subscriptions", totalSubscriptions)
}

func (s *Scrapper) processBatch(ctx context.Context, batch []models.TrackedLink) map[int64]map[string]struct{} {
	failedLinksByChat := make(map[int64]map[string]struct{})
	if len(batch) == 0 {
		return failedLinksByChat
	}

	workerCount := min(s.workers, len(batch))

	jobs := make(chan models.TrackedLink)
	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)

	for range workerCount {
		wg.Go(func() {
			for link := range jobs {
				if err := s.processLink(ctx, link); err != nil {
					mu.Lock()
					addFailedLink(failedLinksByChat, link)
					mu.Unlock()
				}
			}
		})
	}

	for _, link := range batch {
		jobs <- link
	}
	close(jobs)
	wg.Wait()

	return failedLinksByChat
}

func (s *Scrapper) processLink(ctx context.Context, link models.TrackedLink) error {
	s.logger.Info("checking link update", "linkID", link.LinkID, "chatIDs", link.ChatIDs, "url", link.URL, "lastKnownUpdate", link.LastUpdate)

	update, updateErr := s.getResourceUpdate(ctx, link)
	if updateErr != nil {
		s.logger.Warn("failed to get resource update", "linkID", link.LinkID, "chatIDs", link.ChatIDs, "url", link.URL, "error", updateErr)
		return updateErr
	}

	if !update.HasUpdate {
		s.logger.Info("no link updates detected", "linkID", link.LinkID, "chatIDs", link.ChatIDs, "url", link.URL, "lastKnownUpdate", link.LastUpdate)
		return nil
	}

	s.logger.Info("new link update detected", "linkID", link.LinkID, "chatIDs", link.ChatIDs, "url", link.URL, "lastKnownUpdate", link.LastUpdate, "actualLastUpdate", update.LastUpdate)

	if err := s.botClient.SendUpdates(ctx, link.ChatIDs, link.URL, update.Message); err != nil {
		s.logger.Error("failed to send update notification", "linkID", link.LinkID, "chatIDs", link.ChatIDs, "url", link.URL, "error", err)
		return fmt.Errorf("send update notification: %w", err)
	}
	s.logger.Info("update notification sent", "linkID", link.LinkID, "chatIDs", link.ChatIDs, "url", link.URL)

	if err := s.repo.UpdateLinksLastUpdate(ctx, link.LinkID, update.LastUpdate); err != nil {
		s.logger.Error("failed to update last update time", "linkID", link.LinkID, "chatIDs", link.ChatIDs, "url", link.URL, "error", err)
		return fmt.Errorf("update last update time: %w", err)
	}

	s.logger.Info("link update processed", "linkID", link.LinkID, "chatIDs", link.ChatIDs, "url", link.URL, "previousLastUpdate", link.LastUpdate, "newLastUpdate", update.LastUpdate)
	return nil
}

func addFailedLink(failedLinksByChat map[int64]map[string]struct{}, link models.TrackedLink) {
	for _, chatID := range link.ChatIDs {
		if failedLinksByChat[chatID] == nil {
			failedLinksByChat[chatID] = make(map[string]struct{})
		}
		failedLinksByChat[chatID][link.URL] = struct{}{}
	}
}

func (s *Scrapper) sendFailedLinksReports(ctx context.Context, failedLinksByChat map[int64]map[string]struct{}) {
	for chatID, urlsSet := range failedLinksByChat {
		urls := make([]string, 0, len(urlsSet))
		for url := range urlsSet {
			urls = append(urls, url)
		}
		slices.Sort(urls)

		if err := s.botClient.SendFailedLinksReport(ctx, chatID, urls); err != nil {
			s.logger.Error("failed to send failed links report", "chatID", chatID, "error", err)
			continue
		}
		s.logger.Info("failed links report sent", "chatID", chatID, "count", len(urls))
	}
}

func (s *Scrapper) getResourceUpdate(ctx context.Context, link models.TrackedLink) (ports.ResourceUpdate, error) {
	for _, checker := range s.checkers {
		if !checker.CanHandle(link.URL) {
			continue
		}

		s.logger.Info("resolved link source", "url", link.URL, "since", link.LastUpdate)
		update, updateErr := checker.CheckUpdate(ctx, link.URL, link.LastUpdate)
		if updateErr != nil {
			return ports.ResourceUpdate{}, fmt.Errorf("check resource update: %w", updateErr)
		}
		s.logger.Info("fetched resource update", "url", link.URL, "lastUpdate", update.LastUpdate)
		return update, nil
	}

	return ports.ResourceUpdate{}, ports.ErrLinkNotFound
}
