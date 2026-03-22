package main

import (
	"net/http"
	"os"

	app "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot"
	scrapperClient "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/clients/scrapper"
	botinit "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/clients/telegram"
	handlerBot "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/http/bot"
	logs "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/logger"
	settings "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/settings/bot"
)

func main() {
	logger := logs.NewLogger()

	config, err := settings.LoadConfig()
	if err != nil {
		logger.Error("fatal error", "error", err)
		os.Exit(1)
	}

	bot, err := botinit.NewTGBot(logger, config)
	if err != nil {
		logger.Error("fatal error", "error", err)
		os.Exit(1)
	}

	botService := app.NewApp(logger, config, scrapperClient.NewClient(logger, config.ScrapperURL), bot)
	mux := http.NewServeMux()
	h := handlerBot.NewHandler(logger, botService)
	h.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:    ":8090",
		Handler: mux,
	}
	defer func() {
		if err := srv.Close(); err != nil {
			logger.Error("failed to close server", "error", err)
		}
	}()
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server died", "error", err)
		}
	}()

	if err = botService.Run(); err != nil {
		logger.Error("fatal error", "error", err)
		os.Exit(1)
	}
}
