package scrapper

import (
	"testing"
	"time"
)

func TestToDomainLinks(t *testing.T) {
	updatedAt := time.Date(2026, 3, 30, 10, 0, 0, 0, time.UTC)
	resp := []LinkResponse{{
		URL:        "https://github.com/user/repo",
		Tags:       []string{"go", "backend", "go"},
		LastUpdate: updatedAt,
	}}

	got := ToDomainLinks(resp)
	if len(got) != 1 {
		t.Fatalf("got len %d", len(got))
	}
	if got[0].URL != "https://github.com/user/repo" {
		t.Fatalf("got url %q", got[0].URL)
	}
	if !got[0].LastUpdate.Equal(updatedAt) {
		t.Fatalf("got last update %v, want %v", got[0].LastUpdate, updatedAt)
	}
	if len(got[0].Tags) != 2 {
		t.Fatalf("got tags len %d", len(got[0].Tags))
	}
	if _, ok := got[0].Tags["go"]; !ok {
		t.Fatalf("expected go tag in %v", got[0].Tags)
	}
	if _, ok := got[0].Tags["backend"]; !ok {
		t.Fatalf("expected backend tag in %v", got[0].Tags)
	}
}
