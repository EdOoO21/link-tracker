package bot

import (
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	inf "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/interfaces"
	domain "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	logs "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
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
	TrackNoURL           = "Для того, чтобы начать отслеживание ссылки, пожалуйста, передайте ссылку в качестве параметра (/track google.com)."
	TrackExistedURL      = "Ссылка уже отслеживается."
	TrackNotExistedURL   = "Ссылка успешно добавлена к отслеживанию."
	UntrackNotOneURL     = "Для того, чтобы перестать отслеживать ссылку, пожалуйста, передайте только ссылку в качестве параметра (/untrack google.com)."
	UntrackExistedURL    = "Ссылка успешно удалена."
	UntrackNotExistedURL = "Ссылка отстуствует среди подписок."
	LinksNotExist        = "Отслеживаемых ссылок не найдено."
)

type App struct {
	logger logs.Logger
	config *settings.Config
	repo   inf.Repository
}

func NewApp(logger logs.Logger, config *settings.Config, repo inf.Repository) *App {
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
		a.logger.Info("recieved update", "chatID", chatID)
		if update.Message.IsCommand() {
			command := update.Message.Command()
			msg, command := a.ModerateCommand(command, update)
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

func (a *App) ModerateCommand(command string, update tgbotapi.Update) (tgbotapi.MessageConfig, string) {
	chatID := update.Message.Chat.ID
	var msg tgbotapi.MessageConfig
	switch command {
	case "help":
		msg = tgbotapi.NewMessage(chatID, HelpCommand)
	case "start":
		msg = tgbotapi.NewMessage(chatID, StartCommand)
	case "track":
		msg = a.computeTrack(update)
	case "untrack":
		msg = a.computeUnTrack(update)
	case "list":
		msg = a.computeList(update)
	default:
		msg = tgbotapi.NewMessage(chatID, UnknownCommand)
		command = Unknown
	}
	return msg, command
}

func (a *App) computeTrack(update tgbotapi.Update) tgbotapi.MessageConfig {
	args := strings.Fields(update.Message.CommandArguments())
	var url string
	var tags []string
	var msg tgbotapi.MessageConfig
	chatID := update.Message.Chat.ID
	switch len(args) {
	case 0:
		msg = tgbotapi.NewMessage(chatID, TrackNoURL)
	case 1:
		url = args[0]
	default:
		url = args[0]
		tags = args[1:]
	}
	if url != "" {
		if ok := a.repo.TrackLink(chatID, url, tags); !ok {
			msg = tgbotapi.NewMessage(chatID, TrackExistedURL)
		} else {
			msg = tgbotapi.NewMessage(chatID, TrackNotExistedURL)
		}
	}
	return msg
}

func (a *App) computeUnTrack(update tgbotapi.Update) tgbotapi.MessageConfig {
	args := strings.Fields(update.Message.CommandArguments())
	var msg tgbotapi.MessageConfig
	var url string
	chatID := update.Message.Chat.ID
	if len(args) == 1 {
		url = args[0]
	} else {
		msg = tgbotapi.NewMessage(chatID, UntrackNotOneURL)
	}

	if url != "" {
		if ok := a.repo.UnTrackLink(chatID, url); ok {
			msg = tgbotapi.NewMessage(chatID, UntrackExistedURL)
		} else {
			msg = tgbotapi.NewMessage(chatID, UntrackNotExistedURL)
		}
	}
	return msg
}

func (a *App) computeList(update tgbotapi.Update) tgbotapi.MessageConfig {
	tags := strings.Fields(update.Message.CommandArguments())
	var msg tgbotapi.MessageConfig
	chatID := update.Message.Chat.ID
	links := a.repo.ListLinks(chatID, tags)

	if len(links) == 0 {
		msg = tgbotapi.NewMessage(chatID, LinksNotExist)
	} else {
		msg = tgbotapi.NewMessage(chatID, linksOutput(links))
	}
	return msg
}

func linksOutput(links []domain.Link) string {
	var text strings.Builder
	text.WriteString("Ссылки:")
	text.WriteString("\n\n")
	for i, v := range links {
		text.WriteString(strconv.Itoa(i + 1))
		text.WriteString(" ")
		text.WriteString(v.URL)
		text.WriteString("\n")
	}

	return text.String()
}
