package sqlrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

func (r *Repository) IsPresent(ctx context.Context, chatID int64) (bool, error) {
	exists, err := isPresent(ctx, r.pool, chatID)
	if err != nil {
		return false, fmt.Errorf("query chat existence: %w", err)
	}

	return exists, nil
}

func (r *Repository) AddChat(ctx context.Context, chatID int64) error {
	const query = `
		INSERT INTO chats (id)
		VALUES ($1)
	`

	_, err := r.pool.Exec(ctx, query, chatID)
	if err != nil {
		if isPgCode(err, pgUniqueViolation) {
			return ports.ErrChatAlreadyExists
		}
		return fmt.Errorf("insert chat: %w", err)
	}

	return nil
}

func (r *Repository) DeleteChat(ctx context.Context, chatID int64) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("begin delete chat transaction: %w", err)
	}
	defer func() {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			r.logger.Error("failed to rollback delete chat transaction", "error", rollbackErr, "chatID", chatID)
		}
	}()

	present, err := isPresent(ctx, tx, chatID)
	if err != nil {
		return fmt.Errorf("check chat existance: %w", err)
	}
	if !present {
		return ports.ErrChatNotFound
	}

	linkIDs, err := r.listLinkIDsByChat(ctx, tx, chatID)
	if err != nil {
		return fmt.Errorf("list chat links: %w", err)
	}

	tagIDs, err := r.listTagIDsByChat(ctx, tx, chatID)
	if err != nil {
		return fmt.Errorf("list chat tags: %w", err)
	}

	deleteErr := r.deleteChatRow(ctx, tx, chatID)
	if deleteErr != nil {
		return deleteErr
	}

	for _, linkID := range linkIDs {
		linkCleanupErr := r.deleteUnusedLink(ctx, tx, linkID)
		if linkCleanupErr != nil {
			return fmt.Errorf("delete unused link %d: %w", linkID, linkCleanupErr)
		}
	}

	for _, tagID := range tagIDs {
		tagCleanupErr := r.deleteUnusedTag(ctx, tx, tagID)
		if tagCleanupErr != nil {
			return fmt.Errorf("delete unused tag %d: %w", tagID, tagCleanupErr)
		}
	}

	commitErr := tx.Commit(ctx)
	if commitErr != nil {
		return fmt.Errorf("commit delete chat transaction: %w", commitErr)
	}

	return nil
}

func (r *Repository) listLinkIDsByChat(ctx context.Context, tx pgx.Tx, chatID int64) ([]int64, error) {
	const query = `
		SELECT DISTINCT link_id
		FROM chat_links
		WHERE chat_id = $1
	`

	rows, err := tx.Query(ctx, query, chatID)
	if err != nil {
		return nil, fmt.Errorf("query chat link ids: %w", err)
	}
	defer rows.Close()

	linkIDs := make([]int64, 0)
	for rows.Next() {
		var linkID int64
		scanErr := rows.Scan(&linkID)
		if scanErr != nil {
			return nil, fmt.Errorf("scan chat link id: %w", scanErr)
		}
		linkIDs = append(linkIDs, linkID)
	}

	rowsErr := rows.Err()
	if rowsErr != nil {
		return nil, fmt.Errorf("iterate chat link ids: %w", rowsErr)
	}

	return linkIDs, nil
}

func (r *Repository) listTagIDsByChat(ctx context.Context, tx pgx.Tx, chatID int64) ([]int64, error) {
	const query = `
		SELECT DISTINCT tag_id
		FROM link_tags
		WHERE chat_id = $1
	`

	rows, err := tx.Query(ctx, query, chatID)
	if err != nil {
		return nil, fmt.Errorf("query chat tag ids: %w", err)
	}
	defer rows.Close()

	tagIDs := make([]int64, 0)
	for rows.Next() {
		var tagID int64
		scanErr := rows.Scan(&tagID)
		if scanErr != nil {
			return nil, fmt.Errorf("scan chat tag id: %w", scanErr)
		}
		tagIDs = append(tagIDs, tagID)
	}

	rowsErr := rows.Err()
	if rowsErr != nil {
		return nil, fmt.Errorf("iterate chat tag ids: %w", rowsErr)
	}

	return tagIDs, nil
}

func (r *Repository) deleteChatRow(ctx context.Context, tx pgx.Tx, chatID int64) error {
	const query = `
		DELETE FROM chats
		WHERE id = $1
	`

	tag, err := tx.Exec(ctx, query, chatID)
	if err != nil {
		return fmt.Errorf("delete chat: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ports.ErrChatNotFound
	}

	return nil
}
