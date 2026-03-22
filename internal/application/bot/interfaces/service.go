package botinterfaces

import "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/models"

type BotService interface {
	SendUpdateMessages(updates models.SendUpdates) error
}
