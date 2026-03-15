package main

import (
	"net/http"
	"time"

	scrapper "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper"
	handler "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/http/scrapper"
	logs "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/logger"
	repo "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/repository"
)

func main() {
	logger := logs.NewLogger()
	repo := repo.NewRepo()

	scrapperService := scrapper.NewScrapper(logger, repo)
	scrapperService.RunCron(1 * time.Minute)

	mux := http.NewServeMux()
	h := handler.NewHandler(scrapperService, logger)
	h.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	defer func() {
		if err := srv.Close(); err != nil {
			logger.Error("failed to close server", "error", err)
		}
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server died", "error", err)
	}
}
