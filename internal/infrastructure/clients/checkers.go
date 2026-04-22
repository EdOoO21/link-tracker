package clients

import (
	scrapperinterfaces "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/interfaces"
	github "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/clients/github"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/clients/stackoverflow"
)

func NewCheckers() []scrapperinterfaces.Checker {
	return []scrapperinterfaces.Checker{
		github.NewGitHubClient(),
		stackoverflow.NewStackOverflowClient(),
	}
}
