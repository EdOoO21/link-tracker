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
	AddTag(ctx context.Context, tag models.AddTag) error
	GetTags(ctx context.Context, chatID int64, url string) ([]string, error)
	DeleteTag(ctx context.Context, tag models.DeleteTag) error

	AddChat(ctx context.Context, chatID int64) error
	DeleteChat(ctx context.Context, chatID int64) error
}
