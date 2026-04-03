package ormrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/doug-martin/goqu/v9"
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
		return fmt.Errorf("check chat presence: %w", err)
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
	ds := dialect.
		Insert("links").
		Prepared(true).
		Rows(goqu.Record{"url": url}).
		OnConflict(goqu.DoUpdate("url", goqu.Record{"url": goqu.I("excluded.url")})).
		Returning("id")

	sql, args, err := buildSQL(ds)
	if err != nil {
		return 0, fmt.Errorf("build insert or get link query: %w", err)
	}

	var linkID int64
	scanErr := tx.QueryRow(ctx, sql, args...).Scan(&linkID)
	if scanErr != nil {
		return 0, fmt.Errorf("scan link id: %w", scanErr)
	}

	return linkID, nil
}

func (r *Repository) insertSubscription(ctx context.Context, tx pgx.Tx, chatID, linkID int64) error {
	ds := dialect.
		Insert("chat_links").
		Prepared(true).
		Rows(goqu.Record{
			"chat_id": chatID,
			"link_id": linkID,
		})

	sql, args, err := buildSQL(ds)
	if err != nil {
		return fmt.Errorf("build insert chat link query: %w", err)
	}

	_, execErr := tx.Exec(ctx, sql, args...)
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
	ds := dialect.
		Insert("tags").
		Prepared(true).
		Rows(goqu.Record{"name": tag}).
		OnConflict(goqu.DoUpdate("name", goqu.Record{"name": goqu.I("excluded.name")})).
		Returning("id")

	sql, args, err := buildSQL(ds)
	if err != nil {
		return 0, fmt.Errorf("build insert or get tag query: %w", err)
	}

	var tagID int64
	scanErr := tx.QueryRow(ctx, sql, args...).Scan(&tagID)
	if scanErr != nil {
		return 0, fmt.Errorf("scan tag id: %w", scanErr)
	}

	return tagID, nil
}

func (r *Repository) insertChatLinkTag(ctx context.Context, tx pgx.Tx, chatID, linkID, tagID int64) error {
	ds := dialect.
		Insert("link_tags").
		Prepared(true).
		Rows(goqu.Record{
			"chat_id": chatID,
			"link_id": linkID,
			"tag_id":  tagID,
		})

	sql, args, err := buildSQL(ds)
	if err != nil {
		return fmt.Errorf("build insert chat link tag query: %w", err)
	}

	_, execErr := tx.Exec(ctx, sql, args...)
	if execErr != nil {
		return fmt.Errorf("insert chat link tag: %w", execErr)
	}

	return nil
}
