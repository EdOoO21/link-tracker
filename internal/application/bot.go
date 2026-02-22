package application

import (
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/settings"
)

type App struct {
}

func NewApp() *App {
	return &App{}
}

func (a *App) Run(cfg *settings.Config) error {
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return err
	}

	_, err = bot.Request(tgbotapi.NewSetMyCommands(cfg.Commands...))
	if err != nil {
		return err
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		chatID := update.Message.Chat.ID
		var msg tgbotapi.MessageConfig
		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "help":
				msg = tgbotapi.NewMessage(chatID, "/start - начало работы пользователя\n/help - вывод списка доступных команд")
			case "start":
				msg = tgbotapi.NewMessage(chatID, "Добро пожаловать! Используйте /help, чтобы посмотреть доступные команды.")
			default:
				msg = tgbotapi.NewMessage(chatID, "Неизвестная команда. Воспользуйтесь /help, чтобы посмотреть список доступных команд.")
			}
			_, err = bot.Send(msg)
			if err != nil {
				slog.Error("error to send message", "error", err)
			}
		}
	}
	return nil
}
