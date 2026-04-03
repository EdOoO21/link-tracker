package ormrepo

import (
	"context"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
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
		return r.queryLinksWithoutTagFilter(ctx, chatID)
	}

	return r.queryLinksWithTagFilter(ctx, chatID, tags)
}

func (r *Repository) queryLinksWithoutTagFilter(ctx context.Context, chatID int64) (pgx.Rows, error) {
	ds := dialect.
		From(goqu.T("chat_links").As("cl")).
		Prepared(true).
		Select(
			goqu.I("l.id"),
			goqu.I("l.url"),
			goqu.I("l.last_update"),
			goqu.I("t.name"),
		).
		Join(
			goqu.T("links").As("l"),
			goqu.On(goqu.I("l.id").Eq(goqu.I("cl.link_id"))),
		).
		LeftJoin(
			goqu.T("link_tags").As("lt"),
			goqu.On(
				goqu.I("lt.chat_id").Eq(goqu.I("cl.chat_id")),
				goqu.I("lt.link_id").Eq(goqu.I("cl.link_id")),
			),
		).
		LeftJoin(
			goqu.T("tags").As("t"),
			goqu.On(goqu.I("t.id").Eq(goqu.I("lt.tag_id"))),
		).
		Where(goqu.I("cl.chat_id").Eq(chatID)).
		Order(goqu.I("l.id").Asc(), goqu.I("t.name").Asc())

	sql, args, err := buildSQL(ds)
	if err != nil {
		return nil, fmt.Errorf("build links query without tag filter: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query links without tag filter: %w", err)
	}

	return rows, nil
}

func (r *Repository) queryLinksWithTagFilter(ctx context.Context, chatID int64, tags []string) (pgx.Rows, error) {
	filteredLinks := dialect.
		From(goqu.T("chat_links").As("cl")).
		Prepared(true).
		Select(goqu.I("cl.link_id")).
		Join(
			goqu.T("link_tags").As("lt"),
			goqu.On(
				goqu.I("lt.chat_id").Eq(goqu.I("cl.chat_id")),
				goqu.I("lt.link_id").Eq(goqu.I("cl.link_id")),
			),
		).
		Join(
			goqu.T("tags").As("filter_tags"),
			goqu.On(goqu.I("filter_tags.id").Eq(goqu.I("lt.tag_id"))),
		).
		Where(
			goqu.I("cl.chat_id").Eq(chatID),
			goqu.I("filter_tags.name").In(tags),
		).
		GroupBy(goqu.I("cl.link_id")).
		Having(goqu.COUNT(goqu.DISTINCT(goqu.I("filter_tags.name"))).Eq(len(tags)))

	ds := dialect.
		From(goqu.T("filtered_links").As("fl")).
		With("filtered_links", filteredLinks).
		Prepared(true).
		Select(
			goqu.I("l.id"),
			goqu.I("l.url"),
			goqu.I("l.last_update"),
			goqu.I("t.name"),
		).
		Join(
			goqu.T("links").As("l"),
			goqu.On(goqu.I("l.id").Eq(goqu.I("fl.link_id"))),
		).
		LeftJoin(
			goqu.T("link_tags").As("lt"),
			goqu.On(
				goqu.I("lt.chat_id").Eq(chatID),
				goqu.I("lt.link_id").Eq(goqu.I("fl.link_id")),
			),
		).
		LeftJoin(
			goqu.T("tags").As("t"),
			goqu.On(goqu.I("t.id").Eq(goqu.I("lt.tag_id"))),
		).
		Order(goqu.I("l.id").Asc(), goqu.I("t.name").Asc())

	sql, args, err := buildSQL(ds)
	if err != nil {
		return nil, fmt.Errorf("build links query with tag filter: %w", err)
	}

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("query links with tag filter: %w", err)
	}

	return rows, nil
}
