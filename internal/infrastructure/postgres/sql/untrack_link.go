package sqlrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

func (r *Repository) UnTrackLink(ctx context.Context, chatID int64, url string) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("begin untrack link transaction: %w", err)
	}
	defer func() {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			r.logger.Error("failed to rollback untrack link transaction", "error", rollbackErr, "chatID", chatID, "url", url)
		}
	}()

	present, err := isPresent(ctx, tx, chatID)
	if err != nil {
		return fmt.Errorf("check chat presence: %w", err)
	}
	if !present {
		return ports.ErrChatNotFound
	}

	linkID, err := r.getLinkIDByChatAndURL(ctx, tx, chatID, url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ports.ErrLinkNotFound
		}
		return fmt.Errorf("get subscription link: %w", err)
	}

	tagIDs, err := r.listTagIDsByChatAndLink(ctx, tx, chatID, linkID)
	if err != nil {
		return fmt.Errorf("list subscription tags: %w", err)
	}

	deleteErr := r.deleteSubscription(ctx, tx, chatID, linkID)
	if deleteErr != nil {
		return deleteErr
	}

	linkCleanupErr := r.deleteUnusedLink(ctx, tx, linkID)
	if linkCleanupErr != nil {
		return fmt.Errorf("delete unused link: %w", linkCleanupErr)
	}

	for _, tagID := range tagIDs {
		tagCleanupErr := r.deleteUnusedTag(ctx, tx, tagID)
		if tagCleanupErr != nil {
			return fmt.Errorf("delete unused tag %d: %w", tagID, tagCleanupErr)
		}
	}

	commitErr := tx.Commit(ctx)
	if commitErr != nil {
		return fmt.Errorf("commit untrack link transaction: %w", commitErr)
	}

	return nil
}

func (r *Repository) getLinkIDByChatAndURL(ctx context.Context, tx pgx.Tx, chatID int64, url string) (int64, error) {
	const query = `
		SELECT l.id
		FROM chat_links cl
		JOIN links l ON l.id = cl.link_id
		WHERE cl.chat_id = $1 AND l.url = $2
	`

	var linkID int64
	scanErr := tx.QueryRow(ctx, query, chatID, url).Scan(&linkID)
	if scanErr != nil {
		return 0, fmt.Errorf("scan subscription link id: %w", scanErr)
	}

	return linkID, nil
}

func (r *Repository) listTagIDsByChatAndLink(ctx context.Context, tx pgx.Tx, chatID, linkID int64) ([]int64, error) {
	const query = `
		SELECT tag_id
		FROM link_tags
		WHERE chat_id = $1 AND link_id = $2
	`

	rows, err := tx.Query(ctx, query, chatID, linkID)
	if err != nil {
		return nil, fmt.Errorf("query subscription tag ids: %w", err)
	}
	defer rows.Close()

	tagIDs := make([]int64, 0)
	for rows.Next() {
		var tagID int64
		scanErr := rows.Scan(&tagID)
		if scanErr != nil {
			return nil, fmt.Errorf("scan subscription tag id: %w", scanErr)
		}
		tagIDs = append(tagIDs, tagID)
	}

	rowsErr := rows.Err()
	if rowsErr != nil {
		return nil, fmt.Errorf("iterate subscription tag ids: %w", rowsErr)
	}

	return tagIDs, nil
}

func (r *Repository) deleteSubscription(ctx context.Context, tx pgx.Tx, chatID, linkID int64) error {
	const query = `
		DELETE FROM chat_links
		WHERE chat_id = $1 AND link_id = $2
	`

	tag, err := tx.Exec(ctx, query, chatID, linkID)
	if err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ports.ErrLinkNotFound
	}

	return nil
}

func (r *Repository) deleteUnusedLink(ctx context.Context, tx pgx.Tx, linkID int64) error {
	const query = `
		DELETE FROM links
		WHERE id = $1
		  AND NOT EXISTS (
			SELECT 1
			FROM chat_links
			WHERE link_id = $1
		  )
	`

	_, err := tx.Exec(ctx, query, linkID)
	if err != nil {
		return fmt.Errorf("delete unused link row: %w", err)
	}
	return nil
}

func (r *Repository) deleteUnusedTag(ctx context.Context, tx pgx.Tx, tagID int64) error {
	const query = `
		DELETE FROM tags
		WHERE id = $1
		  AND NOT EXISTS (
			SELECT 1
			FROM link_tags
			WHERE tag_id = $1
		  )
	`

	_, err := tx.Exec(ctx, query, tagID)
	if err != nil {
		return fmt.Errorf("delete unused tag row: %w", err)
	}
	return nil
}
