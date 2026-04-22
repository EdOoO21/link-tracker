package scrapperinterfaces

import (
	"context"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

type Checker interface {
	CanHandle(url string) bool
	CheckUpdate(ctx context.Context, url string, since time.Time) (ports.ResourceUpdate, error)
}
