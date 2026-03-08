package application

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	inf "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/interfaces"
	settings "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/settings"
)

const (
	Unknown        = "unknown"
	HelpCommand    = "/start - начало работы пользователя\n/help - вывод списка доступных команд"
	StartCommand   = "Добро пожаловать! Используйте /help, чтобы посмотреть доступные команды."
	UnknownCommand = "Неизвестная команда. Воспользуйтесь /help, чтобы посмотреть список доступных команд."
)

type App struct {
	logger inf.Logger
	config *settings.Config
}

func NewApp(logger inf.Logger, config *settings.Config) *App {
	return &App{
		logger: logger,
		config: config,
	}
}

func (a *App) Run(bot inf.TGBot) error {

	u := tgbotapi.NewUpdate(a.config.Offset)
	u.Timeout = a.config.Timeout
	updates := bot.GetUpdatesChan(u)

	a.logger.Info("got channel for updates")

	for update := range updates {
		if update.Message == nil {
			continue
		}
		chatID := update.Message.Chat.ID
		a.logger.Info("recieved update", "chatID", chatID) // так как у нас лог не публичный, то можем прокинуть в логи
		var msg tgbotapi.MessageConfig
		if update.Message.IsCommand() {
			command := update.Message.Command()
			switch command {
			case "help":
				msg = tgbotapi.NewMessage(chatID, HelpCommand)
			case "start":
				msg = tgbotapi.NewMessage(chatID, StartCommand)
			default:
				msg = tgbotapi.NewMessage(chatID, UnknownCommand)
				command = Unknown // чтобы не засорять логгер при огромных текстах
			}
			a.logger.Info("recieved command", "chatID", chatID, "command", command)
			_, err := bot.Send(msg)
			if err != nil {
				a.logger.Error("error to send message", "error", err, "chatID", chatID, "command", command)
			}
			a.logger.Info("reply sent", "chatID", chatID, "command", command)
		} else {
			a.logger.Info("recieved text", "chatID", chatID)
		}
	}
	return nil
}
