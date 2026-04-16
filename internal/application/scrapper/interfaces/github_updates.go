package scrapperinterfaces

import (
	"context"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

type GithubUpdates interface {
	GetRepoUpdate(ctx context.Context, owner, repo string, since time.Time) (ports.ResourceUpdate, error)

	ParseGitHubURL(url string) (owner, repo string, err error)
}
