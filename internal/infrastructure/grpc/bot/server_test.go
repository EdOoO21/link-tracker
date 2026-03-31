package bot

import (
	"context"
	"errors"
	"reflect"
	"testing"

	serviceModels "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/models"
	pb "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/proto/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type noopLogger struct{}

func (noopLogger) Info(string, ...any)  {}
func (noopLogger) Warn(string, ...any)  {}
func (noopLogger) Error(string, ...any) {}

type mockBotService struct {
	gotUpdates serviceModels.SendUpdates
	err        error
}

func (m *mockBotService) SendUpdateMessages(updates serviceModels.SendUpdates) error {
	m.gotUpdates = updates
	return m.err
}

func TestBotServiceServerSendUpdates(t *testing.T) {
	tests := []struct {
		name        string
		serviceErr  error
		wantCode    codes.Code
		wantUpdates serviceModels.SendUpdates
	}{
		{
			name:       "success maps request to service model",
			serviceErr: nil,
			wantCode:   codes.OK,
			wantUpdates: serviceModels.SendUpdates{
				URL:         "https://github.com/user/repo",
				ChatIDs:     []int64{1, 2},
				Description: "updated",
			},
		},
		{
			name:       "service error becomes internal grpc status",
			serviceErr: errors.New("send failed"),
			wantCode:   codes.Internal,
			wantUpdates: serviceModels.SendUpdates{
				URL:         "https://github.com/user/repo",
				ChatIDs:     []int64{1, 2},
				Description: "updated",
			},
		},
	}

	for _, tt := range tests {
		service := &mockBotService{err: tt.serviceErr}
		server := NewBotServiceServer(noopLogger{}, service)
		req := &pb.SendUpdatesRequest{
			Url:         tt.wantUpdates.URL,
			TgChatIds:   tt.wantUpdates.ChatIDs,
			Description: tt.wantUpdates.Description,
		}

		resp, err := server.SendUpdates(context.Background(), req)
		if !reflect.DeepEqual(service.gotUpdates, tt.wantUpdates) {
			t.Fatalf("got updates %+v, want %+v", service.gotUpdates, tt.wantUpdates)
		}

		if tt.wantCode == codes.OK {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp == nil {
				t.Fatal("expected non-nil response")
			}
			continue
		}

		if resp != nil {
			t.Fatal("expected nil response on error")
		}
		if status.Code(err) != tt.wantCode {
			t.Fatalf("got code %v, want %v", status.Code(err), tt.wantCode)
		}
		if status.Convert(err).Message() != tt.serviceErr.Error() {
			t.Fatalf("got message %q, want %q", status.Convert(err).Message(), tt.serviceErr.Error())
		}
	}
}
