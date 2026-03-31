package settings

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"strconv"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var (
	ErrTokenEmpty         = errors.New("token is empty")
	ErrPathEmpty          = errors.New("commands path is empty")
	ErrScrapperURLEmpty   = errors.New("scrapper URL is empty")
	ErrBotURLEmpty        = errors.New("bot URL is empty")
	ErrScrapperURLInvalid = errors.New("scrapper URL is invalid")
	ErrBotURLInvalid      = errors.New("bot URL is invalid")
)

const (
	LongPollingTimeout = 60
	MeassageOffset     = 0
)

type ServiceURL struct {
	Full string
	Port int
}

func (u ServiceURL) HostPort() string {
	parsed, err := url.Parse(u.Full)
	if err != nil {
		return ""
	}
	return parsed.Host
}

type Config struct {
	Token       string
	Commands    []tgbotapi.BotCommand
	Timeout     int
	Offset      int
	ScrapperURL ServiceURL
	GRPCPort    int
}

func LoadConfig() (*Config, error) {
	token := os.Getenv("APP_TELEGRAM_TOKEN")
	path := os.Getenv("APP_TELEGRAM_COMMANDS_PATH")
	scrapperRawURL := os.Getenv("APP_SCRAPPER_BASE_URL")
	botRawURL := os.Getenv("APP_BOT_BASE_URL")

	if token == "" {
		return nil, ErrTokenEmpty
	}
	if path == "" {
		return nil, ErrPathEmpty
	}
	if scrapperRawURL == "" {
		return nil, ErrScrapperURLEmpty
	}
	if botRawURL == "" {
		return nil, ErrBotURLEmpty
	}

	scrapperURL, err := parseServiceURL(scrapperRawURL, ErrScrapperURLInvalid)
	if err != nil {
		return nil, err
	}
	botURL, err := parseServiceURL(botRawURL, ErrBotURLInvalid)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	var cmds []tgbotapi.BotCommand
	if err = json.Unmarshal(data, &cmds); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &Config{
		Token:       token,
		Commands:    cmds,
		Timeout:     LongPollingTimeout,
		Offset:      MeassageOffset,
		ScrapperURL: scrapperURL,
		GRPCPort:    botURL.Port,
	}, nil
}

func parseServiceURL(raw string, invalidErr error) (ServiceURL, error) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Host == "" || parsed.Port() == "" {
		return ServiceURL{}, invalidErr
	}

	port, err := strconv.Atoi(parsed.Port())
	if err != nil {
		return ServiceURL{}, invalidErr
	}

	return ServiceURL{Full: raw, Port: port}, nil
}
