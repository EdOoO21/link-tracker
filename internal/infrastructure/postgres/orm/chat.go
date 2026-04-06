package ormrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/doug-martin/goqu/v9"
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
	ds := dialect.
		Insert("chats").
		Prepared(true).
		Rows(goqu.Record{"id": chatID})

	sql, args, err := buildSQL(ds)
	if err != nil {
		return fmt.Errorf("build insert chat query: %w", err)
	}

	_, err = r.pool.Exec(ctx, sql, args...)
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
	return r.listDistinctInt64ByChat(ctx, tx, "chat_links", "link_id", chatID)
}

func (r *Repository) listTagIDsByChat(ctx context.Context, tx pgx.Tx, chatID int64) ([]int64, error) {
	return r.listDistinctInt64ByChat(ctx, tx, "link_tags", "tag_id", chatID)
}

func (r *Repository) listDistinctInt64ByChat(
	ctx context.Context,
	tx pgx.Tx,
	table string,
	column string,
	chatID int64,
) ([]int64, error) {
	ds := dialect.
		From(table).
		Prepared(true).
		Select(goqu.C(column)).
		Where(goqu.C("chat_id").Eq(chatID)).
		Distinct()

	sql, args, err := buildSQL(ds)
	if err != nil {
		return nil, fmt.Errorf("build %s query for chat %d: %w", column, chatID, err)
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query %s for chat %d: %w", column, chatID, err)
	}
	defer rows.Close()

	values := make([]int64, 0)
	for rows.Next() {
		var value int64
		scanErr := rows.Scan(&value)
		if scanErr != nil {
			return nil, fmt.Errorf("scan %s for chat %d: %w", column, chatID, scanErr)
		}
		values = append(values, value)
	}

	rowsErr := rows.Err()
	if rowsErr != nil {
		return nil, fmt.Errorf("iterate %s for chat %d: %w", column, chatID, rowsErr)
	}

	return values, nil
}

func (r *Repository) deleteChatRow(ctx context.Context, tx pgx.Tx, chatID int64) error {
	ds := dialect.
		Delete("chats").
		Prepared(true).
		Where(goqu.C("id").Eq(chatID))

	sql, args, err := buildSQL(ds)
	if err != nil {
		return fmt.Errorf("build delete chat query: %w", err)
	}

	tag, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("delete chat: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ports.ErrChatNotFound
	}

	return nil
}
