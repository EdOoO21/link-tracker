package scrapper

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/models"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	ports "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
)

type noopLogger struct{}

func (noopLogger) Info(string, ...any)  {}
func (noopLogger) Warn(string, ...any)  {}
func (noopLogger) Error(string, ...any) {}

type mockScrapperService struct {
	addLinkInput    models.AddLink
	deleteLinkInput models.DeleteLink
	addChatInput    int64
	deleteChatInput int64
	getLinksChatID  int64
	getLinksTags    []string

	addLinkErr    error
	deleteLinkErr error
	addChatErr    error
	deleteChatErr error
	getLinksErr   error
	getLinksResp  []domain.Link
}

func (m *mockScrapperService) AddLink(link models.AddLink) error {
	m.addLinkInput = link
	return m.addLinkErr
}

func (m *mockScrapperService) GetLinks(chatID int64, tags []string) ([]domain.Link, error) {
	m.getLinksChatID = chatID
	m.getLinksTags = tags
	return m.getLinksResp, m.getLinksErr
}

func (m *mockScrapperService) DeleteLink(link models.DeleteLink) error {
	m.deleteLinkInput = link
	return m.deleteLinkErr
}

func (m *mockScrapperService) AddChat(chatID int64) error {
	m.addChatInput = chatID
	return m.addChatErr
}

func (m *mockScrapperService) DeleteChat(chatID int64) error {
	m.deleteChatInput = chatID
	return m.deleteChatErr
}

func TestHandleLinksHeaderValidation(t *testing.T) {
	handler := NewHandler(&mockScrapperService{}, noopLogger{})
	req := httptest.NewRequest(http.MethodGet, "/links", nil)
	w := httptest.NewRecorder()

	handler.handleLinks(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPostLinks(t *testing.T) {
	service := &mockScrapperService{}
	handler := NewHandler(service, noopLogger{})
	req := httptest.NewRequest(http.MethodPost, "/links", bytes.NewBufferString(`{"url":"https://github.com/user/repo","tags":["go","backend"]}`))
	req.Header.Set("Tg-Chat-Id", "42")
	w := httptest.NewRecorder()

	handler.handleLinks(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
	}
	if service.addLinkInput.ChatID != 42 || service.addLinkInput.URL != "https://github.com/user/repo" {
		t.Fatalf("got add link %+v", service.addLinkInput)
	}
}

func TestGetLinks(t *testing.T) {
	updatedAt := time.Date(2026, 3, 30, 12, 0, 0, 0, time.UTC)
	service := &mockScrapperService{getLinksResp: []domain.Link{{
		URL:        "https://github.com/user/repo",
		Tags:       map[string]struct{}{"go": {}, "backend": {}},
		LastUpdate: updatedAt,
	}}}
	handler := NewHandler(service, noopLogger{})
	req := httptest.NewRequest(http.MethodGet, "/links?tag=go", nil)
	req.Header.Set("Tg-Chat-Id", "7")
	w := httptest.NewRecorder()

	handler.handleLinks(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
	}
	if service.getLinksChatID != 7 {
		t.Fatalf("got chat id %d", service.getLinksChatID)
	}
	if len(service.getLinksTags) != 1 || service.getLinksTags[0] != "go" {
		t.Fatalf("got tags %v", service.getLinksTags)
	}
}

func TestDeleteLinksServiceError(t *testing.T) {
	service := &mockScrapperService{deleteLinkErr: ports.ErrLinkNotFound}
	handler := NewHandler(service, noopLogger{})
	req := httptest.NewRequest(http.MethodDelete, "/links", bytes.NewBufferString(`{"url":"https://github.com/user/repo"}`))
	req.Header.Set("Tg-Chat-Id", "42")
	w := httptest.NewRecorder()

	handler.handleLinks(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusConflict)
	}
}

func TestHandleTgChatID(t *testing.T) {
	service := &mockScrapperService{}
	handler := NewHandler(service, noopLogger{})

	postReq := httptest.NewRequest(http.MethodPost, "/tg-chat/13", nil)
	postW := httptest.NewRecorder()
	handler.handleTgChatID(postW, postReq)
	if postW.Code != http.StatusOK {
		t.Fatalf("got post status %d, want %d", postW.Code, http.StatusOK)
	}
	if service.addChatInput != 13 {
		t.Fatalf("got add chat %d", service.addChatInput)
	}

	deleteReq := httptest.NewRequest(http.MethodDelete, "/tg-chat/13", nil)
	deleteW := httptest.NewRecorder()
	handler.handleTgChatID(deleteW, deleteReq)
	if deleteW.Code != http.StatusOK {
		t.Fatalf("got delete status %d, want %d", deleteW.Code, http.StatusOK)
	}
	if service.deleteChatInput != 13 {
		t.Fatalf("got delete chat %d", service.deleteChatInput)
	}
}

func TestWriteServiceErrorFallback(t *testing.T) {
	w := httptest.NewRecorder()
	writeServiceError(w, errors.New("boom"), "fallback")
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("got status %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
