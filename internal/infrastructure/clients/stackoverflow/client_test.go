package stackoverflow

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const (
	RequestTimeout = 5 * time.Second
)

func TestGetQuestionUpdate(t *testing.T) {
	since := time.Date(2026, 3, 30, 17, 0, 0, 0, time.UTC)
	oldCreatedAt := since.Add(-time.Hour)
	answerCreatedAt := since.Add(time.Hour)
	commentCreatedAt := since.Add(2 * time.Hour)
	longBody := strings.Repeat("b", 250)

	server := newStackOverflowUpdatesServer(t, oldCreatedAt, answerCreatedAt, commentCreatedAt, longBody)
	defer server.Close()

	client := NewStackOverflowClient()
	client.baseURL = server.URL
	client.client = server.Client()
	reqCtx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	update, err := client.GetQuestionUpdate(reqCtx, "123", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !update.HasUpdate {
		t.Fatal("expected update to be detected")
	}
	if !update.LastUpdate.Equal(commentCreatedAt) {
		t.Fatalf("got last update %v, want %v", update.LastUpdate, commentCreatedAt)
	}
	if !strings.Contains(update.Message, "Тема вопроса: How to test updates?") {
		t.Fatalf("message %q does not contain question title", update.Message)
	}
	if !strings.Contains(update.Message, "Ответы:") {
		t.Fatalf("message %q does not contain answers section", update.Message)
	}
	if !strings.Contains(update.Message, "Комментарии:") {
		t.Fatalf("message %q does not contain comments section", update.Message)
	}
	if strings.Contains(update.Message, "old answer") || strings.Contains(update.Message, "old comment") {
		t.Fatalf("message %q should not contain old updates", update.Message)
	}
	if !strings.Contains(update.Message, string([]rune(longBody)[:previewLimit])) {
		t.Fatalf("message %q does not contain truncated preview", update.Message)
	}
}

func newStackOverflowUpdatesServer(
	t *testing.T,
	oldCreatedAt time.Time,
	answerCreatedAt time.Time,
	commentCreatedAt time.Time,
	longBody string,
) *httptest.Server {
	t.Helper()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/questions/123":
			requireQuestionQuery(t, r)
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"items":[{"title":"How to test updates?"}]}`)
		case "/questions/123/answers":
			requireCollectionQuery(t, r)
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{"items":[
				{"body":"%s","creation_date":%d,"owner":{"display_name":"alice"}},
				{"body":"old answer","creation_date":%d,"owner":{"display_name":"legacy"}}
			]}`, longBody, answerCreatedAt.Unix(), oldCreatedAt.Unix())
		case "/questions/123/comments":
			requireCollectionQuery(t, r)
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{"items":[
				{"body":"new comment","creation_date":%d,"owner":{"display_name":"bob"}},
				{"body":"old comment","creation_date":%d,"owner":{"display_name":"legacy"}}
			]}`, commentCreatedAt.Unix(), oldCreatedAt.Unix())
		default:
			t.Fatalf("got path %q", r.URL.Path)
		}
	}))
}

func requireQuestionQuery(t *testing.T, r *http.Request) {
	t.Helper()
	if r.URL.RawQuery != "site=stackoverflow" {
		t.Fatalf("got query %q", r.URL.RawQuery)
	}
}

func requireCollectionQuery(t *testing.T, r *http.Request) {
	t.Helper()
	if got := r.URL.Query().Get("site"); got != "stackoverflow" {
		t.Fatalf("got site query %q", got)
	}
	if got := r.URL.Query().Get("sort"); got != "creation" {
		t.Fatalf("got sort query %q", got)
	}
	if got := r.URL.Query().Get("order"); got != "desc" {
		t.Fatalf("got order query %q", got)
	}
	if got := r.URL.Query().Get("pagesize"); got != "5" {
		t.Fatalf("got pagesize query %q", got)
	}
	if got := r.URL.Query().Get("filter"); got != "withbody" {
		t.Fatalf("got filter query %q", got)
	}
}

func TestGetQuestionUpdateWithoutNewEvents(t *testing.T) {
	since := time.Date(2026, 3, 30, 17, 0, 0, 0, time.UTC)
	oldCreatedAt := since.Add(-time.Hour)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/questions/123":
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"items":[{"title":"How to test updates?"}]}`)
		case "/questions/123/answers":
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{"items":[{"body":"old answer","creation_date":%d,"owner":{"display_name":"legacy"}}]}`, oldCreatedAt.Unix())
		case "/questions/123/comments":
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprintf(w, `{"items":[{"body":"old comment","creation_date":%d,"owner":{"display_name":"legacy"}}]}`, oldCreatedAt.Unix())
		default:
			t.Fatalf("got path %q", r.URL.Path)
		}
	}))
	defer server.Close()

	client := NewStackOverflowClient()
	client.baseURL = server.URL
	client.client = server.Client()
	reqCtx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	update, err := client.GetQuestionUpdate(reqCtx, "123", since)
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

func TestGetQuestionUpdateErrors(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		body    string
		status  int
		wantErr string
	}{
		{name: "unexpected status", path: "/questions/123", status: http.StatusNotFound, wantErr: "unexpected status"},
		{name: "empty items", path: "/questions/123", status: http.StatusOK, body: `{"items":[]}`, wantErr: "question not found"},
	}
	reqCtx, cancel := context.WithTimeout(context.Background(), RequestTimeout)
	defer cancel()

	for _, tt := range tests {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == tt.path {
				w.WriteHeader(tt.status)
				if tt.body != "" {
					_, _ = w.Write([]byte(tt.body))
				}
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{"items":[]}`)
		}))

		client := NewStackOverflowClient()
		client.baseURL = server.URL
		client.client = server.Client()

		_, err := client.GetQuestionUpdate(reqCtx, "123", time.Time{})
		server.Close()
		if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
			t.Fatalf("case %q: got err %v, want containing %q", tt.name, err, tt.wantErr)
		}
	}
}

func TestParseStackOverflowURL(t *testing.T) {
	client := NewStackOverflowClient()
	id, err := client.ParseStackOverflowURL("https://stackoverflow.com/questions/123/test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "123" {
		t.Fatalf("got id %q", id)
	}

	if _, parseErr := client.ParseStackOverflowURL("https://google.com/questions/123/test"); parseErr == nil {
		t.Fatal("expected error for non-stackoverflow url")
	}
	if _, parseErr := client.ParseStackOverflowURL("https://stackoverflow.com/users/123"); parseErr == nil {
		t.Fatal("expected error for invalid stackoverflow path")
	}
}
