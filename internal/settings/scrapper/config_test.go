package scrapper

import (
	"errors"
	"testing"
	"time"
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
	if cfg.DB.DatabaseURL != "postgres://localhost:5432/linktracker" {
		t.Fatalf("got database url %q", cfg.DB.DatabaseURL)
	}
	if cfg.DB.DatabaseUser != "postgres" {
		t.Fatalf("got database user %q", cfg.DB.DatabaseUser)
	}
	if cfg.DB.DatabasePassword != "postgres" {
		t.Fatalf("got database password %q", cfg.DB.DatabasePassword)
	}
	if cfg.DB.AccessType != "sql" {
		t.Fatalf("got access type %q", cfg.DB.AccessType)
	}
	if cfg.BatchSize != 100 {
		t.Fatalf("got batch size %d, want 100", cfg.BatchSize)
	}
	if cfg.WorkerCount != 4 {
		t.Fatalf("got worker count %d, want 4", cfg.WorkerCount)
	}
	if cfg.CheckInterval != time.Minute {
		t.Fatalf("got check interval %s, want %s", cfg.CheckInterval, time.Minute)
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

func TestLoadConfigUsesConfiguredBatchSize(t *testing.T) {
	t.Setenv("APP_BOT_BASE_URL", "http://localhost:9090")
	t.Setenv("APP_SCRAPPER_BASE_URL", "http://localhost:9080")
	t.Setenv("APP_DATABASE_URL", "postgres://localhost:5432/linktracker")
	t.Setenv("APP_DATABASE_USER", "postgres")
	t.Setenv("APP_DATABASE_PASSWORD", "postgres")
	t.Setenv("APP_DATABASE_ACCESS_TYPE", "sql")
	t.Setenv("APP_SCRAPPER_BATCH_SIZE", "250")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.BatchSize != 250 {
		t.Fatalf("got batch size %d, want 250", cfg.BatchSize)
	}
}

func TestLoadConfigReturnsBatchSizeError(t *testing.T) {
	t.Setenv("APP_BOT_BASE_URL", "http://localhost:9090")
	t.Setenv("APP_SCRAPPER_BASE_URL", "http://localhost:9080")
	t.Setenv("APP_DATABASE_URL", "postgres://localhost:5432/linktracker")
	t.Setenv("APP_DATABASE_USER", "postgres")
	t.Setenv("APP_DATABASE_PASSWORD", "postgres")
	t.Setenv("APP_DATABASE_ACCESS_TYPE", "sql")
	t.Setenv("APP_SCRAPPER_BATCH_SIZE", "0")

	_, err := LoadConfig()
	if !errors.Is(err, ErrBatchSizeInvalid) {
		t.Fatalf("got err %v, want %v", err, ErrBatchSizeInvalid)
	}
}

func TestLoadConfigUsesConfiguredWorkerCount(t *testing.T) {
	t.Setenv("APP_BOT_BASE_URL", "http://localhost:9090")
	t.Setenv("APP_SCRAPPER_BASE_URL", "http://localhost:9080")
	t.Setenv("APP_DATABASE_URL", "postgres://localhost:5432/linktracker")
	t.Setenv("APP_DATABASE_USER", "postgres")
	t.Setenv("APP_DATABASE_PASSWORD", "postgres")
	t.Setenv("APP_DATABASE_ACCESS_TYPE", "sql")
	t.Setenv("APP_SCRAPPER_WORKER_COUNT", "8")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.WorkerCount != 8 {
		t.Fatalf("got worker count %d, want 8", cfg.WorkerCount)
	}
}

func TestLoadConfigReturnsWorkerCountError(t *testing.T) {
	t.Setenv("APP_BOT_BASE_URL", "http://localhost:9090")
	t.Setenv("APP_SCRAPPER_BASE_URL", "http://localhost:9080")
	t.Setenv("APP_DATABASE_URL", "postgres://localhost:5432/linktracker")
	t.Setenv("APP_DATABASE_USER", "postgres")
	t.Setenv("APP_DATABASE_PASSWORD", "postgres")
	t.Setenv("APP_DATABASE_ACCESS_TYPE", "sql")
	t.Setenv("APP_SCRAPPER_WORKER_COUNT", "0")

	_, err := LoadConfig()
	if !errors.Is(err, ErrWorkerCountInvalid) {
		t.Fatalf("got err %v, want %v", err, ErrWorkerCountInvalid)
	}
}

func TestLoadConfigUsesConfiguredCheckInterval(t *testing.T) {
	t.Setenv("APP_BOT_BASE_URL", "http://localhost:9090")
	t.Setenv("APP_SCRAPPER_BASE_URL", "http://localhost:9080")
	t.Setenv("APP_DATABASE_URL", "postgres://localhost:5432/linktracker")
	t.Setenv("APP_DATABASE_USER", "postgres")
	t.Setenv("APP_DATABASE_PASSWORD", "postgres")
	t.Setenv("APP_DATABASE_ACCESS_TYPE", "sql")
	t.Setenv("APP_SCRAPPER_CHECK_INTERVAL", "30s")

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.CheckInterval != 30*time.Second {
		t.Fatalf("got check interval %s, want 30s", cfg.CheckInterval)
	}
}

func TestLoadConfigReturnsCheckIntervalError(t *testing.T) {
	t.Setenv("APP_BOT_BASE_URL", "http://localhost:9090")
	t.Setenv("APP_SCRAPPER_BASE_URL", "http://localhost:9080")
	t.Setenv("APP_DATABASE_URL", "postgres://localhost:5432/linktracker")
	t.Setenv("APP_DATABASE_USER", "postgres")
	t.Setenv("APP_DATABASE_PASSWORD", "postgres")
	t.Setenv("APP_DATABASE_ACCESS_TYPE", "sql")
	t.Setenv("APP_SCRAPPER_CHECK_INTERVAL", "0s")

	_, err := LoadConfig()
	if !errors.Is(err, ErrCheckIntervalInvalid) {
		t.Fatalf("got err %v, want %v", err, ErrCheckIntervalInvalid)
	}
}
