package scrapper

import (
	"errors"
	"reflect"
	"testing"
	"time"

	models "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/models"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	ports "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

type noopLogger struct{}

func (noopLogger) Info(string, ...any)  {}
func (noopLogger) Warn(string, ...any)  {}
func (noopLogger) Error(string, ...any) {}

type stubRepo struct {
	trackOK   bool
	untrackOK bool
	present   map[int64]bool
	addChatOK bool
	deleteOK  bool
	listResp  []domain.Link
	allLinks  map[int64][]domain.Link
	updateErr error

	trackArgs struct {
		chatID int64
		url    string
		tags   []string
	}
	untrackArgs struct {
		chatID int64
		url    string
	}
	updated struct {
		chatID     int64
		url        string
		lastUpdate time.Time
	}
}

func (s *stubRepo) TrackLink(chatID int64, url string, tags []string) bool {
	s.trackArgs.chatID = chatID
	s.trackArgs.url = url
	s.trackArgs.tags = tags
	return s.trackOK
}

func (s *stubRepo) UnTrackLink(chatID int64, url string) bool {
	s.untrackArgs.chatID = chatID
	s.untrackArgs.url = url
	return s.untrackOK
}

func (s *stubRepo) ListLinks(chatID int64, tags []string) []domain.Link { return s.listResp }
func (s *stubRepo) ListAllLinks() map[int64][]domain.Link               { return s.allLinks }

func (s *stubRepo) UpdateLinksLastUpdate(chatID int64, url string, lastUpdate time.Time) error {
	s.updated.chatID = chatID
	s.updated.url = url
	s.updated.lastUpdate = lastUpdate
	return s.updateErr
}

func (s *stubRepo) IsPresent(chatID int64) bool  { return s.present[chatID] }
func (s *stubRepo) AddChat(chatID int64) bool    { return s.addChatOK }
func (s *stubRepo) DeleteChat(chatID int64) bool { return s.deleteOK }

type stubBotClient struct {
	err error
	got struct {
		chatIDs     []int64
		url         string
		description string
	}
}

func (s *stubBotClient) SendUpdates(chatIDS []int64, url, description string) error {
	s.got.chatIDs = chatIDS
	s.got.url = url
	s.got.description = description
	return s.err
}

type stubGithub struct {
	owner     string
	repo      string
	parseErr  error
	update    ports.ResourceUpdate
	updateErr error
	gotOwner  string
	gotRepo   string
}

func (s *stubGithub) GetRepoUpdate(owner, repo string) (ports.ResourceUpdate, error) {
	s.gotOwner = owner
	s.gotRepo = repo
	return s.update, s.updateErr
}

func (s *stubGithub) ParseGitHubURL(raw string) (owner, repo string, err error) {
	if s.parseErr != nil {
		return "", "", s.parseErr
	}
	return s.owner, s.repo, nil
}

type stubStackOverflow struct {
	questionID string
	parseErr   error
	update     ports.ResourceUpdate
	updateErr  error
	gotID      string
}

func (s *stubStackOverflow) GetQuestionUpdate(questionID string) (ports.ResourceUpdate, error) {
	s.gotID = questionID
	return s.update, s.updateErr
}

func (s *stubStackOverflow) ParseStackOverflowURL(raw string) (string, error) {
	if s.parseErr != nil {
		return "", s.parseErr
	}
	return s.questionID, nil
}

func TestScrapperCRUD(t *testing.T) {
	repo := &stubRepo{
		trackOK:   true,
		untrackOK: true,
		present:   map[int64]bool{1: true},
		addChatOK: true,
		deleteOK:  true,
		listResp:  []domain.Link{{URL: "https://github.com/user/repo"}},
	}
	svc := NewScrapper(noopLogger{}, repo, &stubBotClient{}, &stubGithub{}, &stubStackOverflow{})

	if err := svc.AddLink(models.AddLink{ChatID: 1, URL: "https://github.com/user/repo", Tags: []string{"go"}}); err != nil {
		t.Fatalf("unexpected add link error: %v", err)
	}
	if repo.trackArgs.chatID != 1 || repo.trackArgs.url != "https://github.com/user/repo" {
		t.Fatalf("got track args %+v", repo.trackArgs)
	}

	links, err := svc.GetLinks(1, []string{"go"})
	if err != nil {
		t.Fatalf("unexpected get links error: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("got links len %d", len(links))
	}

	if err := svc.DeleteLink(models.DeleteLink{ChatID: 1, URL: "https://github.com/user/repo"}); err != nil {
		t.Fatalf("unexpected delete link error: %v", err)
	}
	if repo.untrackArgs.url != "https://github.com/user/repo" {
		t.Fatalf("got untrack args %+v", repo.untrackArgs)
	}

	if err := svc.AddChat(2); err != nil {
		t.Fatalf("unexpected add chat error: %v", err)
	}
	if err := svc.DeleteChat(2); err != nil {
		t.Fatalf("unexpected delete chat error: %v", err)
	}
}

func TestScrapperErrors(t *testing.T) {
	service := NewScrapper(noopLogger{}, &stubRepo{present: map[int64]bool{}}, &stubBotClient{}, &stubGithub{}, &stubStackOverflow{})

	if err := service.AddLink(models.AddLink{ChatID: 1}); !errors.Is(err, ports.ErrChatNotFound) {
		t.Fatalf("got err %v, want %v", err, ports.ErrChatNotFound)
	}
	if err := service.DeleteLink(models.DeleteLink{ChatID: 1}); !errors.Is(err, ports.ErrChatNotFound) {
		t.Fatalf("got err %v, want %v", err, ports.ErrChatNotFound)
	}
	if _, err := service.GetLinks(1, nil); !errors.Is(err, ports.ErrChatNotFound) {
		t.Fatalf("got err %v, want %v", err, ports.ErrChatNotFound)
	}

	repo := &stubRepo{present: map[int64]bool{1: true}, trackOK: false, untrackOK: false, addChatOK: false, deleteOK: false}
	service = NewScrapper(noopLogger{}, repo, &stubBotClient{}, &stubGithub{}, &stubStackOverflow{})
	if err := service.AddLink(models.AddLink{ChatID: 1}); !errors.Is(err, ports.ErrLinkAlreadyExists) {
		t.Fatalf("got err %v, want %v", err, ports.ErrLinkAlreadyExists)
	}
	if err := service.DeleteLink(models.DeleteLink{ChatID: 1}); !errors.Is(err, ports.ErrLinkNotFound) {
		t.Fatalf("got err %v, want %v", err, ports.ErrLinkNotFound)
	}
	if err := service.AddChat(1); !errors.Is(err, ports.ErrChatAlreadyExists) {
		t.Fatalf("got err %v, want %v", err, ports.ErrChatAlreadyExists)
	}
	if err := service.DeleteChat(1); !errors.Is(err, ports.ErrChatNotFound) {
		t.Fatalf("got err %v, want %v", err, ports.ErrChatNotFound)
	}
}

func TestGetResourceUpdate(t *testing.T) {
	updatedAt := time.Date(2026, 3, 30, 20, 0, 0, 0, time.UTC)
	tests := []struct {
		name          string
		github        *stubGithub
		stackoverflow *stubStackOverflow
		wantDesc      string
		wantErr       error
	}{
		{name: "github update", github: &stubGithub{owner: "user", repo: "repo", update: ports.ResourceUpdate{LastUpdate: updatedAt}}, stackoverflow: &stubStackOverflow{parseErr: errors.New("nope")}, wantDesc: "Обнаружено обновление GitHub репозитория."},
		{name: "stackoverflow update", github: &stubGithub{parseErr: errors.New("nope")}, stackoverflow: &stubStackOverflow{questionID: "123", update: ports.ResourceUpdate{LastUpdate: updatedAt}}, wantDesc: "Обнаружено обновление вопроса StackOverflow."},
		{name: "unsupported link", github: &stubGithub{parseErr: errors.New("nope")}, stackoverflow: &stubStackOverflow{parseErr: errors.New("nope")}, wantErr: ports.ErrLinkNotFound},
	}

	for _, tt := range tests {
		svc := NewScrapper(noopLogger{}, &stubRepo{}, &stubBotClient{}, tt.github, tt.stackoverflow)
		update, desc, err := svc.getResourceUpdate("https://example.com")
		if tt.wantErr != nil {
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("case %q: got err %v, want %v", tt.name, err, tt.wantErr)
			}
			continue
		}
		if err != nil {
			t.Fatalf("case %q: unexpected error %v", tt.name, err)
		}
		if desc != tt.wantDesc {
			t.Fatalf("case %q: got desc %q, want %q", tt.name, desc, tt.wantDesc)
		}
		if !update.LastUpdate.Equal(updatedAt) {
			t.Fatalf("case %q: got update %v, want %v", tt.name, update.LastUpdate, updatedAt)
		}
	}
}

func TestCheckLinks(t *testing.T) {
	oldTime := time.Date(2026, 3, 30, 18, 0, 0, 0, time.UTC)
	newTime := oldTime.Add(time.Hour)
	repo := &stubRepo{
		allLinks: map[int64][]domain.Link{1: {{URL: "https://github.com/user/repo", LastUpdate: oldTime, Tags: map[string]struct{}{"go": {}}}}},
	}
	botClient := &stubBotClient{}
	github := &stubGithub{owner: "user", repo: "repo", update: ports.ResourceUpdate{LastUpdate: newTime}}
	stack := &stubStackOverflow{parseErr: errors.New("nope")}
	svc := NewScrapper(noopLogger{}, repo, botClient, github, stack)

	svc.CheckLinks()

	if !reflect.DeepEqual(botClient.got.chatIDs, []int64{1}) {
		t.Fatalf("got chat ids %v", botClient.got.chatIDs)
	}
	if botClient.got.url != "https://github.com/user/repo" {
		t.Fatalf("got url %q", botClient.got.url)
	}
	if botClient.got.description != "Обнаружено обновление GitHub репозитория." {
		t.Fatalf("got desc %q", botClient.got.description)
	}
	if repo.updated.chatID != 1 || repo.updated.url != "https://github.com/user/repo" || !repo.updated.lastUpdate.Equal(newTime) {
		t.Fatalf("got update args %+v", repo.updated)
	}
}

func TestCheckLinksSkipsWhenNothingChanged(t *testing.T) {
	current := time.Date(2026, 3, 30, 18, 0, 0, 0, time.UTC)
	repo := &stubRepo{
		allLinks: map[int64][]domain.Link{1: {{URL: "https://github.com/user/repo", LastUpdate: current}}},
	}
	botClient := &stubBotClient{}
	github := &stubGithub{owner: "user", repo: "repo", update: ports.ResourceUpdate{LastUpdate: current}}
	svc := NewScrapper(noopLogger{}, repo, botClient, github, &stubStackOverflow{parseErr: errors.New("nope")})

	svc.CheckLinks()

	if botClient.got.url != "" {
		t.Fatalf("expected no notification, got %+v", botClient.got)
	}
	if repo.updated.url != "" {
		t.Fatalf("expected no repo update, got %+v", repo.updated)
	}
}
