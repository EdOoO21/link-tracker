package bot

import (
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	loginf "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application"
	inf "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/interfaces"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	settings "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/settings/bot"
)

const (
	Unknown     = "unknown"
	HelpCommand = "/start - начало работы пользователя\n" +
		"/help - вывод списка доступных команд\n" +
		"/track - начать отслеживание ссылки (опционально с тегами)\n" +
		"/untrack - прекратить отслеживание ссылки\n" +
		"/list - вывести список всех отслеживаемых ссылок (опционально фильтр по тегу)"
	StartCommand         = "Добро пожаловать! Используйте /help, чтобы посмотреть доступные команды."
	UnknownCommand       = "Неизвестная команда. Воспользуйтесь /help, чтобы посмотреть список доступных команд."
	TrackNoUrl           = "Для того, чтобы начать отслеживание ссылки, пожалуйста, передайте ссылку в качестве параметра (/track google.com)."
	TrackExistedUrl      = "Ссылка уже отслеживается."
	TrackNotExistedUrl   = "Ссылка успешно добавлена к отслеживанию."
	UntrackNotOneUrl     = "Для того, чтобы перестать отслеживать ссылку, пожалуйста, передайте только ссылку в качестве параметра (/untrack google.com)."
	UntrackExistedUrl    = "Ссылка успешно удалена."
	UntrackNotExistedUrl = "Ссылка отстуствует среди подписок."
	LinksNotExist        = "Отслеживаемых ссылок не найдено."
)

type App struct {
	logger loginf.Logger
	config *settings.Config
	repo   inf.Repository
}

func NewApp(logger loginf.Logger, config *settings.Config, repo inf.Repository) *App {
	return &App{
		logger: logger,
		config: config,
		repo:   repo,
	}
}

func (a *App) Run(bot inf.TGBot) error {

	u := tgbotapi.NewUpdate(a.config.Offset)
	u.Timeout = a.config.Timeout
	updates := bot.GetUpdatesChan(u)

	a.logger.Info("got channel for updates")

	for update := range updates {
		if update.Message == nil {
			continue
		}
		chatID := update.Message.Chat.ID
		a.logger.Info("recieved update", "chatID", chatID) // так как у нас лог не публичный, то можем прокинуть в логи
		var msg tgbotapi.MessageConfig
		if update.Message.IsCommand() {
			command := update.Message.Command()
			switch command {
			case "help":
				msg = tgbotapi.NewMessage(chatID, HelpCommand)
			case "start":
				msg = tgbotapi.NewMessage(chatID, StartCommand)
			case "track":
				args := strings.Fields(update.Message.CommandArguments())
				var url string
				var tags []string
				switch len(args) {
				case 0:
					msg = tgbotapi.NewMessage(chatID, TrackNoUrl)
				case 1:
					url = args[0]
				default:
					url = args[0]
					tags = args[1:]
				}

				if url != "" {
					if ok := a.repo.TrackLink(chatID, url, tags); !ok {
						msg = tgbotapi.NewMessage(chatID, TrackExistedUrl)
					} else {
						msg = tgbotapi.NewMessage(chatID, TrackNotExistedUrl)
					}
				}
			case "untrack":
				args := strings.Fields(update.Message.CommandArguments())
				var url string
				if len(args) == 1 {
					url = args[0]
				} else {
					msg = tgbotapi.NewMessage(chatID, UntrackNotOneUrl)
				}

				if url != "" {
					if ok := a.repo.UnTrackLink(chatID, url); ok {
						msg = tgbotapi.NewMessage(chatID, UntrackExistedUrl)
					} else {
						msg = tgbotapi.NewMessage(chatID, UntrackNotExistedUrl)
					}
				}
			case "list":
				tags := strings.Fields(update.Message.CommandArguments())

				links := a.repo.ListLinks(chatID, tags)

				if len(links) == 0 {
					msg = tgbotapi.NewMessage(chatID, LinksNotExist)
				} else {
					msg = tgbotapi.NewMessage(chatID, LinksOutput(links))
				}
			default:
				msg = tgbotapi.NewMessage(chatID, UnknownCommand)
				command = Unknown // чтобы не засорять логгер при огромных текстах
			}
			a.logger.Info("recieved command", "chatID", chatID, "command", command)
			_, err := bot.Send(msg)
			if err != nil {
				a.logger.Error("error to send message", "error", err, "chatID", chatID, "command", command)
			}
			a.logger.Info("reply sent", "chatID", chatID, "command", command)
		} else {
			a.logger.Info("recieved text", "chatID", chatID)
		}
	}
	return nil
}

func LinksOutput(links []domain.Link) string {
	var text strings.Builder

	for i, v := range links {
		text.WriteString(strconv.Itoa(i + 1))
		text.WriteString(" ")
		text.WriteString(v.URL)
		text.WriteString("\n")
	}

	return text.String()
}
