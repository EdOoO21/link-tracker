package settings

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestServiceURLHostPort(t *testing.T) {
	tests := []struct {
		name string
		url  ServiceURL
		want string
	}{
		{name: "valid host port", url: ServiceURL{Full: "http://localhost:9090", Port: 9090}, want: "localhost:9090"},
		{name: "invalid url", url: ServiceURL{Full: "://bad", Port: 0}, want: ""},
	}

	for _, tt := range tests {
		got := tt.url.HostPort()
		if got != tt.want {
			t.Fatalf("case %q: got %q, want %q", tt.name, got, tt.want)
		}
	}
}

func TestParseServiceURL(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    ServiceURL
		wantErr error
	}{
		{name: "valid url", raw: "http://localhost:9080", want: ServiceURL{Full: "http://localhost:9080", Port: 9080}},
		{name: "missing port", raw: "http://localhost", wantErr: ErrScrapperURLInvalid},
		{name: "bad url", raw: "://broken", wantErr: ErrScrapperURLInvalid},
	}

	for _, tt := range tests {
		got, err := parseServiceURL(tt.raw, ErrScrapperURLInvalid)
		if tt.wantErr != nil {
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("case %q: got err %v, want %v", tt.name, err, tt.wantErr)
			}
			continue
		}
		if err != nil {
			t.Fatalf("case %q: unexpected error %v", tt.name, err)
		}
		if got != tt.want {
			t.Fatalf("case %q: got %+v, want %+v", tt.name, got, tt.want)
		}
	}
}

func TestLoadConfig(t *testing.T) {
	tmpDir := t.TempDir()
	commandsPath := filepath.Join(tmpDir, "commands.json")
	data := []byte(`[
		{"command":"start","description":"start bot"},
		{"command":"help","description":"help user"}
	]`)
	if err := os.WriteFile(commandsPath, data, 0o600); err != nil {
		t.Fatalf("write commands file: %v", err)
	}

	t.Setenv("APP_TELEGRAM_TOKEN", "token")
	t.Setenv("APP_TELEGRAM_COMMANDS_PATH", commandsPath)
	t.Setenv("APP_SCRAPPER_BASE_URL", "http://localhost:9080")
	t.Setenv("APP_BOT_BASE_URL", "http://localhost:9090")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Token != "token" {
		t.Fatalf("got token %q", cfg.Token)
	}
	if cfg.Timeout != LongPollingTimeout {
		t.Fatalf("got timeout %d, want %d", cfg.Timeout, LongPollingTimeout)
	}
	if cfg.Offset != MeassageOffset {
		t.Fatalf("got offset %d, want %d", cfg.Offset, MeassageOffset)
	}
	if cfg.ScrapperURL.Full != "http://localhost:9080" || cfg.ScrapperURL.Port != 9080 {
		t.Fatalf("got scrapper url %+v", cfg.ScrapperURL)
	}
	if cfg.GRPCPort != 9090 {
		t.Fatalf("got grpc port %d", cfg.GRPCPort)
	}
	if len(cfg.Commands) != 2 {
		t.Fatalf("got commands len %d", len(cfg.Commands))
	}

	commands := []string{cfg.Commands[0].Command, cfg.Commands[1].Command}
	sort.Strings(commands)
	if !reflect.DeepEqual(commands, []string{"help", "start"}) {
		t.Fatalf("got commands %v", commands)
	}
}

func TestLoadConfigReturnsTokenError(t *testing.T) {
	tmpDir := t.TempDir()
	commandsPath := filepath.Join(tmpDir, "commands.json")
	if err := os.WriteFile(commandsPath, []byte(`[]`), 0o600); err != nil {
		t.Fatalf("write commands file: %v", err)
	}

	t.Setenv("APP_TELEGRAM_TOKEN", "")
	t.Setenv("APP_TELEGRAM_COMMANDS_PATH", commandsPath)
	t.Setenv("APP_SCRAPPER_BASE_URL", "http://localhost:9080")
	t.Setenv("APP_BOT_BASE_URL", "http://localhost:9090")

	_, err := LoadConfig()
	if !errors.Is(err, ErrTokenEmpty) {
		t.Fatalf("got err %v, want %v", err, ErrTokenEmpty)
	}
}
