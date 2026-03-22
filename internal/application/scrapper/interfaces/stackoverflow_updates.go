package scrapperinterfaces

import "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"

type StackOverflowUpdates interface {
	GetQuestionUpdate(questionID string) (ports.ResourceUpdate, error)

	ParseStackOverflowURL(url string) (string, error)
}
