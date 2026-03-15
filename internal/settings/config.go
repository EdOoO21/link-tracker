package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	ErrTokenEmpty = errors.New("token is empty")
	ErrPathEmpty  = errors.New("commands path is empty")
)

const (
	LongPollingTimeout = 60
	MeassageOffset     = 0
)

type Config struct {
	Token    string
	Commands []tgbotapi.BotCommand
	Timeout  int
	Offset   int
}

func LoadConfig() (*Config, error) {
	token := os.Getenv("APP_TELEGRAM_TOKEN")
	path := os.Getenv("APP_TELEGRAM_COMMANDS_PATH")

	if token == "" {
		return nil, ErrTokenEmpty
	}

	if path == "" {
		return nil, ErrPathEmpty
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cmds []tgbotapi.BotCommand

	if err = json.Unmarshal(data, &cmds); err != nil {
		return nil, fmt.Errorf("unmarshal config file: %w", err)
	}

	return &Config{
		Token:    token,
		Commands: cmds,
		Timeout:  LongPollingTimeout,
		Offset:   MeassageOffset,
	}, nil
}
