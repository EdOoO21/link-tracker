package sqlrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

func (r *Repository) TrackLink(ctx context.Context, chatID int64, url string, tags []string) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("begin track link transaction: %w", err)
	}
	defer func() {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			r.logger.Error("failed to rollback track link transaction", "error", rollbackErr, "chatID", chatID, "url", url)
		}
	}()

	present, err := isPresent(ctx, tx, chatID)
	if err != nil {
		return fmt.Errorf("check chat existance: %w", err)
	}
	if !present {
		return ports.ErrChatNotFound
	}

	linkID, err := r.insertOrGetLinkID(ctx, tx, url)
	if err != nil {
		return fmt.Errorf("insert or get link: %w", err)
	}

	subscriptionErr := r.insertSubscription(ctx, tx, chatID, linkID)
	if subscriptionErr != nil {
		return subscriptionErr
	}

	for _, tag := range uniqueStrings(tags) {
		tagID, tagErr := r.insertOrGetTagID(ctx, tx, tag)
		if tagErr != nil {
			return fmt.Errorf("insert or get tag %q: %w", tag, tagErr)
		}

		attachErr := r.insertChatLinkTag(ctx, tx, chatID, linkID, tagID)
		if attachErr != nil {
			return fmt.Errorf("attach tag %q to link: %w", tag, attachErr)
		}
	}

	commitErr := tx.Commit(ctx)
	if commitErr != nil {
		return fmt.Errorf("commit track link transaction: %w", commitErr)
	}

	return nil
}

func (r *Repository) insertOrGetLinkID(ctx context.Context, tx pgx.Tx, url string) (int64, error) {
	const query = `
		INSERT INTO links (url)
		VALUES ($1)
		ON CONFLICT (url) DO UPDATE
		SET url = EXCLUDED.url
		RETURNING id
	`

	var linkID int64
	scanErr := tx.QueryRow(ctx, query, url).Scan(&linkID)
	if scanErr != nil {
		return 0, fmt.Errorf("scan link id: %w", scanErr)
	}

	return linkID, nil
}

func (r *Repository) insertSubscription(ctx context.Context, tx pgx.Tx, chatID, linkID int64) error {
	const query = `
		INSERT INTO chat_links (chat_id, link_id)
		VALUES ($1, $2)
	`

	_, execErr := tx.Exec(ctx, query, chatID, linkID)
	if execErr != nil {
		switch {
		case isPgCode(execErr, pgUniqueViolation):
			return ports.ErrLinkAlreadyExists
		case isPgCode(execErr, pgForeignKeyViolation):
			return ports.ErrChatNotFound
		default:
			return fmt.Errorf("insert chat link: %w", execErr)
		}
	}

	return nil
}

func (r *Repository) insertOrGetTagID(ctx context.Context, tx pgx.Tx, tag string) (int64, error) {
	const query = `
		INSERT INTO tags (name)
		VALUES ($1)
		ON CONFLICT (name) DO UPDATE
		SET name = EXCLUDED.name
		RETURNING id
	`

	var tagID int64
	scanErr := tx.QueryRow(ctx, query, tag).Scan(&tagID)
	if scanErr != nil {
		return 0, fmt.Errorf("scan tag id: %w", scanErr)
	}

	return tagID, nil
}

func (r *Repository) insertChatLinkTag(ctx context.Context, tx pgx.Tx, chatID, linkID, tagID int64) error {
	const query = `
		INSERT INTO link_tags (chat_id, link_id, tag_id)
		VALUES ($1, $2, $3)
	`

	_, execErr := tx.Exec(ctx, query, chatID, linkID, tagID)
	if execErr != nil {
		return fmt.Errorf("insert chat link tag: %w", execErr)
	}

	return nil
}
