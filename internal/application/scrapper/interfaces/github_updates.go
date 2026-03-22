package scrapperinterfaces

import "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"

type GithubUpdates interface {
	GetRepoUpdate(owner, repo string) (ports.ResourceUpdate, error)

	ParseGitHubURL(url string) (owner, repo string, err error)
}
