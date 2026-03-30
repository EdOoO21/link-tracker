package github

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetRepoUpdate(t *testing.T) {
	updatedAt := time.Date(2026, 3, 30, 18, 0, 0, 0, time.UTC)
	pushedAt := updatedAt.Add(-time.Hour)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/user/repo" {
			t.Fatalf("got path %q", r.URL.Path)
		}
		if r.Header.Get("Accept") != "application/vnd.github+json" {
			t.Fatalf("got accept header %q", r.Header.Get("Accept"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"pushed_at":"` + pushedAt.Format(time.RFC3339) + `","updated_at":"` + updatedAt.Format(time.RFC3339) + `"}`))
	}))
	defer server.Close()

	client := NewGitHubClient()
	client.baseURL = server.URL
	client.client = server.Client()

	update, err := client.GetRepoUpdate("user", "repo")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !update.LastUpdate.Equal(updatedAt) {
		t.Fatalf("got last update %v, want %v", update.LastUpdate, updatedAt)
	}
}

func TestGetRepoUpdateErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewGitHubClient()
	client.baseURL = server.URL
	client.client = server.Client()

	if _, err := client.GetRepoUpdate("user", "repo"); err == nil || !strings.Contains(err.Error(), "unexpected status") {
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

	if _, _, err := client.ParseGitHubURL("https://google.com/user/repo"); err == nil {
		t.Fatal("expected error for non-github url")
	}
	if _, _, err := client.ParseGitHubURL("https://github.com/user"); err == nil {
		t.Fatal("expected error for invalid repo url")
	}
}
