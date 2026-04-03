package botinterfaces

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
)

type Repository interface {
	// returns ports.ErrLinkAlreadyExists if link exits, otherwise nil or other error
	TrackLink(ctx context.Context, chatID int64, url string, tags []string) error
	// returns ports.ErrLinkNotFound if link exits, otherwise nil or other error
	UnTrackLink(ctx context.Context, chatID int64, url string) error

	ListLinks(ctx context.Context, chatID int64, tags []string) ([]domain.Link, error)

	// returns ports.ErrChatAlreadyExists if link exits, otherwise nil or other error
	AddChat(ctx context.Context, chatID int64) error
	// returns ports.ErrChatNotFound if link exits, otherwise nil or other error
	DeleteChat(ctx context.Context, chatID int64) error

	// returns true if link exits, otherwise false
	IsLinkPresent(ctx context.Context, chatID int64, link string) bool
}
