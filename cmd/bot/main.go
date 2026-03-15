package main

import (
	"os"

	app "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot"
	botinit "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/clients/telegram"
	logs "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/logger"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/repository"
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

	a := app.NewApp(logger, config, repository.NewRepo())
	if err := a.Run(bot); err != nil {
		logger.Error("fatal error", "error", err)
		os.Exit(1)
	}
}
