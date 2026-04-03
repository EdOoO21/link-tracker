package scrapperinterfaces

import (
	"context"
	"time"

	models "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/models"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
)

type Repository interface {
	TrackLink(ctx context.Context, chatID int64, url string, tags []string) error
	UnTrackLink(ctx context.Context, chatID int64, url string) error
	ListLinks(ctx context.Context, chatID int64, tags []string) ([]domain.Link, error)

	ListAllLinks(ctx context.Context) ([]models.TrackedLink, error)
	UpdateLinksLastUpdate(ctx context.Context, linkID int64, lastUpdate time.Time) error

	IsPresent(ctx context.Context, chatID int64) (bool, error)
	AddChat(ctx context.Context, chatID int64) error
	DeleteChat(ctx context.Context, chatID int64) error
}
