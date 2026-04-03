package scrapperinterfaces

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

type StackOverflowUpdates interface {
	GetQuestionUpdate(ctx context.Context, questionID string) (ports.ResourceUpdate, error)

	ParseStackOverflowURL(url string) (string, error)
}
