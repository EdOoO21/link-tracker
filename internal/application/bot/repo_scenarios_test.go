package bot

import (
	"context"
	"errors"
	"reflect"
	"testing"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	ports "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

type mockRepo struct {
	trackLinkArgs struct {
		chatID int64
		url    string
		tags   []string
	}
	untrackArgs struct {
		chatID int64
		url    string
	}
	listArgs struct {
		chatID int64
		tags   []string
	}
	addChatID int64

	trackErr   error
	untrackErr error
	listErr    error
	addChatErr error
	listResp   []domain.Link
	linkExists bool
}

func (m *mockRepo) TrackLink(_ context.Context, chatID int64, url string, tags []string) error {
	m.trackLinkArgs.chatID = chatID
	m.trackLinkArgs.url = url
	m.trackLinkArgs.tags = tags
	return m.trackErr
}

func (m *mockRepo) UnTrackLink(_ context.Context, chatID int64, url string) error {
	m.untrackArgs.chatID = chatID
	m.untrackArgs.url = url
	return m.untrackErr
}

func (m *mockRepo) ListLinks(_ context.Context, chatID int64, tags []string) ([]domain.Link, error) {
	m.listArgs.chatID = chatID
	m.listArgs.tags = tags
	return m.listResp, m.listErr
}

func (m *mockRepo) AddChat(_ context.Context, chatID int64) error {
	m.addChatID = chatID
	return m.addChatErr
}

func (m *mockRepo) DeleteChat(_ context.Context, _ int64) error { return nil }

func (m *mockRepo) IsLinkPresent(_ context.Context, _ int64, _ string) bool {
	return m.linkExists
}

func commandUpdate(chatID int64, command, args string) tgbotapi.Update {
	text := command
	if args != "" {
		text += " " + args
	}
	return tgbotapi.Update{
		Message: &tgbotapi.Message{
			Text: text,
			Chat: &tgbotapi.Chat{ID: chatID},
			Entities: []tgbotapi.MessageEntity{{
				Type:   "bot_command",
				Offset: 0,
				Length: len(command),
			}},
		},
	}
}

func TestComputeStart(t *testing.T) {
	tests := []struct {
		name     string
		addErr   error
		wantText string
		wantChat int64
	}{
		{name: "success", wantText: StartCommand, wantChat: 10},
		{name: "already exists", addErr: ports.ErrChatAlreadyExists, wantText: StartCommand, wantChat: 10},
		{name: "unexpected error", addErr: errors.New("boom"), wantText: StartCommandFailedChatAdd, wantChat: 10},
	}

	for _, tt := range tests {
		repo := &mockRepo{addChatErr: tt.addErr}
		app := &App{repo: repo, stMachine: StateMachine{10: {State: LinkGot, URL: "https://github.com/user/repo"}}}
		msg := app.computeStart(context.Background(), 10)
		if msg.Text != tt.wantText {
			t.Fatalf("case %q: got text %q, want %q", tt.name, msg.Text, tt.wantText)
		}
		if repo.addChatID != tt.wantChat {
			t.Fatalf("case %q: got chat id %d, want %d", tt.name, repo.addChatID, tt.wantChat)
		}
		if _, ok := app.stMachine[10]; ok {
			t.Fatal("expected state to be cleared on start")
		}
	}
}

func TestComputeList(t *testing.T) {
	chatID := int64(11)
	tests := []struct {
		name     string
		args     string
		listResp []domain.Link
		listErr  error
		wantText string
		wantTags []string
	}{
		{name: "chat not found", listErr: ports.ErrChatNotFound, wantText: ChatIDNotFound},
		{name: "unexpected error", listErr: errors.New("boom"), wantText: LinksListFailed},
		{name: "empty list", wantText: LinksNotExist},
		{name: "filtered list", args: "go backend", listResp: []domain.Link{{URL: "https://github.com/user/repo"}}, wantText: "Ссылки:\n\n1. https://github.com/user/repo\n", wantTags: []string{"go", "backend"}},
	}

	for _, tt := range tests {
		repo := &mockRepo{listResp: tt.listResp, listErr: tt.listErr}
		app := &App{repo: repo, stMachine: StateMachine{chatID: {State: TrackCommandGot}}}
		update := commandUpdate(chatID, "/list", tt.args)
		msg := app.computeList(context.Background(), update)

		if msg.Text != tt.wantText {
			t.Fatalf("case %q: got text %q, want %q", tt.name, msg.Text, tt.wantText)
		}
		if repo.listArgs.chatID != chatID {
			t.Fatalf("case %q: got chat id %d", tt.name, repo.listArgs.chatID)
		}
		if tt.wantTags != nil && !reflect.DeepEqual(repo.listArgs.tags, tt.wantTags) {
			t.Fatalf("case %q: got tags %v, want %v", tt.name, repo.listArgs.tags, tt.wantTags)
		}
		if _, ok := app.stMachine[chatID]; ok {
			t.Fatal("expected state to be cleared on list")
		}
	}
}

func TestComputeUnTrack(t *testing.T) {
	chatID := int64(17)
	tests := []struct {
		name       string
		args       string
		untrackErr error
		wantText   string
	}{
		{name: "invalid args", args: "", wantText: UntrackNotOneURL},
		{name: "link not found", args: "https://github.com/user/repo", untrackErr: ports.ErrLinkNotFound, wantText: UntrackNotExistedURL},
		{name: "chat not found", args: "https://github.com/user/repo", untrackErr: ports.ErrChatNotFound, wantText: ChatIDNotFound},
		{name: "unexpected error", args: "https://github.com/user/repo", untrackErr: errors.New("boom"), wantText: UntrackErrToDeleteLink},
		{name: "success", args: "https://github.com/user/repo", wantText: UntrackExistedURL},
	}

	for _, tt := range tests {
		repo := &mockRepo{untrackErr: tt.untrackErr}
		app := &App{repo: repo, stMachine: StateMachine{chatID: {State: LinkGot}}}
		update := commandUpdate(chatID, "/untrack", tt.args)
		msg := app.computeUnTrack(context.Background(), update)

		if msg.Text != tt.wantText {
			t.Fatalf("case %q: got text %q, want %q", tt.name, msg.Text, tt.wantText)
		}
		if tt.args != "" && repo.untrackArgs.chatID != chatID {
			t.Fatalf("case %q: got chat id %d", tt.name, repo.untrackArgs.chatID)
		}
		if _, ok := app.stMachine[chatID]; ok {
			t.Fatal("expected state to be cleared on untrack")
		}
	}
}

func TestComputeGotTags(t *testing.T) {
	chatID := int64(21)
	tests := []struct {
		name     string
		trackErr error
		wantText string
		wantCmd  string
		wantTags []string
	}{
		{name: "already exists", trackErr: ports.ErrLinkAlreadyExists, wantText: TrackExistedURLWithReset, wantCmd: URLExists, wantTags: []string{"go", "backend"}},
		{name: "chat not found", trackErr: ports.ErrChatNotFound, wantText: ChatIDNotFound, wantCmd: ChatIDNotFound, wantTags: []string{"go", "backend"}},
		{name: "unexpected error", trackErr: errors.New("boom"), wantText: TrackErrorToAddLink, wantCmd: UnexpectedErrorToAddLink, wantTags: []string{"go", "backend"}},
		{name: "success", wantText: TrackNotExistedURL, wantCmd: URLAdded, wantTags: []string{"go", "backend"}},
	}

	for _, tt := range tests {
		repo := &mockRepo{trackErr: tt.trackErr}
		app := &App{repo: repo, stMachine: StateMachine{chatID: {State: LinkGot, URL: "https://github.com/user/repo"}}}
		msg, command := app.computeGotTags(context.Background(), chatID, "go, backend")

		if msg.Text != tt.wantText {
			t.Fatalf("case %q: got text %q, want %q", tt.name, msg.Text, tt.wantText)
		}
		if command != tt.wantCmd {
			t.Fatalf("case %q: got command %q, want %q", tt.name, command, tt.wantCmd)
		}
		if repo.trackLinkArgs.chatID != chatID || repo.trackLinkArgs.url != "https://github.com/user/repo" {
			t.Fatalf("case %q: got track args %+v", tt.name, repo.trackLinkArgs)
		}
		if !reflect.DeepEqual(repo.trackLinkArgs.tags, tt.wantTags) {
			t.Fatalf("case %q: got tags %v, want %v", tt.name, repo.trackLinkArgs.tags, tt.wantTags)
		}
		if _, ok := app.stMachine[chatID]; ok {
			t.Fatal("expected state to be cleared after got tags")
		}
	}
}

func TestComputeGotTagsWithoutTags(t *testing.T) {
	chatID := int64(22)
	repo := &mockRepo{}
	app := &App{repo: repo, stMachine: StateMachine{chatID: {State: LinkGot, URL: "https://github.com/user/repo"}}}

	msg, command := app.computeGotTags(context.Background(), chatID, "-")

	if msg.Text != TrackNotExistedURL {
		t.Fatalf("got text %q, want %q", msg.Text, TrackNotExistedURL)
	}
	if command != URLAdded {
		t.Fatalf("got command %q, want %q", command, URLAdded)
	}
	if len(repo.trackLinkArgs.tags) != 0 {
		t.Fatalf("got tags %v, want empty slice", repo.trackLinkArgs.tags)
	}
}
