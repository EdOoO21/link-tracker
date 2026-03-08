package scrapper

import (
	"net/http"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application"
	inf "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/interfaces"
)

type Scrapper struct {
	repo   inf.Repository
	logger application.Logger
}

func NewScrapper(logger application.Logger, repo inf.Repository) *Scrapper {
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
	for _, links := range s.repo.ListAllLinks() {
		for _, link := range links {

			resp, err := http.Get(link.URL)
			if err != nil {
				s.logger.Error("failed to check link", "url", link.URL, "error", err)
				continue
			}
			defer resp.Body.Close()

			// дописать
		}
	}
}
