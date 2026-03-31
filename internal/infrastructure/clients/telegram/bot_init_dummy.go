package botinit

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	logger "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

type DummyBot struct {
	logger  logger.Logger
	updates tgbotapi.UpdatesChannel
}

func NewDummyBot(logger logger.Logger) *DummyBot {
	return &DummyBot{
		logger:  logger,
		updates: make(chan tgbotapi.Update),
	}
}

func (b *DummyBot) GetUpdatesChan(_ tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel {
	return b.updates
}

func (b *DummyBot) Send(_ tgbotapi.Chattable) (tgbotapi.Message, error) {
	b.logger.Info("dummy telegram send")
	return tgbotapi.Message{}, nil
}
