package botinit

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	logger "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
	settings "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/settings/bot"
)

type BotInit struct {
	bot    *tgbotapi.BotAPI
	logger logger.Logger
}

func NewTGBot(logger logger.Logger, cfg *settings.Config) (*BotInit, error) {
	logger.Info("bot initialization started")
	bot, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		logger.Error("bot initialization failed", "error", err)
		return nil, err
	}
	logger.Error("bot successfully initialized")

	logger.Error("commands list initialization started")
	_, err = bot.Request(tgbotapi.NewSetMyCommands(cfg.Commands...))
	if err != nil {
		logger.Error("commands list initialization failed", "error", err)
		return nil, err
	}
	logger.Error("commands list successfully initialized")

	return &BotInit{bot: bot, logger: logger}, nil
}

func (b *BotInit) GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	return b.bot.GetUpdatesChan(config)
}

func (b *BotInit) Send(c tgbotapi.Chattable) (tgbotapi.Message, error) {
	return b.bot.Send(c)
}
