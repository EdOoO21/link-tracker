package sourcedummy

import (
	"fmt"
	"time"

	githubclient "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/clients/github"
	stackoverflowclient "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/clients/stackoverflow"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

type GitHubClient struct {
	parser *githubclient.Client
}

func NewGitHubClient() *GitHubClient {
	return &GitHubClient{parser: githubclient.NewGitHubClient()}
}

func (c *GitHubClient) GetRepoUpdate(_, _ string) (ports.ResourceUpdate, error) {
	return ports.ResourceUpdate{LastUpdate: time.Now().UTC().Add(time.Second)}, nil
}

func (c *GitHubClient) ParseGitHubURL(url string) (owner, repo string, err error) {
	owner, repo, err = c.parser.ParseGitHubURL(url)
	if err != nil {
		return "", "", fmt.Errorf("parse github url: %w", err)
	}
	return owner, repo, nil
}

type StackOverflowClient struct {
	parser *stackoverflowclient.Client
}

func NewStackOverflowClient() *StackOverflowClient {
	return &StackOverflowClient{parser: stackoverflowclient.NewStackOverflowClient()}
}

func (c *StackOverflowClient) GetQuestionUpdate(_ string) (ports.ResourceUpdate, error) {
	return ports.ResourceUpdate{LastUpdate: time.Now().UTC().Add(time.Second)}, nil
}

func (c *StackOverflowClient) ParseStackOverflowURL(url string) (string, error) {
	questionID, err := c.parser.ParseStackOverflowURL(url)
	if err != nil {
		return "", fmt.Errorf("parse stackoverflow url: %w", err)
	}
	return questionID, nil
}
