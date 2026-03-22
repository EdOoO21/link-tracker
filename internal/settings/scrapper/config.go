package scrapper

import (
	"errors"
	"os"
)

var (
	ErrBotURLEmpty = errors.New("bot URL is empty")
)

type Config struct {
	BotURL string
}

func LoadConfig() (*Config, error) {
	botURL := os.Getenv("APP_BOT_BASE_URL")

	if botURL == "" {
		return nil, ErrBotURLEmpty
	}

	return &Config{
		BotURL: botURL,
	}, nil
}
