package sqlrepo

import (
	"context"
	"fmt"
	"time"

	models "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/models"
)

func (r *Repository) ListAllLinks(ctx context.Context) ([]models.TrackedLink, error) {
	const query = `
		SELECT l.id, l.url, l.last_update, cl.chat_id
		FROM links l
		INNER JOIN chat_links cl ON cl.link_id = l.id
		ORDER BY l.id, cl.chat_id
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query tracked links: %w", err)
	}
	defer rows.Close()

	res := make([]models.TrackedLink, 0)
	for rows.Next() {
		var (
			linkID     int64
			url        string
			lastUpdate time.Time
			chatID     int64
		)

		scanErr := rows.Scan(&linkID, &url, &lastUpdate, &chatID)
		if scanErr != nil {
			return nil, fmt.Errorf("scan tracked link row: %w", scanErr)
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
		return nil, fmt.Errorf("iterate tracked links: %w", rowsErr)
	}

	return res, nil
}
