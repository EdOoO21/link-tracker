package scrapperinterfaces

import (
	"context"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

type GithubUpdates interface {
	GetRepoUpdate(ctx context.Context, owner, repo string) (ports.ResourceUpdate, error)

	ParseGitHubURL(url string) (owner, repo string, err error)
}
