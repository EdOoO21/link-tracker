package ormrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	models "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/models"
)

func (r *Repository) ListAllLinksBatch(ctx context.Context, afterLinkID int64, limit int) ([]models.TrackedLink, error) {
	if limit <= 0 {
		return []models.TrackedLink{}, nil
	}

	batchLinks := dialect.
		From(goqu.T("links").As("l")).
		Prepared(true).
		Select(
			goqu.I("l.id"),
			goqu.I("l.url"),
			goqu.I("l.last_update"),
		).
		Where(goqu.I("l.id").Gt(afterLinkID)).
		Order(goqu.I("l.id").Asc()).
		Limit(uint(limit))

	ds := dialect.
		From(goqu.T("batch_links").As("bl")).
		With("batch_links", batchLinks).
		Prepared(true).
		Select(
			goqu.I("bl.id"),
			goqu.I("bl.url"),
			goqu.I("bl.last_update"),
			goqu.I("cl.chat_id"),
		).
		Join(
			goqu.T("chat_links").As("cl"),
			goqu.On(goqu.I("cl.link_id").Eq(goqu.I("bl.id"))),
		).
		Order(goqu.I("bl.id").Asc(), goqu.I("cl.chat_id").Asc())

	sql, args, err := buildSQL(ds)
	if err != nil {
		return nil, fmt.Errorf("build tracked links batch query: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query tracked links batch: %w", err)
	}
	defer rows.Close()

	res := make([]models.TrackedLink, 0, limit)
	for rows.Next() {
		var (
			linkID     int64
			url        string
			lastUpdate time.Time
			chatID     int64
		)

		scanErr := rows.Scan(&linkID, &url, &lastUpdate, &chatID)
		if scanErr != nil {
			return nil, fmt.Errorf("scan tracked link batch row: %w", scanErr)
		}

		if len(res) == 0 || res[len(res)-1].LinkID != linkID {
			res = append(res, models.TrackedLink{
				LinkID:     linkID,
				URL:        url,
				LastUpdate: lastUpdate,
				ChatIDs:    []int64{chatID},
			})
			continue
		}

		res[len(res)-1].ChatIDs = append(res[len(res)-1].ChatIDs, chatID)
	}

	rowsErr := rows.Err()
	if rowsErr != nil {
		return nil, fmt.Errorf("iterate tracked links batch: %w", rowsErr)
	}

	return res, nil
}
