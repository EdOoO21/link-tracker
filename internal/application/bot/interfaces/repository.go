package botinterfaces

import "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"

type Repository interface {
	// returns ports.ErrLinkAlreadyExists if link exits, otherwise nil or other error
	TrackLink(chatID int64, url string, tags []string) error
	// returns ports.ErrLinkNotFound if link exits, otherwise nil or other error
	UnTrackLink(chatID int64, url string) error

	ListLinks(chatID int64, tags []string) ([]domain.Link, error)

	// returns ports.ErrChatAlreadyExists if link exits, otherwise nil or other error
	AddChat(chatID int64) error
	// returns ports.ErrChatNotFound if link exits, otherwise nil or other error
	DeleteChat(chatID int64) error

	// returns true if link exits, otherwise false
	IsLinkPresent(chatID int64, link string) bool
}
