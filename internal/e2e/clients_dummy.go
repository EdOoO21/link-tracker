package e2e

import (
	"context"
	"fmt"
	"time"

	scrapperinterfaces "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/interfaces"
	githubclient "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/clients/github"
	stackoverflowclient "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/clients/stackoverflow"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

func NewCheckers() []scrapperinterfaces.Checker {
	return []scrapperinterfaces.Checker{
		NewGitHubClient(),
		NewStackOverflowClient(),
	}
}

type GitHubClient struct {
	parser *githubclient.Client
}

func NewGitHubClient() *GitHubClient {
	return &GitHubClient{parser: githubclient.NewGitHubClient()}
}

func (c *GitHubClient) CanHandle(raw string) bool {
	_, _, err := c.parser.ParseGitHubURL(raw)
	return err == nil
}

func (c *GitHubClient) CheckUpdate(_ context.Context, raw string, _ time.Time) (ports.ResourceUpdate, error) {
	if _, _, err := c.parser.ParseGitHubURL(raw); err != nil {
		return ports.ResourceUpdate{}, fmt.Errorf("parse github url: %w", err)
	}

	return ports.ResourceUpdate{
		HasUpdate:  true,
		LastUpdate: time.Now().UTC().Add(time.Second),
		Message:    "Обнаружено тестовое GitHub обновление.",
	}, nil
}

type StackOverflowClient struct {
	parser *stackoverflowclient.Client
}

func NewStackOverflowClient() *StackOverflowClient {
	return &StackOverflowClient{parser: stackoverflowclient.NewStackOverflowClient()}
}

func (c *StackOverflowClient) CanHandle(raw string) bool {
	_, err := c.parser.ParseStackOverflowURL(raw)
	return err == nil
}

func (c *StackOverflowClient) CheckUpdate(_ context.Context, raw string, _ time.Time) (ports.ResourceUpdate, error) {
	if _, err := c.parser.ParseStackOverflowURL(raw); err != nil {
		return ports.ResourceUpdate{}, fmt.Errorf("parse stackoverflow url: %w", err)
	}

	return ports.ResourceUpdate{
		HasUpdate:  true,
		LastUpdate: time.Now().UTC().Add(time.Second),
		Message:    "Обнаружено тестовое StackOverflow обновление.",
	}, nil
}
