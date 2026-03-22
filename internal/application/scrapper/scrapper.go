package scrapper

import (
	"time"

	repo "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/interfaces"
	models "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/models"
	domain "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	ports "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

type Scrapper struct {
	repo   repo.Repository
	logger ports.Logger
}

func NewScrapper(logger ports.Logger, repo repo.Repository) *Scrapper {
	return &Scrapper{
		repo:   repo,
		logger: logger,
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
	panic("not implemented")
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
