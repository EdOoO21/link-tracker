package scrapperinterfaces

import (
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
)

type Repository interface {
	// returns false if url was in repository otherwise true
	TrackLink(chatID int64, url string, tags []string) bool
	// returns true if url was in repository otherwise false
	UnTrackLink(chatID int64, url string) bool
	ListLinks(chatID int64, tags []string) []domain.Link

	ListAllLinks() map[int64][]domain.Link
	UpdateLinksLastUpdate(chatID int64, url string, lastUpdate time.Time) error

	IsPresent(chatID int64) bool
	AddChat(chatID int64) bool
	DeleteChat(chatID int64) bool
}
