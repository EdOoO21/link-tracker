package sqlrepo

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

func (r *Repository) UpdateLinksLastUpdate(ctx context.Context, linkID int64, lastUpdate time.Time) error {
	const query = `
		UPDATE links
		SET last_update = $2
		WHERE id = $1
		RETURNING id
	`

	var updatedLinkID int64
	updateErr := r.pool.QueryRow(ctx, query, linkID, lastUpdate).Scan(&updatedLinkID)
	if updateErr != nil {
		if !errors.Is(updateErr, pgx.ErrNoRows) {
			return fmt.Errorf("update link last update: %w", updateErr)
		}
		return ports.ErrLinkNotFound
	}

	return nil
}
