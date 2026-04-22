package scrapper

import (
	"context"
	"errors"
	"slices"
	"sync"
	"testing"
	"time"

	scrapperinterfaces "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/interfaces"
	models "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/models"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	ports "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

type noopLogger struct{}

func (noopLogger) Info(string, ...any)  {}
func (noopLogger) Warn(string, ...any)  {}
func (noopLogger) Error(string, ...any) {}

type stubRepo struct {
	mu sync.Mutex

	trackErr   error
	untrackErr error
	addTagErr  error
	getTagsErr error
	delTagErr  error
	present    map[int64]bool
	presentErr error
	addChatErr error
	deleteErr  error
	listResp   []domain.Link
	listErr    error
	tagsResp   []string
	allLinks   []models.TrackedLink
	listAllErr error
	updateErr  error

	trackArgs struct {
		chatID int64
		url    string
		tags   []string
	}
	untrackArgs struct {
		chatID int64
		url    string
	}
	tagArgs struct {
		chatID int64
		url    string
		tag    string
	}
	getTagsArgs struct {
		chatID int64
		url    string
	}
	updated struct {
		linkID     int64
		lastUpdate time.Time
	}
}

func (s *stubRepo) TrackLink(_ context.Context, chatID int64, url string, tags []string) error {
	s.trackArgs.chatID = chatID
	s.trackArgs.url = url
	s.trackArgs.tags = tags
	return s.trackErr
}

func (s *stubRepo) UnTrackLink(_ context.Context, chatID int64, url string) error {
	s.untrackArgs.chatID = chatID
	s.untrackArgs.url = url
	return s.untrackErr
}

func (s *stubRepo) ListLinks(_ context.Context, _ int64, _ []string) ([]domain.Link, error) {
	return s.listResp, s.listErr
}

func (s *stubRepo) AddTag(_ context.Context, chatID int64, url, tag string) error {
	s.tagArgs.chatID = chatID
	s.tagArgs.url = url
	s.tagArgs.tag = tag
	return s.addTagErr
}

func (s *stubRepo) DeleteTag(_ context.Context, chatID int64, url, tag string) error {
	s.tagArgs.chatID = chatID
	s.tagArgs.url = url
	s.tagArgs.tag = tag
	return s.delTagErr
}

func (s *stubRepo) GetTags(_ context.Context, chatID int64, url string) ([]string, error) {
	s.getTagsArgs.chatID = chatID
	s.getTagsArgs.url = url
	return s.tagsResp, s.getTagsErr
}

func (s *stubRepo) ListAllLinksBatch(_ context.Context, afterLinkID int64, limit int) ([]models.TrackedLink, error) {
	if s.listAllErr != nil {
		return nil, s.listAllErr
	}
	if limit <= 0 {
		return []models.TrackedLink{}, nil
	}

	batch := make([]models.TrackedLink, 0, limit)
	for _, link := range s.allLinks {
		if link.LinkID <= afterLinkID {
			continue
		}
		batch = append(batch, link)
		if len(batch) == limit {
			break
		}
	}

	return batch, nil
}

func (s *stubRepo) UpdateLinksLastUpdate(_ context.Context, linkID int64, lastUpdate time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.updated.linkID = linkID
	s.updated.lastUpdate = lastUpdate
	return s.updateErr
}

func (s *stubRepo) IsPresent(_ context.Context, chatID int64) (bool, error) {
	return s.present[chatID], s.presentErr
}

func (s *stubRepo) AddChat(_ context.Context, _ int64) error { return s.addChatErr }
func (s *stubRepo) DeleteChat(_ context.Context, _ int64) error {
	return s.deleteErr
}

type stubBotClient struct {
	mu  sync.Mutex
	err error
	got struct {
		chatIDs     []int64
		url         string
		description string
	}
	failedReport struct {
		chatID int64
		urls   []string
	}
	failedReports []struct {
		chatID int64
		urls   []string
	}
}

func (s *stubBotClient) SendUpdates(_ context.Context, chatIDs []int64, url, description string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.got.chatIDs = chatIDs
	s.got.url = url
	s.got.description = description
	return s.err
}

func (s *stubBotClient) SendFailedLinksReport(_ context.Context, chatID int64, urls []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failedReport.chatID = chatID
	s.failedReport.urls = append([]string(nil), urls...)
	s.failedReports = append(s.failedReports, struct {
		chatID int64
		urls   []string
	}{
		chatID: chatID,
		urls:   append([]string(nil), urls...),
	})
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
	gotSince  time.Time
}

func (s *stubGithub) GetRepoUpdate(_ context.Context, owner, repo string, since time.Time) (ports.ResourceUpdate, error) {
	s.gotOwner = owner
	s.gotRepo = repo
	s.gotSince = since
	return s.update, s.updateErr
}

func (s *stubGithub) CanHandle(_ string) bool {
	return s.parseErr == nil
}

func (s *stubGithub) CheckUpdate(ctx context.Context, raw string, since time.Time) (ports.ResourceUpdate, error) {
	owner, repo, err := s.ParseGitHubURL(raw)
	if err != nil {
		return ports.ResourceUpdate{}, err
	}

	return s.GetRepoUpdate(ctx, owner, repo, since)
}

func (s *stubGithub) ParseGitHubURL(_ string) (owner, repo string, err error) {
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
	gotSince   time.Time
}

func (s *stubStackOverflow) GetQuestionUpdate(_ context.Context, questionID string, since time.Time) (ports.ResourceUpdate, error) {
	s.gotID = questionID
	s.gotSince = since
	return s.update, s.updateErr
}

func (s *stubStackOverflow) CanHandle(_ string) bool {
	return s.parseErr == nil
}

func (s *stubStackOverflow) CheckUpdate(ctx context.Context, raw string, since time.Time) (ports.ResourceUpdate, error) {
	questionID, err := s.ParseStackOverflowURL(raw)
	if err != nil {
		return ports.ResourceUpdate{}, err
	}

	return s.GetQuestionUpdate(ctx, questionID, since)
}

func (s *stubStackOverflow) ParseStackOverflowURL(_ string) (string, error) {
	if s.parseErr != nil {
		return "", s.parseErr
	}
	return s.questionID, nil
}

func TestScrapperCRUD(t *testing.T) {
	repo := &stubRepo{
		listResp: []domain.Link{{URL: "https://github.com/user/repo"}},
	}
	svc := NewScrapper(noopLogger{}, repo, &stubBotClient{}, []scrapperinterfaces.Checker{&stubGithub{}, &stubStackOverflow{}}, 100, 4)
	ctx := context.Background()
	if err := svc.AddLink(ctx, models.AddLink{ChatID: 1, URL: "https://github.com/user/repo", Tags: []string{"go"}}); err != nil {
		t.Fatalf("unexpected add link error: %v", err)
	}
	if repo.trackArgs.chatID != 1 || repo.trackArgs.url != "https://github.com/user/repo" {
		t.Fatalf("got track args %+v", repo.trackArgs)
	}

	links, err := svc.GetLinks(ctx, 1, []string{"go"})
	if err != nil {
		t.Fatalf("unexpected get links error: %v", err)
	}
	if len(links) != 1 {
		t.Fatalf("got links len %d", len(links))
	}

	if deleteErr := svc.DeleteLink(ctx, models.DeleteLink{ChatID: 1, URL: "https://github.com/user/repo"}); deleteErr != nil {
		t.Fatalf("unexpected delete link error: %v", deleteErr)
	}
	if repo.untrackArgs.url != "https://github.com/user/repo" {
		t.Fatalf("got untrack args %+v", repo.untrackArgs)
	}

	if addTagErr := svc.AddTag(ctx, models.AddTag{ChatID: 1, URL: "https://github.com/user/repo", Tag: "backend"}); addTagErr != nil {
		t.Fatalf("unexpected add tag error: %v", addTagErr)
	}
	if repo.tagArgs.chatID != 1 || repo.tagArgs.url != "https://github.com/user/repo" || repo.tagArgs.tag != "backend" {
		t.Fatalf("got tag args %+v", repo.tagArgs)
	}

	repo.tagsResp = []string{"backend", "go"}
	tags, tagsErr := svc.GetTags(ctx, 1, "https://github.com/user/repo")
	if tagsErr != nil {
		t.Fatalf("unexpected get tags error: %v", tagsErr)
	}
	if !slices.Equal(tags, []string{"backend", "go"}) {
		t.Fatalf("got tags %v", tags)
	}

	if deleteTagErr := svc.DeleteTag(ctx, models.DeleteTag{ChatID: 1, URL: "https://github.com/user/repo", Tag: "backend"}); deleteTagErr != nil {
		t.Fatalf("unexpected delete tag error: %v", deleteTagErr)
	}

	if addChatErr := svc.AddChat(ctx, 2); addChatErr != nil {
		t.Fatalf("unexpected add chat error: %v", addChatErr)
	}
	if deleteChatErr := svc.DeleteChat(ctx, 2); deleteChatErr != nil {
		t.Fatalf("unexpected delete chat error: %v", deleteChatErr)
	}
}

func TestScrapperErrors(t *testing.T) {
	service := NewScrapper(noopLogger{}, &stubRepo{
		trackErr:   ports.ErrChatNotFound,
		untrackErr: ports.ErrChatNotFound,
		addTagErr:  ports.ErrChatNotFound,
		getTagsErr: ports.ErrChatNotFound,
		delTagErr:  ports.ErrChatNotFound,
		listErr:    ports.ErrChatNotFound,
	}, &stubBotClient{}, []scrapperinterfaces.Checker{&stubGithub{}, &stubStackOverflow{}}, 100, 4)
	ctx := context.Background()
	if err := service.AddLink(ctx, models.AddLink{ChatID: 1}); !errors.Is(err, ports.ErrChatNotFound) {
		t.Fatalf("got err %v, want %v", err, ports.ErrChatNotFound)
	}
	if err := service.DeleteLink(ctx, models.DeleteLink{ChatID: 1}); !errors.Is(err, ports.ErrChatNotFound) {
		t.Fatalf("got err %v, want %v", err, ports.ErrChatNotFound)
	}
	if _, err := service.GetLinks(ctx, 1, nil); !errors.Is(err, ports.ErrChatNotFound) {
		t.Fatalf("got err %v, want %v", err, ports.ErrChatNotFound)
	}
	if err := service.AddTag(ctx, models.AddTag{ChatID: 1}); !errors.Is(err, ports.ErrChatNotFound) {
		t.Fatalf("got err %v, want %v", err, ports.ErrChatNotFound)
	}
	if _, err := service.GetTags(ctx, 1, "https://github.com/user/repo"); !errors.Is(err, ports.ErrChatNotFound) {
		t.Fatalf("got err %v, want %v", err, ports.ErrChatNotFound)
	}
	if err := service.DeleteTag(ctx, models.DeleteTag{ChatID: 1}); !errors.Is(err, ports.ErrChatNotFound) {
		t.Fatalf("got err %v, want %v", err, ports.ErrChatNotFound)
	}

	repo := &stubRepo{
		trackErr:   ports.ErrLinkAlreadyExists,
		untrackErr: ports.ErrLinkNotFound,
		addTagErr:  ports.ErrTagAlreadyExists,
		getTagsErr: ports.ErrLinkNotFound,
		delTagErr:  ports.ErrTagNotFound,
		addChatErr: ports.ErrChatAlreadyExists,
		deleteErr:  ports.ErrChatNotFound,
	}
	service = NewScrapper(noopLogger{}, repo, &stubBotClient{}, []scrapperinterfaces.Checker{&stubGithub{}, &stubStackOverflow{}}, 100, 4)
	if err := service.AddLink(ctx, models.AddLink{ChatID: 1}); !errors.Is(err, ports.ErrLinkAlreadyExists) {
		t.Fatalf("got err %v, want %v", err, ports.ErrLinkAlreadyExists)
	}
	if err := service.DeleteLink(ctx, models.DeleteLink{ChatID: 1}); !errors.Is(err, ports.ErrLinkNotFound) {
		t.Fatalf("got err %v, want %v", err, ports.ErrLinkNotFound)
	}
	if err := service.AddTag(ctx, models.AddTag{ChatID: 1}); !errors.Is(err, ports.ErrTagAlreadyExists) {
		t.Fatalf("got err %v, want %v", err, ports.ErrTagAlreadyExists)
	}
	if _, err := service.GetTags(ctx, 1, "https://github.com/user/repo"); !errors.Is(err, ports.ErrLinkNotFound) {
		t.Fatalf("got err %v, want %v", err, ports.ErrLinkNotFound)
	}
	if err := service.DeleteTag(ctx, models.DeleteTag{ChatID: 1}); !errors.Is(err, ports.ErrTagNotFound) {
		t.Fatalf("got err %v, want %v", err, ports.ErrTagNotFound)
	}
	if err := service.AddChat(ctx, 1); !errors.Is(err, ports.ErrChatAlreadyExists) {
		t.Fatalf("got err %v, want %v", err, ports.ErrChatAlreadyExists)
	}
	if err := service.DeleteChat(ctx, 1); !errors.Is(err, ports.ErrChatNotFound) {
		t.Fatalf("got err %v, want %v", err, ports.ErrChatNotFound)
	}
}

func TestGetResourceUpdate(t *testing.T) {
	updatedAt := time.Date(2026, 3, 30, 20, 0, 0, 0, time.UTC)
	tests := []struct {
		name          string
		github        *stubGithub
		stackoverflow *stubStackOverflow
		wantErr       error
	}{
		{name: "github update", github: &stubGithub{owner: "user", repo: "repo", update: ports.ResourceUpdate{HasUpdate: true, LastUpdate: updatedAt, Message: "github message"}}, stackoverflow: &stubStackOverflow{parseErr: errors.New("nope")}},
		{name: "stackoverflow update", github: &stubGithub{parseErr: errors.New("nope")}, stackoverflow: &stubStackOverflow{questionID: "123", update: ports.ResourceUpdate{HasUpdate: true, LastUpdate: updatedAt, Message: "stack message"}}},
		{name: "unsupported link", github: &stubGithub{parseErr: errors.New("nope")}, stackoverflow: &stubStackOverflow{parseErr: errors.New("nope")}, wantErr: ports.ErrLinkNotFound},
	}
	ctx := context.Background()
	for _, tt := range tests {
		svc := NewScrapper(noopLogger{}, &stubRepo{}, &stubBotClient{}, []scrapperinterfaces.Checker{tt.github, tt.stackoverflow}, 100, 4)
		update, err := svc.getResourceUpdate(ctx, models.TrackedLink{
			URL:        "https://example.com",
			LastUpdate: updatedAt.Add(-time.Minute),
		})
		if tt.wantErr != nil {
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("case %q: got err %v, want %v", tt.name, err, tt.wantErr)
			}
			continue
		}
		if err != nil {
			t.Fatalf("case %q: unexpected error %v", tt.name, err)
		}
		if !update.LastUpdate.Equal(updatedAt) {
			t.Fatalf("case %q: got update %v, want %v", tt.name, update.LastUpdate, updatedAt)
		}
		if tt.github != nil && tt.github.parseErr == nil && !tt.github.gotSince.Equal(updatedAt.Add(-time.Minute)) {
			t.Fatalf("case %q: got github since %v", tt.name, tt.github.gotSince)
		}
		if tt.stackoverflow != nil && tt.stackoverflow.parseErr == nil && !tt.stackoverflow.gotSince.Equal(updatedAt.Add(-time.Minute)) {
			t.Fatalf("case %q: got stackoverflow since %v", tt.name, tt.stackoverflow.gotSince)
		}
	}
}

