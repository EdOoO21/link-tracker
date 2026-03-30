package scrappergrpc

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	empty "github.com/golang/protobuf/ptypes/empty"
	ports "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
	pc "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type noopLogger struct{}

func (noopLogger) Info(string, ...any)  {}
func (noopLogger) Warn(string, ...any)  {}
func (noopLogger) Error(string, ...any) {}

type mockScrapperServiceClient struct {
	addLinkRequest   *pc.AddLinkRequest
	listLinksRequest *pc.ListLinksRequest
	listLinksResp    *pc.ListLinksResponse
	addLinkErr       error
	listLinksErr     error
}

func (m *mockScrapperServiceClient) AddChat(ctx context.Context, in *pc.ChatRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func (m *mockScrapperServiceClient) DeleteChat(ctx context.Context, in *pc.ChatRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func (m *mockScrapperServiceClient) AddLink(ctx context.Context, in *pc.AddLinkRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	m.addLinkRequest = in
	return &empty.Empty{}, m.addLinkErr
}

func (m *mockScrapperServiceClient) DeleteLink(ctx context.Context, in *pc.DeleteLinkRequest, opts ...grpc.CallOption) (*empty.Empty, error) {
	return &empty.Empty{}, nil
}

func (m *mockScrapperServiceClient) ListLinks(ctx context.Context, in *pc.ListLinksRequest, opts ...grpc.CallOption) (*pc.ListLinksResponse, error) {
	m.listLinksRequest = in
	return m.listLinksResp, m.listLinksErr
}

func TestGRPCScrapperClientTrackLink(t *testing.T) {
	mockClient := &mockScrapperServiceClient{}
	client := &Client{client: mockClient, logger: noopLogger{}}

	err := client.TrackLink(7, "https://github.com/user/repo", []string{"go", "backend"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := &pc.AddLinkRequest{ChatId: 7, Url: "https://github.com/user/repo", Tags: []string{"go", "backend"}}
	if !reflect.DeepEqual(mockClient.addLinkRequest, want) {
		t.Fatalf("got %+v, want %+v", mockClient.addLinkRequest, want)
	}
}

func TestGRPCScrapperClientListLinks(t *testing.T) {
	updatedAt := time.Date(2026, 3, 30, 15, 0, 0, 0, time.UTC)
	mockClient := &mockScrapperServiceClient{
		listLinksResp: &pc.ListLinksResponse{
			Links: []*pc.Link{{
				Url:        "https://github.com/user/repo",
				Tags:       []string{"go", "backend"},
				LastUpdate: timestamppb.New(updatedAt),
			}},
			Size: 1,
		},
	}
	client := &Client{client: mockClient, logger: noopLogger{}}

	links, err := client.ListLinks(5, []string{"go"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mockClient.listLinksRequest == nil {
		t.Fatal("expected list request to be sent")
	}
	if mockClient.listLinksRequest.GetChatId() != 5 {
		t.Fatalf("got chat id %d", mockClient.listLinksRequest.GetChatId())
	}
	if !reflect.DeepEqual(mockClient.listLinksRequest.GetTags(), []string{"go"}) {
		t.Fatalf("got tags %v", mockClient.listLinksRequest.GetTags())
	}
	if len(links) != 1 {
		t.Fatalf("got links len %d", len(links))
	}
	if links[0].URL != "https://github.com/user/repo" {
		t.Fatalf("got url %q", links[0].URL)
	}
	if _, ok := links[0].Tags["go"]; !ok {
		t.Fatalf("expected go tag in %v", links[0].Tags)
	}
	if _, ok := links[0].Tags["backend"]; !ok {
		t.Fatalf("expected backend tag in %v", links[0].Tags)
	}
	if !links[0].LastUpdate.Equal(updatedAt) {
		t.Fatalf("got last update %v, want %v", links[0].LastUpdate, updatedAt)
	}
}

func TestMapRPCError(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr error
	}{
		{name: "chat not found", err: status.Error(codes.NotFound, ports.ErrChatNotFound.Error()), wantErr: ports.ErrChatNotFound},
		{name: "link not found", err: status.Error(codes.NotFound, ports.ErrLinkNotFound.Error()), wantErr: ports.ErrLinkNotFound},
		{name: "chat exists", err: status.Error(codes.AlreadyExists, ports.ErrChatAlreadyExists.Error()), wantErr: ports.ErrChatAlreadyExists},
		{name: "link exists", err: status.Error(codes.AlreadyExists, ports.ErrLinkAlreadyExists.Error()), wantErr: ports.ErrLinkAlreadyExists},
	}

	for _, tt := range tests {
		got := mapRPCError("action", tt.err)
		if !errors.Is(got, tt.wantErr) {
			t.Fatalf("case %q: got err %v, want wrapping %v", tt.name, got, tt.wantErr)
		}
	}
}
