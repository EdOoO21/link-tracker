package sqlrepo

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

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

func isPresent(ctx context.Context, q queryRower, chatID int64) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1
			FROM chats
			WHERE id = $1
		)
	`

	var exists bool
	scanErr := q.QueryRow(ctx, query, chatID).Scan(&exists)
	if scanErr != nil {
		return false, fmt.Errorf("scan chat existence row: %w", scanErr)
	}

	return exists, nil
}
