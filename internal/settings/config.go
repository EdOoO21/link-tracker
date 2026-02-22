package settings

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Config struct {
	Token    string
	Commands []tgbotapi.BotCommand
}

func NewConfig(token string, cmds []tgbotapi.BotCommand) *Config {
	return &Config{
		Token:    token,
		Commands: cmds,
	}
}
