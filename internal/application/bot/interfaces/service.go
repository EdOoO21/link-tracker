package botinterfaces

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/models"
)

type BotService interface {
	SendUpdateMessages(ctx context.Context, updates models.SendUpdates) error
	SendFailedLinksReport(ctx context.Context, report models.FailedLinksReport) error
}
