package sqlrepo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
	settings "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/settings/scrapper"
)

const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
)

type Repository struct {
	pool   *pgxpool.Pool
	logger ports.Logger
}

func NewRepository(ctx context.Context, logger ports.Logger, cfg settings.DBConfig) (*Repository, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	poolConfig.ConnConfig.User = cfg.DatabaseUser
	poolConfig.ConnConfig.Password = cfg.DatabasePassword

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Repository{pool: pool, logger: logger}, nil
}

func (r *Repository) Close() {
	r.pool.Close()
}
