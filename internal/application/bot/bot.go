package bot

import (
	"errors"
	"net/url"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	inf "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/interfaces"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/models"
	domain "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	ports "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
	settings "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/settings/bot"
)

const (
	Unknown                  = "unknown"
	NotCommand               = "not a command"
	TextInvalidLinkGot       = "invalid link got"
	TextValidLinkGot         = "valid link got"
	URLAdded                 = "url added"
	URLExists                = "url exists"
	UnexpectedErrorToAddLink = "url addition failed"
	TextNotSupportedLinkGot  = "url domain not supported"
)

const (
	HelpCommand = "/start - начало работы пользователя\n" +
		"/help - вывод списка доступных команд\n" +
		"/track - начать отслеживание ссылки (опционально с тегами)\n" +
		"/untrack - прекратить отслеживание ссылки\n" +
		"/list - вывести список всех отслеживаемых ссылок (опционально фильтр по тегу)"
	StartCommand              = "Добро пожаловать! Используйте /help, чтобы посмотреть доступные команды. Вы можете отслеживать источники."
	StartCommandFailedChatAdd = "Добро пожаловать! Используйте /help, чтобы посмотреть доступные команды.\nК сожалению, не удалось предоставить вам возомжность отслеживания источников, попробуйте позже снова с помощью /start."
	UnknownCommand            = "Неизвестная команда. Воспользуйтесь /help, чтобы посмотреть список доступных команд."
	AddingToTrackSeqStarted   = "Введите корректный URL источника: "
	AddingToTrackSeqRestarted = "Процесс добавления источника начинается заново, введите корректный URL источника: "
	TrackNoURL                = "Для того, чтобы начать отслеживание ссылки, пожалуйста, передайте ссылку в качестве параметра (/track google.com)."
	TrackExistedURL           = "Ссылка уже отслеживается."
	TrackNotExistedURL        = "Ссылка успешно добавлена к отслеживанию."
	TrackErrorToAddLink       = "Произошла непредвиденная при добавлении ссыкли к отслеживанию.\nПопробуйте снова через некоторое время."
	UntrackNotOneURL          = "Для того, чтобы перестать отслеживать ссылку, пожалуйста, передайте только ссылку в качестве параметра (/untrack google.com)."
	UntrackExistedURL         = "Ссылка успешно удалена."
	UntrackNotExistedURL      = "Ссылка отстуствует среди подписок."
	UntrackErrToDeleteLink    = "Произошла непредвиденная при удалении ссыкли из отслеживания.\nПопробуйте снова через некоторое время."
	LinksNotExist             = "Отслеживаемых ссылок не найдено."
	LinksListFailed           = "Произошла непредвиденная при поиске всех отслеживаемых ссылок.\nПопробуйте снова через некоторое время."
	CancelCommand             = "Добавление ссылки к отслеживанию прервано."
	ChatIDNotFound            = "Вам не предоставлена возможность отслеживать источники.\n Попробуйте /start."
	UnknownText               = "Неопознанный текст. Воспользуйтесь /help для списка доступных команд."
	InvalidURL                = "Невалидная ссылка, пожалуйста, попробуйте снова."
	ValidURL                  = `Ссыслка успешно принята, далее отправьте теги в формате "тег, тег, тег..."`
	NotSupportedURL           = "Не поддерживаемый домен. На данный момент поддерживаются только github.com и stackoverflow.com"
)

const (
	LenOfToLongCommand = 30
)

type App struct {
	logger    ports.Logger
	config    *settings.Config
	repo      inf.Repository
	stMachine StateMachine
	bot       inf.TGBot
}

func NewApp(logger ports.Logger, config *settings.Config, repo inf.Repository, bot inf.TGBot) *App {
	return &App{
		logger:    logger,
		config:    config,
		repo:      repo,
		stMachine: make(StateMachine),
		bot:       bot,
	}
}

func (a *App) SendUpdateMessages(updates models.SendUpdates) error {
	failed := false
	for _, chatID := range updates.ChatIDS {
		msg := tgbotapi.NewMessage(chatID, linkUpdated(updates.URL, updates.Description))
		_, err := a.bot.Send(msg)
		if err != nil {
			failed = true
			a.logger.Error("error to send update message", "error", err, "chatID", chatID)
		}
		a.logger.Info("update sent", "chatID", chatID)
	}
	if failed {
		return errors.New("failed to send some update messages")
	}
	return nil
}

