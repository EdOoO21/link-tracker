package botinterfaces

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type TGBot interface {
	GetUpdatesChan(config tgbotapi.UpdateConfig) tgbotapi.UpdatesChannel
	Send(c tgbotapi.Chattable) (tgbotapi.Message, error)
}
