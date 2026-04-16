package github

import (
	"fmt"
	"strings"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

type githubIssuesResponse []githubIssueResponse

type githubIssueResponse struct {
	Title       string              `json:"title"`
	Body        string              `json:"body"`
	CreatedAt   time.Time           `json:"created_at"`
	User        githubUserResponse  `json:"user"`
	PullRequest *githubPullResponse `json:"pull_request"`
}

type githubUserResponse struct {
	Login string `json:"login"`
}

type githubPullResponse struct{}

const previewLimit = 50

func (r githubIssuesResponse) toResourceUpdate(since time.Time) ports.ResourceUpdate {
	newIssues := r.newerThan(since)
	if len(newIssues) == 0 {
		return ports.ResourceUpdate{
			HasUpdate:  false,
			LastUpdate: since,
			Message:    "",
		}
	}

	return ports.ResourceUpdate{
		HasUpdate:  true,
		LastUpdate: newIssues.latestCreatedAt(),
		Message:    newIssues.message(),
	}
}

func (r githubIssuesResponse) newerThan(since time.Time) githubIssuesResponse {
	filtered := make(githubIssuesResponse, 0, len(r))
	for _, issue := range r {
		if issue.CreatedAt.After(since) {
			filtered = append(filtered, issue)
		}
	}

	return filtered
}

func (r githubIssuesResponse) latestCreatedAt() time.Time {
	latest := r[0].CreatedAt
	for _, issue := range r[1:] {
		if issue.CreatedAt.After(latest) {
			latest = issue.CreatedAt
		}
	}

	return latest
}

func (r githubIssuesResponse) message() string {
	var builder strings.Builder
	builder.WriteString("Обнаружены новые GitHub обновления:\n\n")

	for i, issue := range r {
		fmt.Fprintf(&builder, "%d. %s\n", i+1, issue.kind())
		builder.WriteString("Название: ")
		builder.WriteString(strings.TrimSpace(issue.Title))
		builder.WriteString("\nАвтор: ")
		builder.WriteString(strings.TrimSpace(issue.User.Login))
		builder.WriteString("\nСоздано: ")
		builder.WriteString(issue.CreatedAt.Format(time.RFC3339))
		builder.WriteString("\nПревью: ")
		builder.WriteString(issue.preview())
		if i != len(r)-1 {
			builder.WriteString("\n\n")
		}
	}

	return builder.String()
}

func (r githubIssueResponse) kind() string {
	if r.PullRequest != nil {
		return "Pull Request"
	}

	return "Issue"
}

func (r githubIssueResponse) preview() string {
	body := strings.TrimSpace(r.Body)
	if body == "" {
		return "отсутствует"
	}

	runes := []rune(body)
	if len(runes) <= previewLimit {
		return body
	}

	return strings.TrimSpace(string(runes[:previewLimit]))
}
