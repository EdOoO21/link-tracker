package scrapper

import (
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
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			s.CheckLinks()
		}
	}()
}

func (s *Scrapper) CheckLinks() {
	for chatID, links := range s.repo.ListAllLinks() {
		for _, link := range links {
			update, description, err := s.getResourceUpdate(link.URL)
			if err != nil {
				s.logger.Warn("failed to get resource update", "chatID", chatID, "url", link.URL, "error", err)
				continue
			}

			if !update.LastUpdate.After(link.LastUpdate) {
				continue
			}

			if err = s.botClient.SendUpdates([]int64{chatID}, link.URL, description); err != nil {
				s.logger.Error("failed to send update notification", "chatID", chatID, "url", link.URL, "error", err)
				continue
			}

			if err = s.repo.UpdateLinksLastUpdate(chatID, link.URL, update.LastUpdate); err != nil {
				s.logger.Error("failed to update last update time", "chatID", chatID, "url", link.URL, "error", err)
				continue
			}

			s.logger.Info("link update processed", "chatID", chatID, "url", link.URL, "lastUpdate", update.LastUpdate)
		}
	}
}

func (s *Scrapper) getResourceUpdate(rawURL string) (ports.ResourceUpdate, string, error) {
	if owner, repo, err := s.sources.Github.ParseGitHubURL(rawURL); err == nil {
		update, err := s.sources.Github.GetRepoUpdate(owner, repo)
		if err != nil {
			return ports.ResourceUpdate{}, "", err
		}
		return update, "Обнаружено обновление GitHub репозитория.", nil
	}

	if questionID, err := s.sources.StackOverflow.ParseStackOverflowURL(rawURL); err == nil {
		update, err := s.sources.StackOverflow.GetQuestionUpdate(questionID)
		if err != nil {
			return ports.ResourceUpdate{}, "", err
		}
		return update, "Обнаружено обновление вопроса StackOverflow.", nil
	}

	return ports.ResourceUpdate{}, "", ports.ErrLinkNotFound
}

func (s *Scrapper) AddLink(link models.AddLink) error {
	if !s.repo.IsPresent(link.ChatID) {
		return ports.ErrChatNotFound
	}
	if ok := s.repo.TrackLink(link.ChatID, link.URL, link.Tags); !ok {
		return ports.ErrLinkAlreadyExists
	}
	return nil
}

func (s *Scrapper) DeleteLink(link models.DeleteLink) error {
	if !s.repo.IsPresent(link.ChatID) {
		return ports.ErrChatNotFound
	}

	if ok := s.repo.UnTrackLink(link.ChatID, link.URL); !ok {
		return ports.ErrLinkNotFound
	}
	return nil
}

func (s *Scrapper) GetLinks(chatID int64, tags []string) ([]domain.Link, error) {
	if !s.repo.IsPresent(chatID) {
		return nil, ports.ErrChatNotFound
	}

	return s.repo.ListLinks(chatID, tags), nil
}

func (s *Scrapper) AddChat(chatID int64) error {
	if ok := s.repo.AddChat(chatID); !ok {
		return ports.ErrChatAlreadyExists
	}
	return nil
}

func (s *Scrapper) DeleteChat(chatID int64) error {
	if ok := s.repo.DeleteChat(chatID); !ok {
		return ports.ErrChatNotFound
	}
	return nil
}
