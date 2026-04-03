package postgresrepo

import (
	"context"
	"fmt"

	scrapperinterfaces "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/interfaces"
	ormrepo "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/postgres/orm"
	sqlrepo "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/postgres/sql"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
	settings "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/settings/scrapper"
)

type Repository interface {
	scrapperinterfaces.Repository
	Close()
}

func NewRepository(ctx context.Context, logger ports.Logger, cfg settings.DBConfig) (Repository, error) {
	switch cfg.AccessType {
	case "sql":
		repository, err := sqlrepo.NewRepository(ctx, logger, cfg)
		if err != nil {
			return nil, fmt.Errorf("create sql postgres repository: %w", err)
		}
		return repository, nil
	case "orm":
		repository, err := ormrepo.NewRepository(ctx, logger, cfg)
		if err != nil {
			return nil, fmt.Errorf("create orm postgres repository: %w", err)
		}
		return repository, nil
	default:
		return nil, fmt.Errorf("unsupported postgres access type %q", cfg.AccessType)
	}
}
