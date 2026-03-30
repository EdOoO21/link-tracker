package repository

import (
	"errors"
	"testing"
	"time"

	ports "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

func TestUsersLinksChatLifecycle(t *testing.T) {
	repo := NewRepo()

	if !repo.AddChat(1) {
		t.Fatal("expected add chat to succeed")
	}
	if repo.AddChat(1) {
		t.Fatal("expected duplicate add chat to fail")
	}
	if !repo.IsPresent(1) {
		t.Fatal("expected chat to be present")
	}
	if !repo.DeleteChat(1) {
		t.Fatal("expected delete chat to succeed")
	}
	if repo.DeleteChat(1) {
		t.Fatal("expected second delete chat to fail")
	}
}

func TestUsersLinksTrackAndList(t *testing.T) {
	repo := NewRepo()
	repo.AddChat(1)

	if !repo.TrackLink(1, "https://github.com/user/repo", []string{"go", "backend"}) {
		t.Fatal("expected track link to succeed")
	}
	if repo.TrackLink(1, "https://github.com/user/repo", []string{"go"}) {
		t.Fatal("expected duplicate track link to fail")
	}

	all := repo.ListLinks(1, nil)
	if len(all) != 1 {
		t.Fatalf("got links len %d", len(all))
	}
	filtered := repo.ListLinks(1, []string{"go"})
	if len(filtered) != 1 {
		t.Fatalf("got filtered len %d", len(filtered))
	}
	missing := repo.ListLinks(1, []string{"missing"})
	if len(missing) != 0 {
		t.Fatalf("got missing-tag len %d", len(missing))
	}
	if _, ok := all[0].Tags["backend"]; !ok {
		t.Fatalf("expected backend tag in %v", all[0].Tags)
	}
}

func TestUsersLinksUntrack(t *testing.T) {
	repo := NewRepo()
	repo.AddChat(1)
	repo.TrackLink(1, "https://github.com/user/repo", []string{"go"})

	if !repo.UnTrackLink(1, "https://github.com/user/repo") {
		t.Fatal("expected untrack to succeed")
	}
	if repo.UnTrackLink(1, "https://github.com/user/repo") {
		t.Fatal("expected second untrack to fail")
	}
}

func TestUsersLinksListAllLinksReturnsMapCopy(t *testing.T) {
	repo := NewRepo()
	repo.AddChat(1)
	repo.TrackLink(1, "https://github.com/user/repo", []string{"go"})

	copyMap := repo.ListAllLinks()
	delete(copyMap, 1)

	if !repo.IsPresent(1) {
		t.Fatal("expected original repo to keep chat after deleting from copy")
	}
}

func TestUsersLinksUpdateLastUpdate(t *testing.T) {
	repo := NewRepo()
	repo.AddChat(1)
	repo.TrackLink(1, "https://github.com/user/repo", []string{"go"})
	updatedAt := time.Date(2026, 3, 30, 18, 0, 0, 0, time.UTC)

	if err := repo.UpdateLinksLastUpdate(1, "https://github.com/user/repo", updatedAt); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	links := repo.ListLinks(1, nil)
	if len(links) != 1 {
		t.Fatalf("got links len %d", len(links))
	}
	if !links[0].LastUpdate.Equal(updatedAt) {
		t.Fatalf("got last update %v, want %v", links[0].LastUpdate, updatedAt)
	}

	err := repo.UpdateLinksLastUpdate(1, "https://github.com/user/missing", updatedAt)
	if !errors.Is(err, ports.ErrLinkNotFound) {
		t.Fatalf("got err %v, want %v", err, ports.ErrLinkNotFound)
	}
}
