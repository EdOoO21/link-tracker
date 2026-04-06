package scrapper

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/models"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	ports "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
	pb "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/proto/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type noopLogger struct{}

func (noopLogger) Info(string, ...any)  {}
func (noopLogger) Warn(string, ...any)  {}
func (noopLogger) Error(string, ...any) {}

type mockScrapperService struct {
	gotAddLink    models.AddLink
	gotAddTag     models.AddTag
	gotDeleteTag  models.DeleteTag
	addLinkErr    error
	addTagErr     error
	deleteTagErr  error
	listLinksResp []domain.Link
	listLinksErr  error
	getTagsResp   []string
	getTagsErr    error
}

func (m *mockScrapperService) AddLink(_ context.Context, link models.AddLink) error {
	m.gotAddLink = link
	return m.addLinkErr
}

func (m *mockScrapperService) AddTag(_ context.Context, tag models.AddTag) error {
	m.gotAddTag = tag
	return m.addTagErr
}

func (m *mockScrapperService) GetTags(_ context.Context, _ int64, _ string) ([]string, error) {
	return m.getTagsResp, m.getTagsErr
}

func (m *mockScrapperService) DeleteTag(_ context.Context, tag models.DeleteTag) error {
	m.gotDeleteTag = tag
	return m.deleteTagErr
}

func (m *mockScrapperService) GetLinks(_ context.Context, _ int64, _ []string) ([]domain.Link, error) {
	return m.listLinksResp, m.listLinksErr
}

func (m *mockScrapperService) DeleteLink(_ context.Context, _ models.DeleteLink) error { return nil }
func (m *mockScrapperService) AddChat(_ context.Context, _ int64) error                { return nil }
func (m *mockScrapperService) DeleteChat(_ context.Context, _ int64) error             { return nil }

func TestScrapperServiceServerAddLink(t *testing.T) {
	service := &mockScrapperService{}
	server := NewScrapperServiceServer(noopLogger{}, service)
	req := &pb.AddLinkRequest{ChatId: 42, Url: "https://github.com/user/repo", Tags: []string{"go", "backend"}}

	resp, err := server.AddLink(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	want := models.AddLink{ChatID: 42, URL: "https://github.com/user/repo", Tags: []string{"go", "backend"}}
	if !reflect.DeepEqual(service.gotAddLink, want) {
		t.Fatalf("got %+v, want %+v", service.gotAddLink, want)
	}
}

func TestScrapperServiceServerListLinks(t *testing.T) {
	updatedAt := time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC)
	service := &mockScrapperService{
		listLinksResp: []domain.Link{{
			URL:        "https://github.com/user/repo",
			Tags:       map[string]struct{}{"backend": {}, "go": {}},
			LastUpdate: updatedAt,
		}},
	}
	server := NewScrapperServiceServer(noopLogger{}, service)

	resp, err := server.ListLinks(context.Background(), &pb.ListLinksRequest{ChatId: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetSize() != 1 {
		t.Fatalf("got size %d", resp.GetSize())
	}
	if len(resp.GetLinks()) != 1 {
		t.Fatalf("got links len %d", len(resp.GetLinks()))
	}
	if resp.GetLinks()[0].GetUrl() != "https://github.com/user/repo" {
		t.Fatalf("got url %q", resp.GetLinks()[0].GetUrl())
	}
	if !reflect.DeepEqual(resp.GetLinks()[0].GetTags(), []string{"backend", "go"}) {
		t.Fatalf("got tags %v", resp.GetLinks()[0].GetTags())
	}
	if got := resp.GetLinks()[0].GetLastUpdate().AsTime(); !got.Equal(updatedAt) {
		t.Fatalf("got last update %v, want %v", got, updatedAt)
	}
}

func TestScrapperServiceServerAddTag(t *testing.T) {
	service := &mockScrapperService{}
	server := NewScrapperServiceServer(noopLogger{}, service)
	req := &pb.AddTagRequest{ChatId: 42, Url: "https://github.com/user/repo", Tag: "backend"}

	resp, err := server.AddTag(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}

	want := models.AddTag{ChatID: 42, URL: "https://github.com/user/repo", Tag: "backend"}
	if !reflect.DeepEqual(service.gotAddTag, want) {
		t.Fatalf("got %+v, want %+v", service.gotAddTag, want)
	}
}

func TestScrapperServiceServerGetTags(t *testing.T) {
	service := &mockScrapperService{getTagsResp: []string{"backend", "go"}}
	server := NewScrapperServiceServer(noopLogger{}, service)

	resp, err := server.GetTags(context.Background(), &pb.GetTagsRequest{ChatId: 1, Url: "https://github.com/user/repo"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.GetSize() != 2 {
		t.Fatalf("got size %d", resp.GetSize())
	}
	if !reflect.DeepEqual(resp.GetTags(), []string{"backend", "go"}) {
		t.Fatalf("got tags %v", resp.GetTags())
	}
}

func TestScrapperToStatusError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		wantCode codes.Code
	}{
		{name: "chat not found", err: ports.ErrChatNotFound, wantCode: codes.NotFound},
		{name: "link not found", err: ports.ErrLinkNotFound, wantCode: codes.NotFound},
		{name: "chat exists", err: ports.ErrChatAlreadyExists, wantCode: codes.AlreadyExists},
		{name: "link exists", err: ports.ErrLinkAlreadyExists, wantCode: codes.AlreadyExists},
		{name: "tag exists", err: ports.ErrTagAlreadyExists, wantCode: codes.AlreadyExists},
		{name: "tag not found", err: ports.ErrTagNotFound, wantCode: codes.NotFound},
		{name: "unknown", err: errors.New("boom"), wantCode: codes.Internal},
	}

	for _, tt := range tests {
		got := toStatusError(tt.err)
		if status.Code(got) != tt.wantCode {
			t.Fatalf("case %q: got code %v, want %v", tt.name, status.Code(got), tt.wantCode)
		}
	}
}
