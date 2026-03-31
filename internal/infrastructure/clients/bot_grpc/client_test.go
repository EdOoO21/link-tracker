package bot

import (
	"context"
	"errors"
	"testing"

	empty "github.com/golang/protobuf/ptypes/empty"
	pc "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/proto/gen"
	"google.golang.org/grpc"
)

type noopLogger struct{}

func (noopLogger) Info(string, ...any)  {}
func (noopLogger) Warn(string, ...any)  {}
func (noopLogger) Error(string, ...any) {}

type mockBotServiceClient struct {
	gotRequest *pc.SendUpdatesRequest
	resp       *empty.Empty
	err        error
}

func (m *mockBotServiceClient) SendUpdates(_ context.Context, in *pc.SendUpdatesRequest, _ ...grpc.CallOption) (*empty.Empty, error) {
	m.gotRequest = in
	if m.resp == nil {
		m.resp = &empty.Empty{}
	}
	return m.resp, m.err
}

func TestGRPCClientSendUpdates(t *testing.T) {
	mockClient := &mockBotServiceClient{}
	client := &GRPCClient{client: mockClient, logger: noopLogger{}}

	err := client.SendUpdates([]int64{1, 2}, "https://github.com/user/repo", "updated")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mockClient.gotRequest == nil {
		t.Fatal("expected request to be sent")
	}
	if mockClient.gotRequest.GetUrl() != "https://github.com/user/repo" {
		t.Fatalf("got url %q", mockClient.gotRequest.GetUrl())
	}
	if len(mockClient.gotRequest.GetTgChatIds()) != 2 || mockClient.gotRequest.GetTgChatIds()[0] != 1 || mockClient.gotRequest.GetTgChatIds()[1] != 2 {
		t.Fatalf("got chat ids %v", mockClient.gotRequest.GetTgChatIds())
	}
	if mockClient.gotRequest.GetDescription() != "updated" {
		t.Fatalf("got description %q", mockClient.gotRequest.GetDescription())
	}
}

func TestGRPCClientSendUpdatesReturnsWrappedError(t *testing.T) {
	mockClient := &mockBotServiceClient{err: errors.New("boom")}
	client := &GRPCClient{client: mockClient, logger: noopLogger{}}

	err := client.SendUpdates([]int64{1}, "url", "desc")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "updates send: boom" {
		t.Fatalf("got error %q", err.Error())
	}
}
