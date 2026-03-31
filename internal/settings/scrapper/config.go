package scrapper

import (
	"errors"
	"net/url"
	"os"
	"strconv"
)

var (
	ErrBotURLEmpty           = errors.New("bot URL is empty")
	ErrScrapperURLEmpty      = errors.New("scrapper URL is empty")
	ErrBotURLInvalid         = errors.New("bot URL is invalid")
	ErrScrapperURLInvalid    = errors.New("scrapper URL is invalid")
	ErrDatabaseURLEmpty      = errors.New("database URL is empty")
	ErrDatabaseUserEmpty     = errors.New("database user is empty")
	ErrDatabasePasswordEmpty = errors.New("database password is empty")
	ErrAccessTypeEmpty       = errors.New("access type is empty")
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
	BotURL           ServiceURL
	GRPCPort         int
	DatabaseURL      string
	DatabaseUser     string
	DatabasePassword string
	AccessType       string
}

func LoadConfig() (*Config, error) {
	botRawURL := os.Getenv("APP_BOT_BASE_URL")
	scrapperRawURL := os.Getenv("APP_SCRAPPER_BASE_URL")
	databaseURL := os.Getenv("APP_DATABASE_URL")
	databaseUser := os.Getenv("APP_DATABASE_USER")
	databasePassword := os.Getenv("APP_DATABASE_PASSWORD")
	accessType := os.Getenv("APP_DATABASE_ACCESS_TYPE")

	if botRawURL == "" {
		return nil, ErrBotURLEmpty
	}
	if scrapperRawURL == "" {
		return nil, ErrScrapperURLEmpty
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

	botURL, err := parseServiceURL(botRawURL, ErrBotURLInvalid)
	if err != nil {
		return nil, err
	}
	scrapperURL, err := parseServiceURL(scrapperRawURL, ErrScrapperURLInvalid)
	if err != nil {
		return nil, err
	}

	return &Config{
		BotURL:           botURL,
		GRPCPort:         scrapperURL.Port,
		DatabaseURL:      databaseURL,
		DatabaseUser:     databaseUser,
		DatabasePassword: databasePassword,
		AccessType:       accessType,
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
