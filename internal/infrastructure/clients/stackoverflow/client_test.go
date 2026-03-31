package stackoverflow

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestGetQuestionUpdate(t *testing.T) {
	unixTS := time.Date(2026, 3, 30, 19, 0, 0, 0, time.UTC).Unix()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/questions/123" {
			t.Fatalf("got path %q", r.URL.Path)
		}
		if r.URL.RawQuery != "site=stackoverflow" {
			t.Fatalf("got query %q", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"items":[{"last_activity_date":%d}]}`, unixTS)
	}))
	defer server.Close()

	client := NewStackOverflowClient()
	client.baseURL = server.URL
	client.client = server.Client()

	update, err := client.GetQuestionUpdate("123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if update.LastUpdate.Unix() != unixTS {
		t.Fatalf("got unix %d, want %d", update.LastUpdate.Unix(), unixTS)
	}
}

func TestGetQuestionUpdateErrors(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		status  int
		wantErr string
	}{
		{name: "unexpected status", status: http.StatusNotFound, wantErr: "unexpected status"},
		{name: "empty items", status: http.StatusOK, body: `{"items":[]}`, wantErr: "question not found"},
	}

	for _, tt := range tests {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(tt.status)
			if tt.body != "" {
				_, _ = w.Write([]byte(tt.body))
			}
		}))

		client := NewStackOverflowClient()
		client.baseURL = server.URL
		client.client = server.Client()

		_, err := client.GetQuestionUpdate("123")
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