func (a *App) Run() error {
	u := tgbotapi.NewUpdate(a.config.Offset)
	u.Timeout = a.config.Timeout
	updates := a.bot.GetUpdatesChan(u)
	a.logger.Info("got channel for updates")
	for update := range updates {
		if update.Message == nil {
			continue
		}
		var msg tgbotapi.MessageConfig
		var command string
		chatID := update.Message.Chat.ID
		a.logger.Info("recieved update", "chatID", chatID)
		if !update.Message.IsCommand() {
			a.logger.Info("recieved text", "chatID", chatID)
			msg, command = a.computeText(chatID, update.Message.Text)
		} else {
			command = update.Message.Command()
			msg, command = a.moderateCommand(command, update)
			a.logger.Info("recieved command", "chatID", chatID, "command", command)
		}
		_, err := a.bot.Send(msg)
		if err != nil {
			a.logger.Error("error to send message", "error", err, "chatID", chatID, "command", command)
		}
		a.logger.Info("reply sent", "chatID", chatID, "command", command)
	}
	return nil
}

func (a *App) computeText(chatID int64, text string) (tgbotapi.MessageConfig, string) {
	var msg tgbotapi.MessageConfig
	command := NotCommand
	switch a.state(chatID).State {
	case TrackCommandGot:
		msg, command = a.computeGotLink(chatID, text)
	case LinkGot:
		msg, command = a.computeGotTags(chatID, text)
	default:
		msg = tgbotapi.NewMessage(chatID, UnknownText)
	}
	return msg, command
}

func (a *App) computeGotLink(chatID int64, text string) (tgbotapi.MessageConfig, string) {
	u, err := url.ParseRequestURI(text)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return tgbotapi.NewMessage(chatID, InvalidURL), TextInvalidLinkGot
	}
	if u.Host != "github.com" && u.Host != "stackoverflow.com" {
		return tgbotapi.NewMessage(chatID, NotSupportedURL), TextNotSupportedLinkGot
	}
	a.state(chatID).URL = u.String()
	a.state(chatID).State = LinkGot

	return tgbotapi.NewMessage(chatID, ValidURL), TextValidLinkGot
}

