package scrapper

import (
	"testing"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/models"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
)

func TestToAddLink(t *testing.T) {
	req := AddLinkRequest{URL: "https://github.com/user/repo", Tags: []string{"go", "backend"}}
	got := ToAddLink(req, 42)
	want := models.AddLink{ChatID: 42, URL: "https://github.com/user/repo", Tags: []string{"go", "backend"}}

	if got.ChatID != want.ChatID || got.URL != want.URL || len(got.Tags) != len(want.Tags) || got.Tags[0] != want.Tags[0] || got.Tags[1] != want.Tags[1] {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestToDeleteLink(t *testing.T) {
	req := DeleteLinkRequest{URL: "https://github.com/user/repo"}
	got := ToDeleteLink(req, 13)
	want := models.DeleteLink{ChatID: 13, URL: "https://github.com/user/repo"}

	if got != want {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestToGetLinkResponse(t *testing.T) {
	updatedAt := time.Date(2026, 3, 30, 12, 30, 0, 0, time.UTC)
	link := domain.Link{
		URL:        "https://github.com/user/repo",
		LastUpdate: updatedAt,
		Tags:       map[string]struct{}{"go": {}, "backend": {}},
	}

	got := ToGetLinkResponse(link)
	if got.URL != link.URL {
		t.Fatalf("got url %q, want %q", got.URL, link.URL)
	}
	if !got.LastUpdate.Equal(updatedAt) {
		t.Fatalf("got last update %v, want %v", got.LastUpdate, updatedAt)
	}
	if len(got.Tags) != 2 {
		t.Fatalf("got tags len %d", len(got.Tags))
	}

	tagSet := map[string]struct{}{}
	for _, tag := range got.Tags {
		tagSet[tag] = struct{}{}
	}
	if _, ok := tagSet["go"]; !ok {
		t.Fatalf("expected go tag in %v", got.Tags)
	}
	if _, ok := tagSet["backend"]; !ok {
		t.Fatalf("expected backend tag in %v", got.Tags)
	}
}
