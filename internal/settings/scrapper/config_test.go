package scrapper

import (
	"errors"
	"testing"
)

func TestServiceURLHostPort(t *testing.T) {
	tests := []struct {
		name string
		url  ServiceURL
		want string
	}{
		{name: "valid host port", url: ServiceURL{Full: "http://localhost:9080", Port: 9080}, want: "localhost:9080"},
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
		{name: "valid url", raw: "http://localhost:9090", want: ServiceURL{Full: "http://localhost:9090", Port: 9090}},
		{name: "missing port", raw: "http://localhost", wantErr: ErrBotURLInvalid},
		{name: "bad url", raw: "bad://", wantErr: ErrBotURLInvalid},
	}

	for _, tt := range tests {
		got, err := parseServiceURL(tt.raw, ErrBotURLInvalid)
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
	t.Setenv("APP_BOT_BASE_URL", "http://localhost:9090")
	t.Setenv("APP_SCRAPPER_BASE_URL", "http://localhost:9080")
	t.Setenv("APP_DATABASE_URL", "postgres://localhost:5432/linktracker")
	t.Setenv("APP_DATABASE_USER", "postgres")
	t.Setenv("APP_DATABASE_PASSWORD", "postgres")
	t.Setenv("APP_DATABASE_ACCESS_TYPE", "sql")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BotURL.Full != "http://localhost:9090" || cfg.BotURL.Port != 9090 {
		t.Fatalf("got bot url %+v", cfg.BotURL)
	}
	if cfg.GRPCPort != 9080 {
		t.Fatalf("got grpc port %d", cfg.GRPCPort)
	}
	if cfg.DatabaseURL != "postgres://localhost:5432/linktracker" {
		t.Fatalf("got database url %q", cfg.DatabaseURL)
	}
	if cfg.DatabaseUser != "postgres" {
		t.Fatalf("got database user %q", cfg.DatabaseUser)
	}
	if cfg.DatabasePassword != "postgres" {
		t.Fatalf("got database password %q", cfg.DatabasePassword)
	}
	if cfg.AccessType != "sql" {
		t.Fatalf("got access type %q", cfg.AccessType)
	}
}

func TestLoadConfigReturnsBotURLError(t *testing.T) {
	t.Setenv("APP_BOT_BASE_URL", "")
	t.Setenv("APP_SCRAPPER_BASE_URL", "http://localhost:9080")
	t.Setenv("APP_DATABASE_URL", "postgres://localhost:5432/linktracker")
	t.Setenv("APP_DATABASE_USER", "postgres")
	t.Setenv("APP_DATABASE_PASSWORD", "postgres")
	t.Setenv("APP_DATABASE_ACCESS_TYPE", "sql")

	_, err := LoadConfig()
	if !errors.Is(err, ErrBotURLEmpty) {
		t.Fatalf("got err %v, want %v", err, ErrBotURLEmpty)
	}
}
