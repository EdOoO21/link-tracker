package scrapperinterfaces

import (
	"context"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
)

type Repository interface {
	// returns false if url was in repository otherwise true
	TrackLink(ctx context.Context, chatID int64, url string, tags []string) bool
	// returns true if url was in repository otherwise false
	UnTrackLink(ctx context.Context, chatID int64, url string) bool
	ListLinks(ctx context.Context, chatID int64, tags []string) []domain.Link

	ListAllLinks(ctx context.Context) map[int64][]domain.Link
	UpdateLinksLastUpdate(ctx context.Context, chatID int64, url string, lastUpdate time.Time) error

	IsPresent(ctx context.Context, chatID int64) bool
	AddChat(ctx context.Context, chatID int64) bool
	DeleteChat(ctx context.Context, chatID int64) bool
}
