package bot

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/models"
)

type noopLogger struct{}

func (noopLogger) Info(string, ...any)  {}
func (noopLogger) Warn(string, ...any)  {}
func (noopLogger) Error(string, ...any) {}

type mockBotService struct {
	gotUpdates models.SendUpdates
	err        error
}

func (m *mockBotService) SendUpdateMessages(updates models.SendUpdates) error {
	m.gotUpdates = updates
	return m.err
}

func TestPostUpdates(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		serviceErr error
		wantStatus int
	}{
		{name: "success", body: `{"url":"https://github.com/user/repo","tgChatIds":[1,2],"description":"updated"}`, wantStatus: http.StatusOK},
		{name: "invalid json", body: `{`, wantStatus: http.StatusBadRequest},
		{name: "empty chat ids", body: `{"url":"https://github.com/user/repo","tgChatIds":[],"description":"updated"}`, wantStatus: http.StatusBadRequest},
		{name: "empty url", body: `{"url":"","tgChatIds":[1],"description":"updated"}`, wantStatus: http.StatusBadRequest},
		{name: "service error", body: `{"url":"https://github.com/user/repo","tgChatIds":[1],"description":"updated"}`, serviceErr: errors.New("boom"), wantStatus: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		service := &mockBotService{err: tt.serviceErr}
		handler := NewHandler(noopLogger{}, service)
		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewBufferString(tt.body))
		w := httptest.NewRecorder()

		handler.PostUpdates(w, req)

		if w.Code != tt.wantStatus {
			t.Fatalf("case %q: got status %d, want %d", tt.name, w.Code, tt.wantStatus)
		}
		if tt.wantStatus == http.StatusOK {
			if service.gotUpdates.URL != "https://github.com/user/repo" {
				t.Fatalf("case %q: got url %q", tt.name, service.gotUpdates.URL)
			}
		}
	}
}

func TestHandleUpdateMethodNotAllowed(t *testing.T) {
	handler := NewHandler(noopLogger{}, &mockBotService{})
	req := httptest.NewRequest(http.MethodGet, "/updates", nil)
	w := httptest.NewRecorder()

	handler.handleUpdate(w, req)

	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusMethodNotAllowed)
	}
}
