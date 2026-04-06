package ormrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/doug-martin/goqu/v9"
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

	linkID, err := r.getLinkIDByChatAndURLFromPool(ctx, chatID, url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ports.ErrLinkNotFound
		}
		return nil, fmt.Errorf("get subscription link: %w", err)
	}

	ds := dialect.
		From(goqu.T("link_tags").As("lt")).
		Prepared(true).
		Select(goqu.I("t.name")).
		Join(
			goqu.T("tags").As("t"),
			goqu.On(goqu.I("t.id").Eq(goqu.I("lt.tag_id"))),
		).
		Where(
			goqu.I("lt.chat_id").Eq(chatID),
			goqu.I("lt.link_id").Eq(linkID),
		).
		Order(goqu.I("t.name").Asc())

	sql, args, buildErr := buildSQL(ds)
	if buildErr != nil {
		return nil, fmt.Errorf("build get tags query: %w", buildErr)
	}

	rows, queryErr := r.pool.Query(ctx, sql, args...)
	if queryErr != nil {
		return nil, fmt.Errorf("query link tags: %w", queryErr)
	}
	defer rows.Close()

	tags := make([]string, 0)
	for rows.Next() {
		var tagName string
		scanErr := rows.Scan(&tagName)
		if scanErr != nil {
			return nil, fmt.Errorf("scan tag name: %w", scanErr)
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
		return fmt.Errorf("build add tag to link query: %w", err)
	}

	_, execErr := tx.Exec(ctx, sql, args...)
	if execErr != nil {
		if isPgCode(execErr, pgUniqueViolation) {
			return ports.ErrTagAlreadyExists
		}
		return fmt.Errorf("insert link tag: %w", execErr)
	}

	return nil
}

func (r *Repository) getTagIDByName(ctx context.Context, tx pgx.Tx, tag string) (int64, error) {
	ds := dialect.
		From("tags").
		Prepared(true).
		Select(goqu.C("id")).
		Where(goqu.C("name").Eq(tag))

	sql, args, err := buildSQL(ds)
	if err != nil {
		return 0, fmt.Errorf("build get tag id query: %w", err)
	}

	var tagID int64
	scanErr := tx.QueryRow(ctx, sql, args...).Scan(&tagID)
	if scanErr != nil {
		return 0, fmt.Errorf("scan tag id by name: %w", scanErr)
	}

	return tagID, nil
}

func (r *Repository) deleteTagFromLink(ctx context.Context, tx pgx.Tx, chatID, linkID, tagID int64) error {
	ds := dialect.
		Delete("link_tags").
		Prepared(true).
		Where(goqu.Ex{
			"chat_id": chatID,
			"link_id": linkID,
			"tag_id":  tagID,
		})

	sql, args, err := buildSQL(ds)
	if err != nil {
		return fmt.Errorf("build delete link tag query: %w", err)
	}

	tag, execErr := tx.Exec(ctx, sql, args...)
	if execErr != nil {
		return fmt.Errorf("delete link tag: %w", execErr)
	}

	if tag.RowsAffected() == 0 {
		return ports.ErrTagNotFound
	}

	return nil
}

func (r *Repository) getLinkIDByChatAndURLFromPool(ctx context.Context, chatID int64, url string) (int64, error) {
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
		return 0, fmt.Errorf("build get link by chat and url query: %w", err)
	}

	var linkID int64
	scanErr := r.pool.QueryRow(ctx, sql, args...).Scan(&linkID)
	if scanErr != nil {
		return 0, fmt.Errorf("scan subscription link id: %w", scanErr)
	}

	return linkID, nil
}
