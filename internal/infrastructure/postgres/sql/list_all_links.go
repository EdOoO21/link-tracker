package sqlrepo

import (
	"context"
	"fmt"
	"time"

	models "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/models"
)

func (r *Repository) ListAllLinksBatch(ctx context.Context, afterLinkID int64, limit int) ([]models.TrackedLink, error) {
	if limit <= 0 {
		return []models.TrackedLink{}, nil
	}

	const query = `
		WITH batch_links AS (
			SELECT l.id, l.url, l.last_update
			FROM links l
			WHERE l.id > $1
			ORDER BY l.id
			LIMIT $2
		)
		SELECT bl.id, bl.url, bl.last_update, cl.chat_id
		FROM batch_links bl
		INNER JOIN chat_links cl ON cl.link_id = bl.id
		ORDER BY bl.id, cl.chat_id
	`

	rows, err := r.pool.Query(ctx, query, afterLinkID, limit)
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