func (a *App) computeGotTags(chatID int64, text string) (tgbotapi.MessageConfig, string) {
	var msg tgbotapi.MessageConfig
	var command string
	t := strings.Split(text, ",")
	tags := make([]string, 0)
	for i := range t {
		tag := strings.TrimSpace(t[i])
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	if err := a.repo.TrackLink(chatID, a.state(chatID).URL, tags); errors.Is(err, ports.ErrLinkAlreadyExists) {
		msg = tgbotapi.NewMessage(chatID, TrackExistedURL)
		command = URLExists
	} else if errors.Is(err, ports.ErrChatNotFound) {
		msg = tgbotapi.NewMessage(chatID, ChatIDNotFound)
		command = ChatIDNotFound
	} else if err != nil {
		msg = tgbotapi.NewMessage(chatID, TrackErrorToAddLink)
		command = UnexpectedErrorToAddLink
	} else {
		msg = tgbotapi.NewMessage(chatID, TrackNotExistedURL)
		command = URLAdded
	}

	delete(a.stMachine, chatID)

	return msg, command
}

func (a *App) moderateCommand(command string, update tgbotapi.Update) (tgbotapi.MessageConfig, string) {
	chatID := update.Message.Chat.ID
	var msg tgbotapi.MessageConfig
	switch command {
	case "help":
		msg = a.computeHelp(chatID)
	case "start":
		msg = a.computeStart(chatID)
	case "track":
		msg = a.computeTrack(update)
	case "untrack":
		msg = a.computeUnTrack(update)
	case "list":
		msg = a.computeList(update)
	case "cancel":
		msg = a.computeCancel(chatID)
	default:
		msg, command = a.computeUnknownCommand(chatID, command)
	}
	return msg, command
}

func (a *App) computeHelp(chatID int64) tgbotapi.MessageConfig {
	if a.hasActiveState(chatID) {
		delete(a.stMachine, chatID)
	}
	return tgbotapi.NewMessage(chatID, HelpCommand)
}

func (a *App) computeStart(chatID int64) tgbotapi.MessageConfig {
	var msg tgbotapi.MessageConfig
	if a.hasActiveState(chatID) {
		delete(a.stMachine, chatID)
	}
	err := a.repo.AddChat(chatID)
	if err != nil && !errors.Is(err, ports.ErrChatAlreadyExists) {
		msg = tgbotapi.NewMessage(chatID, StartCommandFailedChatAdd)
	} else {
		msg = tgbotapi.NewMessage(chatID, StartCommand)
	}
	return msg
}

func (a *App) computeTrack(update tgbotapi.Update) tgbotapi.MessageConfig {
	var msg tgbotapi.MessageConfig
	chatID := update.Message.Chat.ID

	if a.state(chatID).State == NothingGot {
		msg = tgbotapi.NewMessage(chatID, AddingToTrackSeqStarted)
	} else {
		msg = tgbotapi.NewMessage(chatID, AddingToTrackSeqRestarted)
	}
	if _, ok := a.stMachine[chatID]; !ok {
		a.stMachine[chatID] = &StateInfo{State: TrackCommandGot, URL: ""}
	} else {
		a.stMachine[chatID].State = TrackCommandGot
		a.stMachine[chatID].URL = ""
	}
	return msg
}

func (a *App) computeUnTrack(update tgbotapi.Update) tgbotapi.MessageConfig {
	args := strings.Fields(update.Message.CommandArguments())
	var msg tgbotapi.MessageConfig
	var url string
	chatID := update.Message.Chat.ID
	if a.hasActiveState(chatID) {
		delete(a.stMachine, chatID)
	}
	if len(args) == 1 {
		url = args[0]
	} else {
		msg = tgbotapi.NewMessage(chatID, UntrackNotOneURL)
	}

	if url != "" {
		if err := a.repo.UnTrackLink(chatID, url); errors.Is(err, ports.ErrLinkNotFound) {
			msg = tgbotapi.NewMessage(chatID, UntrackNotExistedURL)
		} else if errors.Is(err, ports.ErrChatNotFound) {
			msg = tgbotapi.NewMessage(chatID, ChatIDNotFound)
		} else if err != nil {
			msg = tgbotapi.NewMessage(chatID, UntrackErrToDeleteLink)
		} else {
			msg = tgbotapi.NewMessage(chatID, UntrackExistedURL)
		}
	}
	return msg
}

func (a *App) computeList(update tgbotapi.Update) tgbotapi.MessageConfig {
	tags := strings.Fields(update.Message.CommandArguments())
	var msg tgbotapi.MessageConfig
	chatID := update.Message.Chat.ID
	if a.hasActiveState(chatID) {
		delete(a.stMachine, chatID)
	}
	links, err := a.repo.ListLinks(chatID, tags)
	if errors.Is(err, ports.ErrChatNotFound) {
		msg = tgbotapi.NewMessage(chatID, ChatIDNotFound)
	} else if err != nil {
		msg = tgbotapi.NewMessage(chatID, LinksListFailed)
	} else {
		if len(links) == 0 {
			msg = tgbotapi.NewMessage(chatID, LinksNotExist)
		} else {
			msg = tgbotapi.NewMessage(chatID, linksOutput(links))
		}
	}

	return msg
}

func (a *App) computeCancel(chatID int64) tgbotapi.MessageConfig {
	delete(a.stMachine, chatID)
	return tgbotapi.NewMessage(chatID, CancelCommand)
}

func (a *App) computeUnknownCommand(chatID int64, command string) (tgbotapi.MessageConfig, string) {
	if a.hasActiveState(chatID) {
		delete(a.stMachine, chatID)
	}
	msg := tgbotapi.NewMessage(chatID, UnknownCommand)

	if len(command) > LenOfToLongCommand {
		command = Unknown
	}
	return msg, command
}

func (a *App) hasActiveState(chatID int64) bool {
	stateInfo, ok := a.stMachine[chatID]
	return ok && stateInfo != nil && stateInfo.State != NothingGot
}

func (a *App) state(chatID int64) *StateInfo {
	stateInfo, ok := a.stMachine[chatID]
	if !ok || stateInfo == nil {
		stateInfo = &StateInfo{State: NothingGot}
		a.stMachine[chatID] = stateInfo
	}
	return stateInfo
}

func linkUpdated(link, description string) string {
	var text strings.Builder
	text.WriteString("Ссылка была обновлена!\n\n")
	text.WriteString(link)
	text.WriteString("\n\n")
	text.WriteString("Описание: \n\n")
	if description != "" {
		text.WriteString(description)
	} else {
		text.WriteString("отсутствует")
	}

	return text.String()
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
