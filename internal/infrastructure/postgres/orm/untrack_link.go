package ormrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/doug-martin/goqu/v9"
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
	ds := dialect.
		From(goqu.T("chat_links").As("cl")).
		Prepared(true).
		Select(goqu.I("l.id")).
		Join(
			goqu.T("links").As("l"),
			goqu.On(goqu.I("l.id").Eq(goqu.I("cl.link_id"))),
		).
		Where(
			goqu.I("cl.chat_id").Eq(chatID),
			goqu.I("l.url").Eq(url),
		)

	sql, args, err := buildSQL(ds)
	if err != nil {
		return 0, fmt.Errorf("build get subscription link query: %w", err)
	}

	var linkID int64
	scanErr := tx.QueryRow(ctx, sql, args...).Scan(&linkID)
	if scanErr != nil {
		return 0, fmt.Errorf("scan subscription link id: %w", scanErr)
	}

	return linkID, nil
}

func (r *Repository) listTagIDsByChatAndLink(ctx context.Context, tx pgx.Tx, chatID, linkID int64) ([]int64, error) {
	ds := dialect.
		From("link_tags").
		Prepared(true).
		Select(goqu.C("tag_id")).
		Where(goqu.Ex{
			"chat_id": chatID,
			"link_id": linkID,
		})

	sql, args, err := buildSQL(ds)
	if err != nil {
		return nil, fmt.Errorf("build subscription tag ids query: %w", err)
	}

	rows, err := tx.Query(ctx, sql, args...)
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
	ds := dialect.
		Delete("chat_links").
		Prepared(true).
		Where(goqu.Ex{
			"chat_id": chatID,
			"link_id": linkID,
		})

	sql, args, err := buildSQL(ds)
	if err != nil {
		return fmt.Errorf("build delete subscription query: %w", err)
	}

	tag, err := tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("delete subscription: %w", err)
	}

	if tag.RowsAffected() == 0 {
		return ports.ErrLinkNotFound
	}

	return nil
}

func (r *Repository) deleteUnusedLink(ctx context.Context, tx pgx.Tx, linkID int64) error {
	subquery := dialect.
		From("chat_links").
		Prepared(true).
		Select(goqu.V(1)).
		Where(goqu.C("link_id").Eq(linkID))

	ds := dialect.
		Delete("links").
		Prepared(true).
		Where(
			goqu.C("id").Eq(linkID),
			goqu.L("NOT EXISTS (?)", subquery),
		)

	sql, args, err := buildSQL(ds)
	if err != nil {
		return fmt.Errorf("build delete unused link query: %w", err)
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("delete unused link row: %w", err)
	}
	return nil
}

func (r *Repository) deleteUnusedTag(ctx context.Context, tx pgx.Tx, tagID int64) error {
	subquery := dialect.
		From("link_tags").
		Prepared(true).
		Select(goqu.V(1)).
		Where(goqu.C("tag_id").Eq(tagID))

	ds := dialect.
		Delete("tags").
		Prepared(true).
		Where(
			goqu.C("id").Eq(tagID),
			goqu.L("NOT EXISTS (?)", subquery),
		)

	sql, args, err := buildSQL(ds)
	if err != nil {
		return fmt.Errorf("build delete unused tag query: %w", err)
	}

	_, err = tx.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("delete unused tag row: %w", err)
	}
	return nil
}
