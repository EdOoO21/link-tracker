package scrapperinterfaces

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/models"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
)

type ScrapperService interface {
	AddLink(ctx context.Context, link models.AddLink) error
	GetLinks(ctx context.Context, chatID int64, tags []string) ([]domain.Link, error)
	DeleteLink(ctx context.Context, link models.DeleteLink) error

	AddChat(ctx context.Context, chatID int64) error
	DeleteChat(ctx context.Context, chatID int64) error
}
