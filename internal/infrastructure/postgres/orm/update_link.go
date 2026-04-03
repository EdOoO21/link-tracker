package ormrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jackc/pgx/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

func (r *Repository) UpdateLinksLastUpdate(ctx context.Context, linkID int64, lastUpdate time.Time) error {
	ds := dialect.
		Update("links").
		Prepared(true).
		Set(goqu.Record{"last_update": lastUpdate}).
		Where(goqu.C("id").Eq(linkID)).
		Returning("id")

	sql, args, err := buildSQL(ds)
	if err != nil {
		return fmt.Errorf("build update link query: %w", err)
	}

	var updatedLinkID int64
	updateErr := r.pool.QueryRow(ctx, sql, args...).Scan(&updatedLinkID)
	if updateErr != nil {
		if !errors.Is(updateErr, pgx.ErrNoRows) {
			return fmt.Errorf("update link last update: %w", updateErr)
		}
		return ports.ErrLinkNotFound
	}

	return nil
}
