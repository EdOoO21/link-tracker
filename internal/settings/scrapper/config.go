package scrapper

import (
	"errors"
	"net/url"
	"os"
	"strconv"
)

var (
	ErrBotURLEmpty        = errors.New("bot URL is empty")
	ErrScrapperURLEmpty   = errors.New("scrapper URL is empty")
	ErrBotURLInvalid      = errors.New("bot URL is invalid")
	ErrScrapperURLInvalid = errors.New("scrapper URL is invalid")
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
	BotURL   ServiceURL
	GRPCPort int
}

func LoadConfig() (*Config, error) {
	botRawURL := os.Getenv("APP_BOT_BASE_URL")
	scrapperRawURL := os.Getenv("APP_SCRAPPER_BASE_URL")

	if botRawURL == "" {
		return nil, ErrBotURLEmpty
	}
	if scrapperRawURL == "" {
		return nil, ErrScrapperURLEmpty
	}

	botURL, err := parseServiceURL(botRawURL, ErrBotURLInvalid)
	if err != nil {
		return nil, err
	}
	scrapperURL, err := parseServiceURL(scrapperRawURL, ErrScrapperURLInvalid)
	if err != nil {
		return nil, err
	}

	return &Config{
		BotURL:   botURL,
		GRPCPort: scrapperURL.Port,
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
