package scrapperinterfaces

import (
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/models"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
)

type ScrapperService interface {
	AddLink(link models.AddLink) error
	GetLinks(chatID int64, tags []string) ([]domain.Link, error)
	DeleteLink(link models.DeleteLink) error

	AddChat(chatID int64) error
	DeleteChat(chatID int64) error
}
