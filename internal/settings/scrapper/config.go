package scrapper

import (
	"errors"
	"os"
)

var (
	ErrBotURLEmpty          = errors.New("bot URL is empty")
	ErrDatabaseURLEmpty     = errors.New("database URL is empty")
	ErrDatabaseUserEmpty    = errors.New("database user is empty")
	ErrDatabasePasswordEmpty = errors.New("database password is empty")
	ErrAccessTypeEmpty      = errors.New("access type is empty")
)

type Config struct {
	BotURL           string
	DatabaseURL      string
	DatabaseUser     string
	DatabasePassword string
	AccessType       string
}

func LoadConfig() (*Config, error) {
	botURL := os.Getenv("APP_BOT_BASE_URL")
	databaseURL := os.Getenv("APP_DATABASE_URL")
	databaseUser := os.Getenv("APP_DATABASE_USER")
	databasePassword := os.Getenv("APP_DATABASE_PASSWORD")
	accessType := os.Getenv("APP_DATABASE_ACCESS_TYPE")

	if botURL == "" {
		return nil, ErrBotURLEmpty
	}
	if databaseURL == "" {
		return nil, ErrDatabaseURLEmpty
	}
	if databaseUser == "" {
		return nil, ErrDatabaseUserEmpty
	}
	if databasePassword == "" {
		return nil, ErrDatabasePasswordEmpty
	}
	if accessType == "" {
		return nil, ErrAccessTypeEmpty
	}

	return &Config{
		BotURL:           botURL,
		DatabaseURL:      databaseURL,
		DatabaseUser:     databaseUser,
		DatabasePassword: databasePassword,
		AccessType:       accessType,
	}, nil
}
