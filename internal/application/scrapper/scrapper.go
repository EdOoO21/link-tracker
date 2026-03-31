package scrapper

import (
	"fmt"
	"time"

	scrapperinterfaces "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/interfaces"
	models "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/models"
	domain "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	ports "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

type Sources struct {
	Github        scrapperinterfaces.GithubUpdates
	StackOverflow scrapperinterfaces.StackOverflowUpdates
}

type Scrapper struct {
	repo      scrapperinterfaces.Repository
	logger    ports.Logger
	sources   Sources
	botClient scrapperinterfaces.BotClient
}

func NewScrapper(logger ports.Logger,
	repo scrapperinterfaces.Repository,
	botClient scrapperinterfaces.BotClient,
	github scrapperinterfaces.GithubUpdates,
	stackOverflow scrapperinterfaces.StackOverflowUpdates) *Scrapper {
	return &Scrapper{
		repo:      repo,
		logger:    logger,
		botClient: botClient,
		sources: Sources{
			Github:        github,
			StackOverflow: stackOverflow,
		},
	}
}

func (s *Scrapper) RunCron(interval time.Duration) {
	s.logger.Info("scrapper cron scheduled", "interval", interval)
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			s.logger.Info("scrapper cron tick")
			s.CheckLinks()
		}
	}()
}

func (s *Scrapper) CheckLinks() {
	allLinks := s.repo.ListAllLinks()
	totalLinks := 0
	for _, links := range allLinks {
		totalLinks += len(links)
	}

	s.logger.Info("scrapper links scan started", "chats", len(allLinks), "links", totalLinks)

	for chatID, links := range allLinks {
		for _, link := range links {
			s.logger.Info("checking link update", "chatID", chatID, "url", link.URL, "lastKnownUpdate", link.LastUpdate)

			update, description, err := s.getResourceUpdate(link.URL)
			if err != nil {
				s.logger.Warn("failed to get resource update", "chatID", chatID, "url", link.URL, "error", err)
				continue
			}

			if !update.LastUpdate.After(link.LastUpdate) {
				s.logger.Info("no link updates detected", "chatID", chatID, "url", link.URL, "lastKnownUpdate", link.LastUpdate, "actualLastUpdate", update.LastUpdate)
				continue
			}

			s.logger.Info("new link update detected", "chatID", chatID, "url", link.URL, "lastKnownUpdate", link.LastUpdate, "actualLastUpdate", update.LastUpdate)

			if err = s.botClient.SendUpdates([]int64{chatID}, link.URL, description); err != nil {
				s.logger.Error("failed to send update notification", "chatID", chatID, "url", link.URL, "error", err)
				continue
			}
			s.logger.Info("update notification sent", "chatID", chatID, "url", link.URL)

			if err = s.repo.UpdateLinksLastUpdate(chatID, link.URL, update.LastUpdate); err != nil {
				s.logger.Error("failed to update last update time", "chatID", chatID, "url", link.URL, "error", err)
				continue
			}

			s.logger.Info("link update processed", "chatID", chatID, "url", link.URL, "previousLastUpdate", link.LastUpdate, "newLastUpdate", update.LastUpdate)
		}
	}

	s.logger.Info("scrapper links scan finished", "chats", len(allLinks), "links", totalLinks)
}

func (s *Scrapper) AddLink(link models.AddLink) error {
	s.logger.Info("add link requested", "chatID", link.ChatID, "url", link.URL, "tags", link.Tags)
	if !s.repo.IsPresent(link.ChatID) {
		s.logger.Warn("add link failed: chat not found", "chatID", link.ChatID, "url", link.URL)
		return ports.ErrChatNotFound
	}
	if ok := s.repo.TrackLink(link.ChatID, link.URL, link.Tags); !ok {
		s.logger.Warn("add link failed: link already exists", "chatID", link.ChatID, "url", link.URL)
		return ports.ErrLinkAlreadyExists
	}
	s.logger.Info("link added", "chatID", link.ChatID, "url", link.URL, "tags", link.Tags)
	return nil
}

func (s *Scrapper) DeleteLink(link models.DeleteLink) error {
	s.logger.Info("delete link requested", "chatID", link.ChatID, "url", link.URL)
	if !s.repo.IsPresent(link.ChatID) {
		s.logger.Warn("delete link failed: chat not found", "chatID", link.ChatID, "url", link.URL)
		return ports.ErrChatNotFound
	}

	if ok := s.repo.UnTrackLink(link.ChatID, link.URL); !ok {
		s.logger.Warn("delete link failed: link not found", "chatID", link.ChatID, "url", link.URL)
		return ports.ErrLinkNotFound
	}
	s.logger.Info("link deleted", "chatID", link.ChatID, "url", link.URL)
	return nil
}

func (s *Scrapper) GetLinks(chatID int64, tags []string) ([]domain.Link, error) {
	s.logger.Info("get links requested", "chatID", chatID, "tags", tags)
	if !s.repo.IsPresent(chatID) {
		s.logger.Warn("get links failed: chat not found", "chatID", chatID, "tags", tags)
		return nil, ports.ErrChatNotFound
	}

	links := s.repo.ListLinks(chatID, tags)
	s.logger.Info("links loaded", "chatID", chatID, "tags", tags, "count", len(links))
	return links, nil
}

func (s *Scrapper) AddChat(chatID int64) error {
	s.logger.Info("add chat requested", "chatID", chatID)
	if ok := s.repo.AddChat(chatID); !ok {
		s.logger.Warn("add chat failed: chat already exists", "chatID", chatID)
		return ports.ErrChatAlreadyExists
	}
	s.logger.Info("chat added", "chatID", chatID)
	return nil
}

func (s *Scrapper) DeleteChat(chatID int64) error {
	s.logger.Info("delete chat requested", "chatID", chatID)
	if ok := s.repo.DeleteChat(chatID); !ok {
		s.logger.Warn("delete chat failed: chat not found", "chatID", chatID)
		return ports.ErrChatNotFound
	}
	s.logger.Info("chat deleted", "chatID", chatID)
	return nil
}

func (s *Scrapper) getResourceUpdate(rawURL string) (ports.ResourceUpdate, string, error) {
	if owner, repo, parseErr := s.sources.Github.ParseGitHubURL(rawURL); parseErr == nil {
		s.logger.Info("resolved link source", "url", rawURL, "source", "github", "owner", owner, "repo", repo)
		update, updateErr := s.sources.Github.GetRepoUpdate(owner, repo)
		if updateErr != nil {
			return ports.ResourceUpdate{}, "", fmt.Errorf("get github update: %w", updateErr)
		}
		s.logger.Info("fetched github resource update", "url", rawURL, "lastUpdate", update.LastUpdate)
		return update, "Обнаружено обновление GitHub репозитория.", nil
	}

	if questionID, parseErr := s.sources.StackOverflow.ParseStackOverflowURL(rawURL); parseErr == nil {
		s.logger.Info("resolved link source", "url", rawURL, "source", "stackoverflow", "questionID", questionID)
		update, updateErr := s.sources.StackOverflow.GetQuestionUpdate(questionID)
		if updateErr != nil {
			return ports.ResourceUpdate{}, "", fmt.Errorf("get stackoverflow update: %w", updateErr)
		}
		s.logger.Info("fetched stackoverflow resource update", "url", rawURL, "lastUpdate", update.LastUpdate)
		return update, "Обнаружено обновление вопроса StackOverflow.", nil
	}

	return ports.ResourceUpdate{}, "", ports.ErrLinkNotFound
}