func TestCheckLinks(t *testing.T) {
	oldTime := time.Date(2026, 3, 30, 18, 0, 0, 0, time.UTC)
	newTime := oldTime.Add(time.Hour)
	repo := &stubRepo{
		allLinks: []models.TrackedLink{{
			LinkID:     42,
			URL:        "https://github.com/user/repo",
			LastUpdate: oldTime,
			ChatIDs:    []int64{2, 1},
		}},
	}
	botClient := &stubBotClient{}
	github := &stubGithub{
		owner: "user",
		repo:  "repo",
		update: ports.ResourceUpdate{
			HasUpdate:  true,
			LastUpdate: newTime,
			Message:    "github message",
		},
	}
	stack := &stubStackOverflow{parseErr: errors.New("nope")}
	svc := NewScrapper(noopLogger{}, repo, botClient, []scrapperinterfaces.Checker{github, stack}, 100, 4)

	svc.CheckLinks(context.Background())

	if !slices.Equal(botClient.got.chatIDs, []int64{2, 1}) {
		t.Fatalf("got chat ids %v", botClient.got.chatIDs)
	}
	if botClient.got.url != "https://github.com/user/repo" {
		t.Fatalf("got url %q", botClient.got.url)
	}
	if botClient.got.description != "github message" {
		t.Fatalf("got desc %q", botClient.got.description)
	}
	if repo.updated.linkID != 42 || !repo.updated.lastUpdate.Equal(newTime) {
		t.Fatalf("got update args %+v", repo.updated)
	}
}

