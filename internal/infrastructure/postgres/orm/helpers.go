package ormrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	// Register the postgres dialect for goqu.
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var dialect = goqu.Dialect("postgres")

type sqlBuilder interface {
	ToSQL() (string, []interface{}, error)
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))

	for _, value := range values {
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}

	return result
}

func isPgCode(err error, code string) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}

	return pgErr.Code == code
}

type queryRower interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func buildSQL(ds sqlBuilder) (string, []any, error) {
	sql, args, err := ds.ToSQL()
	if err != nil {
		return "", nil, fmt.Errorf("build goqu SQL: %w", err)
	}

	return sql, args, nil
}

func isPresent(ctx context.Context, q queryRower, chatID int64) (bool, error) {
	ds := dialect.
		From("chats").
		Prepared(true).
		Select(goqu.COUNT(goqu.Star())).
		Where(goqu.C("id").Eq(chatID))

	sql, args, err := buildSQL(ds)
	if err != nil {
		return false, err
	}

	var count int64
	scanErr := q.QueryRow(ctx, sql, args...).Scan(&count)
	if scanErr != nil {
		return false, fmt.Errorf("scan chat existence row: %w", scanErr)
	}

	return count > 0, nil
}
