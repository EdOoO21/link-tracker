package sqlrepo_test

import (
	"context"
	"testing"

	contracttest "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/postgres/contract_test"
	sqlrepo "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/postgres/sql"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
	settings "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/settings/scrapper"
)

func TestRepositoryContract(t *testing.T) {
	contracttest.RunRepositorySuite(t, func(
		ctx context.Context,
		logger ports.Logger,
		cfg settings.DBConfig,
	) (contracttest.Repository, error) {
		return sqlrepo.NewRepository(ctx, logger, cfg)
	})
}