func TestCheckLinksSkipsWhenNothingChanged(t *testing.T) {
	current := time.Date(2026, 3, 30, 18, 0, 0, 0, time.UTC)
	repo := &stubRepo{
		allLinks: []models.TrackedLink{{
			LinkID:     42,
			URL:        "https://github.com/user/repo",
			LastUpdate: current,
			ChatIDs:    []int64{1},
		}},
	}
	botClient := &stubBotClient{}
	github := &stubGithub{
		owner: "user",
		repo:  "repo",
		update: ports.ResourceUpdate{
			HasUpdate:  false,
			LastUpdate: current,
			Message:    "",
		},
	}
	svc := NewScrapper(noopLogger{}, repo, botClient, []scrapperinterfaces.Checker{github, &stubStackOverflow{parseErr: errors.New("nope")}}, 100, 4)

	svc.CheckLinks(context.Background())

	if botClient.got.url != "" {
		t.Fatalf("expected no notification, got %+v", botClient.got)
	}
	if repo.updated.linkID != 0 {
		t.Fatalf("expected no repo update, got %+v", repo.updated)
	}
}

func TestCheckLinksSendsFailedLinksReport(t *testing.T) {
	current := time.Date(2026, 3, 30, 18, 0, 0, 0, time.UTC)
	repo := &stubRepo{
		allLinks: []models.TrackedLink{{
			LinkID:     42,
			URL:        "https://github.com/user/repo",
			LastUpdate: current,
			ChatIDs:    []int64{2, 1},
		}},
	}
	botClient := &stubBotClient{}
	github := &stubGithub{
		parseErr: errors.New("nope"),
	}
	stack := &stubStackOverflow{
		parseErr: errors.New("nope"),
	}
	svc := NewScrapper(noopLogger{}, repo, botClient, []scrapperinterfaces.Checker{github, stack}, 100, 4)

	svc.CheckLinks(context.Background())

	if len(botClient.failedReports) != 2 {
		t.Fatalf("expected 2 reports, got %+v", botClient.failedReports)
	}

	gotByChat := make(map[int64][]string, len(botClient.failedReports))
	for _, report := range botClient.failedReports {
		gotByChat[report.chatID] = report.urls
	}

	for _, chatID := range []int64{1, 2} {
		urls, ok := gotByChat[chatID]
		if !ok {
			t.Fatalf("missing report for chat %d: %+v", chatID, botClient.failedReports)
		}
		if len(urls) != 1 || urls[0] != "https://github.com/user/repo" {
			t.Fatalf("got report urls for chat %d: %v", chatID, urls)
		}
	}
}
