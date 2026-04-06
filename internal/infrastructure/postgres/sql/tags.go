package sqlrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

func (r *Repository) AddTag(ctx context.Context, chatID int64, url, tag string) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("begin add tag transaction: %w", err)
	}
	defer func() {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			r.logger.Error("failed to rollback add tag transaction", "error", rollbackErr, "chatID", chatID, "url", url, "tag", tag)
		}
	}()

	present, err := isPresent(ctx, tx, chatID)
	if err != nil {
		return fmt.Errorf("check chat existance: %w", err)
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

	tagID, err := r.insertOrGetTagID(ctx, tx, tag)
	if err != nil {
		return fmt.Errorf("insert or get tag: %w", err)
	}

	addErr := r.addTagToLink(ctx, tx, chatID, linkID, tagID)
	if addErr != nil {
		return addErr
	}

	commitErr := tx.Commit(ctx)
	if commitErr != nil {
		return fmt.Errorf("commit add tag transaction: %w", commitErr)
	}

	return nil
}

func (r *Repository) DeleteTag(ctx context.Context, chatID int64, url, tag string) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return fmt.Errorf("begin delete tag transaction: %w", err)
	}
	defer func() {
		rollbackErr := tx.Rollback(ctx)
		if rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			r.logger.Error("failed to rollback delete tag transaction", "error", rollbackErr, "chatID", chatID, "url", url, "tag", tag)
		}
	}()

	present, err := isPresent(ctx, tx, chatID)
	if err != nil {
		return fmt.Errorf("check chat existance: %w", err)
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

	tagID, err := r.getTagIDByName(ctx, tx, tag)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ports.ErrTagNotFound
		}
		return fmt.Errorf("get tag id: %w", err)
	}

	deleteErr := r.deleteTagFromLink(ctx, tx, chatID, linkID, tagID)
	if deleteErr != nil {
		return deleteErr
	}

	cleanupErr := r.deleteUnusedTag(ctx, tx, tagID)
	if cleanupErr != nil {
		return fmt.Errorf("delete unused tag %d: %w", tagID, cleanupErr)
	}

	commitErr := tx.Commit(ctx)
	if commitErr != nil {
		return fmt.Errorf("commit delete tag transaction: %w", commitErr)
	}

	return nil
}

func (r *Repository) GetTags(ctx context.Context, chatID int64, url string) ([]string, error) {
	present, err := isPresent(ctx, r.pool, chatID)
	if err != nil {
		return nil, fmt.Errorf("check chat existence before get tags: %w", err)
	}
	if !present {
		return nil, ports.ErrChatNotFound
	}

	const linkQuery = `
		SELECT l.id
		FROM chat_links cl
		JOIN links l ON l.id = cl.link_id
		WHERE cl.chat_id = $1 AND l.url = $2
	`

	var linkID int64
	scanErr := r.pool.QueryRow(ctx, linkQuery, chatID, url).Scan(&linkID)
	if scanErr != nil {
		if errors.Is(scanErr, pgx.ErrNoRows) {
			return nil, ports.ErrLinkNotFound
		}
		return nil, fmt.Errorf("scan subscription link id: %w", scanErr)
	}

	const query = `
		SELECT t.name
		FROM link_tags lt
		JOIN tags t ON t.id = lt.tag_id
		WHERE lt.chat_id = $1 AND lt.link_id = $2
		ORDER BY t.name
	`

	rows, queryErr := r.pool.Query(ctx, query, chatID, linkID)
	if queryErr != nil {
		return nil, fmt.Errorf("query link tags: %w", queryErr)
	}
	defer rows.Close()

	tags := make([]string, 0)
	for rows.Next() {
		var tagName string
		tagScanErr := rows.Scan(&tagName)
		if tagScanErr != nil {
			return nil, fmt.Errorf("scan tag name: %w", tagScanErr)
		}
		tags = append(tags, tagName)
	}

	rowsErr := rows.Err()
	if rowsErr != nil {
		return nil, fmt.Errorf("iterate tag rows: %w", rowsErr)
	}

	return tags, nil
}

func (r *Repository) addTagToLink(ctx context.Context, tx pgx.Tx, chatID, linkID, tagID int64) error {
	const query = `
		INSERT INTO link_tags (chat_id, link_id, tag_id)
		VALUES ($1, $2, $3)
	`

	_, execErr := tx.Exec(ctx, query, chatID, linkID, tagID)
	if execErr != nil {
		if isPgCode(execErr, pgUniqueViolation) {
			return ports.ErrTagAlreadyExists
		}
		return fmt.Errorf("insert link tag: %w", execErr)
	}

	return nil
}

func (r *Repository) getTagIDByName(ctx context.Context, tx pgx.Tx, tag string) (int64, error) {
	const query = `
		SELECT id
		FROM tags
		WHERE name = $1
	`

	var tagID int64
	scanErr := tx.QueryRow(ctx, query, tag).Scan(&tagID)
	if scanErr != nil {
		return 0, fmt.Errorf("scan tag id by name: %w", scanErr)
	}

	return tagID, nil
}

func (r *Repository) deleteTagFromLink(ctx context.Context, tx pgx.Tx, chatID, linkID, tagID int64) error {
	const query = `
		DELETE FROM link_tags
		WHERE chat_id = $1 AND link_id = $2 AND tag_id = $3
	`

	tag, err := tx.Exec(ctx, query, chatID, linkID, tagID)
	if err != nil {
		return fmt.Errorf("delete link tag: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ports.ErrTagNotFound
	}

	return nil
}
