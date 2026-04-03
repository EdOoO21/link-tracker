package sqlrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

func (r *Repository) ListLinks(ctx context.Context, chatID int64, tags []string) ([]domain.Link, error) {
	present, err := isPresent(ctx, r.pool, chatID)
	if err != nil {
		return nil, fmt.Errorf("check chat existence before list links: %w", err)
	}
	if !present {
		return nil, ports.ErrChatNotFound
	}

	filterTags := uniqueStrings(tags)

	rows, err := r.queryLinks(ctx, chatID, filterTags)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	links := make([]domain.Link, 0)
	for rows.Next() {
		var (
			linkID     int64
			url        string
			lastUpdate time.Time
			tag        pgtype.Text
		)

		scanErr := rows.Scan(&linkID, &url, &lastUpdate, &tag)
		if scanErr != nil {
			return nil, fmt.Errorf("scan link row: %w", scanErr)
		}

		if len(links) == 0 || links[len(links)-1].URL != url {
			links = append(links, domain.Link{
				URL:        url,
				Tags:       make(map[string]struct{}),
				LastUpdate: lastUpdate,
			})
		}

		if tag.Valid {
			links[len(links)-1].Tags[tag.String] = struct{}{}
		}
	}

	rowsErr := rows.Err()
	if rowsErr != nil {
		return nil, fmt.Errorf("iterate link rows: %w", rowsErr)
	}

	return links, nil
}

func (r *Repository) queryLinks(ctx context.Context, chatID int64, tags []string) (pgx.Rows, error) {
	if len(tags) == 0 {
		const query = `
			SELECT l.id, l.url, l.last_update, t.name
			FROM chat_links cl
			INNER JOIN links l ON l.id = cl.link_id
			LEFT JOIN link_tags lt ON lt.chat_id = cl.chat_id AND lt.link_id = cl.link_id
			LEFT JOIN tags t ON t.id = lt.tag_id
			WHERE cl.chat_id = $1
			ORDER BY l.id, t.name
		`

		rows, err := r.pool.Query(ctx, query, chatID)
		if err != nil {
			return nil, fmt.Errorf("query links without tag filter: %w", err)
		}
		return rows, nil
	}

	const query = `
		WITH filtered_links AS (
			SELECT cl.link_id
			FROM chat_links cl
			INNER JOIN link_tags lt ON lt.chat_id = cl.chat_id AND lt.link_id = cl.link_id
			INNER JOIN tags filter_tags ON filter_tags.id = lt.tag_id
			WHERE cl.chat_id = $1
			  AND filter_tags.name = ANY($2)
			GROUP BY cl.link_id
			HAVING COUNT(DISTINCT filter_tags.name) = $3
		)
		SELECT l.id, l.url, l.last_update, t.name
		FROM filtered_links fl
		INNER JOIN links l ON l.id = fl.link_id
		LEFT JOIN link_tags lt ON lt.chat_id = $1 AND lt.link_id = fl.link_id
		LEFT JOIN tags t ON t.id = lt.tag_id
		ORDER BY l.id, t.name
	`

	rows, err := r.pool.Query(ctx, query, chatID, tags, len(tags))
	if err != nil {
		return nil, fmt.Errorf("query links with tag filter: %w", err)
	}

	return rows, nil
}
