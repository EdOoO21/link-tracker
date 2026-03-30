package bot

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type noopLogger struct{}

func (noopLogger) Info(string, ...any)  {}
func (noopLogger) Warn(string, ...any)  {}
func (noopLogger) Error(string, ...any) {}

func TestSendUpdates(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    string
	}{
		{name: "success", statusCode: http.StatusOK},
		{name: "bad request", statusCode: http.StatusBadRequest, wantErr: "invalid request"},
		{name: "internal error", statusCode: http.StatusInternalServerError, wantErr: "failed to send updates to some chats"},
		{name: "unexpected status", statusCode: http.StatusTeapot, wantErr: "unexpected status: 418"},
	}

	for _, tt := range tests {
		var gotMethod string
		var gotContentType string
		var gotBody SendUpdatesRequest
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotContentType = r.Header.Get("Content-Type")
			if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			w.WriteHeader(tt.statusCode)
		}))

		client := NewClient(noopLogger{}, server.URL)
		err := client.SendUpdates([]int64{1, 2}, "https://github.com/user/repo", "updated")
		server.Close()

		if tt.wantErr == "" {
			if err != nil {
				t.Fatalf("case %q: unexpected error %v", tt.name, err)
			}
		} else if err == nil || err.Error() != tt.wantErr {
			t.Fatalf("case %q: got err %v, want %q", tt.name, err, tt.wantErr)
		}

		if gotMethod != http.MethodPost {
			t.Fatalf("case %q: got method %q", tt.name, gotMethod)
		}
		if gotContentType != "application/json" {
			t.Fatalf("case %q: got content type %q", tt.name, gotContentType)
		}
		if gotBody.URL != "https://github.com/user/repo" || gotBody.Description != "updated" {
			t.Fatalf("case %q: got body %+v", tt.name, gotBody)
		}
	}
}
