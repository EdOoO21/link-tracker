package scrapper

import (
	"errors"
	"time"

	repo "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/interfaces"
	models "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/models"
	domain "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	logger "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

var (
	ChatNotFound      = errors.New("chat not found")
	ChatAlreadyExists = errors.New("chat already exists")
	LinkAlreadyExists = errors.New("link already exists")
	LinkNotFound      = errors.New("link not found")
)

type Scrapper struct {
	repo   repo.Repository
	logger logger.Logger
}

func NewScrapper(logger logger.Logger, repo repo.Repository) *Scrapper {
	return &Scrapper{
		repo:   repo,
		logger: logger,
	}
}

func (s *Scrapper) RunCron(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			s.checkLinks()
		}
	}()
}

func (s *Scrapper) checkLinks() {
	// for _, links := range s.repo.ListAllLinks() {
	// 	for _, link := range links {

	// 		resp, err := http.Get(link.URL)
	// 		if err != nil {
	// 			s.logger.Error("failed to check link", "url", link.URL, "error", err)
	// 			continue
	// 		}
	// 		defer resp.Body.Close()

	// 		// дописать
	// 	}
	// }
	panic("not implemented")
}

func (s *Scrapper) AddLink(link models.AddLink) error {
	if !s.repo.IsPresent(link.ChatID) {
		return ChatNotFound
	}
	if ok := s.repo.TrackLink(link.ChatID, link.URL, link.Tags); !ok {
		return LinkAlreadyExists
	}
	return nil
}

func (s *Scrapper) DeleteLink(link models.DeleteLink) error {
	if !s.repo.IsPresent(link.ChatID) {
		return ChatNotFound
	}

	if ok := s.repo.UnTrackLink(link.ChatID, link.URL); !ok {
		return LinkNotFound
	}
	return nil
}

func (s *Scrapper) GetLinks(chatID int64, tags []string) ([]domain.Link, error) {
	if !s.repo.IsPresent(chatID) {
		return nil, ChatNotFound
	}

	return s.repo.ListLinks(chatID, tags), nil
}

func (s *Scrapper) AddChat(chatID int64) error {
	if ok := s.repo.AddChat(chatID); !ok {
		return ChatAlreadyExists
	}
	return nil
}

func (s *Scrapper) DeleteChat(chatID int64) error {
	if ok := s.repo.DeleteChat(chatID); !ok {
		return ChatNotFound
	}
	return nil
}
