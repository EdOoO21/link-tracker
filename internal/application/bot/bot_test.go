package bot

import (
	"testing"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
)

func TestComputeGotLink(t *testing.T) {
	chatID := int64(1)
	tests := []struct {
		name          string
		input         string
		linkExists    bool
		wantText      string
		wantCommand   string
		wantState     State
		wantStoredURL string
	}{
		{name: "invalid url", input: "tbank://github.com/user/repo", wantText: InvalidURL, wantCommand: TextInvalidLinkGot, wantState: TrackCommandGot},
		{name: "unsupported domain", input: "https://google.com", wantText: NotSupportedURL, wantCommand: TextNotSupportedLinkGot, wantState: TrackCommandGot},
		{name: "already tracked", input: "https://github.com/user/repo", linkExists: true, wantText: TrackExistedURLNoReset, wantCommand: URLExists, wantState: TrackCommandGot},
		{name: "valid github url", input: "https://github.com/user/repo", wantText: ValidURL, wantCommand: TextValidLinkGot, wantState: LinkGot, wantStoredURL: "https://github.com/user/repo"},
	}

	for _, tt := range tests {
		app := &App{
			repo:      &mockRepo{linkExists: tt.linkExists},
			stMachine: StateMachine{chatID: {State: TrackCommandGot}},
		}
		msg, command := app.computeGotLink(chatID, tt.input)

		if msg.Text != tt.wantText {
			t.Fatalf("case %q: got text %q, want %q", tt.name, msg.Text, tt.wantText)
		}
		if command != tt.wantCommand {
			t.Fatalf("case %q: got command %q, want %q", tt.name, command, tt.wantCommand)
		}
		if app.stMachine[chatID].State != tt.wantState {
			t.Fatalf("case %q: got state %v, want %v", tt.name, app.stMachine[chatID].State, tt.wantState)
		}
		if app.stMachine[chatID].URL != tt.wantStoredURL {
			t.Fatalf("case %q: got stored url %q, want %q", tt.name, app.stMachine[chatID].URL, tt.wantStoredURL)
		}
	}
}

func TestComputeTrack(t *testing.T) {
	chatID := int64(7)
	update := tgbotapi.Update{Message: &tgbotapi.Message{Chat: &tgbotapi.Chat{ID: chatID}}}

	app := &App{stMachine: make(StateMachine)}
	msg := app.computeTrack(update)
	if msg.Text != AddingToTrackSeqStarted {
		t.Fatalf("got text %q, want %q", msg.Text, AddingToTrackSeqStarted)
	}
	if app.stMachine[chatID] == nil || app.stMachine[chatID].State != TrackCommandGot {
		t.Fatalf("expected state %v, got %+v", TrackCommandGot, app.stMachine[chatID])
	}

	app.stMachine[chatID].URL = "https://old"
	msg = app.computeTrack(update)
	if msg.Text != AddingToTrackSeqRestarted {
		t.Fatalf("got text %q, want %q", msg.Text, AddingToTrackSeqRestarted)
	}
	if app.stMachine[chatID].URL != "" {
		t.Fatalf("expected url reset, got %q", app.stMachine[chatID].URL)
	}
}

func TestComputeHelpAndUnknownCommandClearState(t *testing.T) {
	chatID := int64(5)
	app := &App{stMachine: StateMachine{chatID: {State: LinkGot, URL: "https://github.com/user/repo"}}}

	help := app.computeHelp(chatID)
	if help.Text != HelpCommand {
		t.Fatalf("got help text %q", help.Text)
	}
	if _, ok := app.stMachine[chatID]; ok {
		t.Fatal("expected state to be deleted after help")
	}

	app.stMachine[chatID] = &StateInfo{State: TrackCommandGot, URL: "https://github.com/user/repo"}
	msg, command := app.computeUnknownCommand(chatID, "super-long-command-name-that-should-be-truncated")
	if msg.Text != UnknownCommand {
		t.Fatalf("got unknown text %q", msg.Text)
	}
	if command != Unknown {
		t.Fatalf("got command %q, want %q", command, Unknown)
	}
	if _, ok := app.stMachine[chatID]; ok {
		t.Fatal("expected state to be deleted after unknown command")
	}
}

func TestComputeCancel(t *testing.T) {
	chatID := int64(3)
	app := &App{stMachine: StateMachine{chatID: {State: LinkGot, URL: "https://github.com/user/repo"}}}

	msg := app.computeCancel(chatID)
	if msg.Text != CancelCommand {
		t.Fatalf("got text %q, want %q", msg.Text, CancelCommand)
	}
	if _, ok := app.stMachine[chatID]; ok {
		t.Fatal("expected state to be deleted")
	}
}

func TestLinkUpdated(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        string
	}{
		{name: "with description", description: "repo changed", want: "Ссылка была обновлена!\n\nhttps://github.com/user/repo\n\nОписание: \n\nrepo changed"},
		{name: "without description", description: "", want: "Ссылка была обновлена!\n\nhttps://github.com/user/repo\n\nОписание: \n\nотсутствует"},
	}

	for _, tt := range tests {
		got := linkUpdated("https://github.com/user/repo", tt.description)
		if got != tt.want {
			t.Fatalf("case %q: got %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestLinksOutput(t *testing.T) {
	links := []domain.Link{
		{URL: "https://github.com/user/repo1", LastUpdate: time.Time{}},
		{URL: "https://github.com/user/repo2", LastUpdate: time.Time{}},
	}

	got := linksOutput(links)
	want := "Ссылки:\n\n1. https://github.com/user/repo1\n2. https://github.com/user/repo2\n"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestComputeTextWithoutStateReturnsUnknownText(t *testing.T) {
	chatID := int64(11)
	app := &App{stMachine: make(StateMachine)}

	msg, command := app.computeText(chatID, "hello")
	if msg.Text != UnknownText {
		t.Fatalf("got text %q, want %q", msg.Text, UnknownText)
	}
	if command != NotCommand {
		t.Fatalf("got command %q, want %q", command, NotCommand)
	}
}
