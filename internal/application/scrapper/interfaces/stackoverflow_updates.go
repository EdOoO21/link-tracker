package scrapperinterfaces

import (
	"context"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

type StackOverflowUpdates interface {
	GetQuestionUpdate(ctx context.Context, questionID string, since time.Time) (ports.ResourceUpdate, error)

	ParseStackOverflowURL(url string) (string, error)
}
