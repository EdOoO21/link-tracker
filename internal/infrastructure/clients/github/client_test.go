package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	RequestTimeout = 5 * time.Second
)

func TestGetRepoUpdate(t *testing.T) {
	since := time.Date(2026, 3, 30, 17, 0, 0, 0, time.UTC)
	oldCreatedAt := since.Add(-time.Hour)
	issueCreatedAt := since.Add(time.Hour)
	prCreatedAt := since.Add(2 * time.Hour)
	longBody := strings.Repeat("a", 250)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/user/repo/issues" {
			t.Fatalf("got path %q", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/vnd.github+json" {
			t.Fatalf("got accept header %q", r.Header.Get("Accept"))
		}
		if got := r.URL.Query().Get("state"); got != "all" {
			t.Fatalf("got state query %q", got)
		}
		if got := r.URL.Query().Get("sort"); got != "created" {
			t.Fatalf("got sort query %q", got)
		}
		if got := r.URL.Query().Get("direction"); got != "desc" {
			t.Fatalf("got direction query %q", got)
		}
		if got := r.URL.Query().Get("per_page"); got != "10" {
			t.Fatalf("got per_page query %q", got)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"title":"new pr","body":"pr body","created_at":"` + prCreatedAt.Format(time.RFC3339) + `","user":{"login":"hubot"},"pull_request":{}},
			{"title":"new issue","body":"` + longBody + `","created_at":"` + issueCreatedAt.Format(time.RFC3339) + `","user":{"login":"octocat"}},
			{"title":"old issue","body":"old body","created_at":"` + oldCreatedAt.Format(time.RFC3339) + `","user":{"login":"legacy"}}
		]`))
	}))
	defer server.Close()

	client := NewGitHubClient()
	client.baseURL = server.URL
	client.client = server.Client()
	reqCtx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	update, err := client.GetRepoUpdate(reqCtx, "user", "repo", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !update.HasUpdate {
		t.Fatal("expected update to be detected")
	}
	if !update.LastUpdate.Equal(prCreatedAt) {
		t.Fatalf("got last update %v, want %v", update.LastUpdate, prCreatedAt)
	}
	if !strings.Contains(update.Message, "Pull Request") {
		t.Fatalf("message %q does not contain pull request info", update.Message)
	}
	if !strings.Contains(update.Message, "Issue") {
		t.Fatalf("message %q does not contain issue info", update.Message)
	}
	if strings.Contains(update.Message, "old issue") {
		t.Fatalf("message %q should not contain old issue", update.Message)
	}
	if !strings.Contains(update.Message, string([]rune(longBody)[:previewLimit])) {
		t.Fatalf("message %q does not contain truncated preview", update.Message)
	}
}

func TestGetRepoUpdateWithoutNewIssues(t *testing.T) {
	since := time.Date(2026, 3, 30, 17, 0, 0, 0, time.UTC)
	oldCreatedAt := since.Add(-time.Hour)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[
			{"title":"old issue","body":"old body","created_at":"` + oldCreatedAt.Format(time.RFC3339) + `","user":{"login":"legacy"}}
		]`))
	}))
	defer server.Close()

	client := NewGitHubClient()
	client.baseURL = server.URL
	client.client = server.Client()
	reqCtx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	update, err := client.GetRepoUpdate(reqCtx, "user", "repo", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if update.HasUpdate {
		t.Fatal("expected no updates")
	}
	if !update.LastUpdate.Equal(since) {
		t.Fatalf("got last update %v, want %v", update.LastUpdate, since)
	}
	if update.Message != "" {
		t.Fatalf("got message %q, want empty", update.Message)
	}
}

func TestGetRepoUpdateErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewGitHubClient()
	client.baseURL = server.URL
	client.client = server.Client()
	reqCtx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()
	if _, err := client.GetRepoUpdate(reqCtx, "user", "repo", time.Time{}); err == nil || !strings.Contains(err.Error(), "unexpected status") {
		t.Fatalf("got err %v", err)
	}
}

func TestParseGitHubURL(t *testing.T) {
	client := NewGitHubClient()
	owner, repo, err := client.ParseGitHubURL("https://github.com/user/repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if owner != "user" || repo != "repo" {
		t.Fatalf("got owner=%q repo=%q", owner, repo)
	}

	if _, _, parseErr := client.ParseGitHubURL("https://google.com/user/repo"); parseErr == nil {
		t.Fatal("expected error for non-github url")
	}
	if _, _, parseErr := client.ParseGitHubURL("https://github.com/user"); parseErr == nil {
		t.Fatal("expected error for invalid repo url")
	}
}
