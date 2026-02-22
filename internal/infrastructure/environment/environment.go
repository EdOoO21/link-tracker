package environment

import (
	"encoding/json"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	settings "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/settings"
)

type Reader struct{}

func NewReader() *Reader { return &Reader{} }

func (r *Reader) GetEnv() (*settings.Config, error) {

	// я у себя в окружении сделал переменные окружения

	token := os.Getenv("APP_TELEGRAM_TOKEN")
	path := os.Getenv("APP_TELEGRAM_COMMANDS_PATH")
	var cmds []tgbotapi.BotCommand
	// дописать ошибки если path или token пустые

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, &cmds); err != nil {
		return nil, err
	}

	cfg := settings.NewConfig(token, cmds)
	return cfg, nil
}
