package main

import (
	"net/http"
	"os"
	"time"

	scrapper "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper"
	botClient "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/clients/bot"
	github "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/clients/github"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/clients/stackoverflow"
	handler "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/http/scrapper"
	logs "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/logger"
	repo "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/repository"
	settings "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/settings/scrapper"
)

func main() {
	logger := logs.NewLogger()
	repo := repo.NewRepo()
	cfg, err := settings.LoadConfig()
	if err != nil {
		logger.Error("fatal error", "error", err)
		os.Exit(1)
	}

	githubClient := github.NewGitHubClient()
	stackOverflowClient := stackoverflow.NewStackOverflowClient()
	botClient := botClient.NewClient(logger, cfg.BotURL)

	scrapperService := scrapper.NewScrapper(logger, repo, botClient, githubClient, stackOverflowClient)
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
