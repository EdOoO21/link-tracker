package bot

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"unicode"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	inf "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/interfaces"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/models"
	domain "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	ports "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
	settings "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/settings/bot"
)

var errUnsupportedTrackedLink = errors.New("unsupported tracked link")

const (
	Unknown                  = "unknown"
	NotCommand               = "not a command"
	TextInvalidLinkGot       = "invalid link got"
	TextInvalidTagGot        = "invalid tag got"
	TextValidLinkGot         = "valid link got"
	URLAdded                 = "url added"
	URLExists                = "url exists"
	UnexpectedErrorToAddLink = "url addition failed"
	TextNotSupportedLinkGot  = "url domain not supported"
)

const (
	HelpCommand = "/start - начало работы пользователя\n" +
		"/help - вывод списка доступных команд\n" +
		"/track - начать процесс добавления ссылки к отслеживанию\n" +
		"/cancel - прекратить процесс добавления ссылки к отслеживанию\n" +
		"/untrack - прекратить отслеживание ссылки\n" +
		"/list - вывести список всех отслеживаемых ссылок (опционально фильтр по тегу)\n" +
		"/add_tag - добавить тег к отслеживаемой ссылке\n" +
		"/get_tags - вывести теги конкретной ссылки\n" +
		"/delete_tag - удалить тег у отслеживаемой ссылки"
	StartCommand              = "Добро пожаловать! Используйте /help, чтобы посмотреть доступные команды. Вы можете отслеживать источники."
	StartCommandFailedChatAdd = "Добро пожаловать! Используйте /help, чтобы посмотреть доступные команды.\nК сожалению, не удалось предоставить вам возомжность отслеживания источников, попробуйте позже снова с помощью /start."
	UnknownCommand            = "Неизвестная команда. Воспользуйтесь /help, чтобы посмотреть список доступных команд."
	AddingToTrackSeqStarted   = "Введите URL источника: GitHub-репозиторий вида https://github.com/owner/repo или вопрос StackOverflow вида https://stackoverflow.com/questions/123/title"
	AddingToTrackSeqRestarted = "Процесс добавления источника начинается заново. Введите GitHub-репозиторий вида https://github.com/owner/repo или вопрос StackOverflow вида https://stackoverflow.com/questions/123/title"
	TrackExistedURLWithReset  = "Ссылка уже отслеживается, процесс завершен неудачно."
	TrackExistedURLNoReset    = "Ссылка уже отслеживается, попробуйте отправить новую ссылку снова."
	TrackNotExistedURL        = "Ссылка успешно добавлена к отслеживанию."
	TrackErrorToAddLink       = "Произошла непредвиденная при добавлении ссыкли к отслеживанию.\nПопробуйте снова через некоторое время."
	UntrackNotOneURL          = "Для того, чтобы перестать отслеживать ссылку, пожалуйста, передайте только ссылку в качестве параметра (/untrack google.com)."
	UntrackExistedURL         = "Ссылка успешно удалена."
	UntrackNotExistedURL      = "Ссылка отстуствует среди подписок."
	UntrackErrToDeleteLink    = "Произошла непредвиденная при удалении ссыкли из отслеживания.\nПопробуйте снова через некоторое время."
	LinksNotExist             = "Отслеживаемых ссылок не найдено."
	LinksListFailed           = "Произошла непредвиденная при поиске всех отслеживаемых ссылок.\nПопробуйте снова через некоторое время."
	CancelCommand             = "Добавление ссылки к отслеживанию прервано."
	CancelCommandNothingGot   = "Нечего отменять: процесс добавления ссылки не был запущен или был прерван ранее другой командой. Используйте /help для описания команд."
	ChatIDNotFound            = "Вам не предоставлена возможность отслеживать источники.\nПопробуйте /start."
	UnknownText               = "Неопознанный текст. Воспользуйтесь /help для списка доступных команд."
	InvalidURL                = "Невалидная ссылка. Поддерживаются только GitHub-репозиторий вида https://github.com/owner/repo и вопрос StackOverflow вида https://stackoverflow.com/questions/123/title."
	ValidURL                  = `Ссылка успешно принята, далее отправьте теги без пробелов в формате "тег, тег, тег..." или "-" если без тегов`
	InvalidTags               = `Теги не должны содержать пробелы. Отправьте теги снова в формате "тег, тег, тег..." или "-" если без тегов`
	NotSupportedURL           = "Не поддерживаемый домен. На данный момент поддерживаются только GitHub-репозитории на github.com и вопросы StackOverflow на stackoverflow.com."
	AddTagUsage               = "Для добавления тега передайте ссылку и тег без пробелов (/add_tag https://github.com/user/repo backend)."
	AddTagSucceeded           = "Тег успешно добавлен."
	AddTagFailed              = "Произошла непредвиденная ошибка при добавлении тега.\nПопробуйте снова через некоторое время."
	AddTagAlreadyExists       = "Такой тег уже привязан к ссылке."
	DeleteTagUsage            = "Для удаления тега передайте ссылку и тег без пробелов (/delete_tag https://github.com/user/repo backend)."
	DeleteTagSucceeded        = "Тег успешно удален."
	DeleteTagFailed           = "Произошла непредвиденная ошибка при удалении тега.\nПопробуйте снова через некоторое время."
	DeleteTagNotFound         = "Такой тег не найден у ссылки."
	GetTagsUsage              = "Для просмотра тегов передайте ссылку (/get_tags https://github.com/user/repo)."
	GetTagsFailed             = "Произошла непредвиденная ошибка при получении тегов.\nПопробуйте снова через некоторое время."
	GetTagsEmpty              = "У ссылки нет тегов."
)

const (
	LenOfToLongCommand = 30
	URLAndTagArgsCount = 2
)

type App struct {
	logger    ports.Logger
	config    *settings.Config
	repo      inf.ScrapperClient
	stMachine StateMachine
	bot       inf.TGBot
}

func NewApp(logger ports.Logger, config *settings.Config, repo inf.ScrapperClient, bot inf.TGBot) *App {
	return &App{
		logger:    logger,
		config:    config,
		repo:      repo,
		stMachine: make(StateMachine),
		bot:       bot,
	}
}

func (a *App) SendUpdateMessages(ctx context.Context, updates models.SendUpdates) error {
	failed := false
	for _, chatID := range updates.ChatIDs {
		msg := tgbotapi.NewMessage(chatID, linkUpdated(updates.URL, updates.Description))
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("send updates to user: %w", err)
		}
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

func (a *App) Run(ctx context.Context) error {
	u := tgbotapi.NewUpdate(a.config.Offset)
	u.Timeout = a.config.Timeout
	updates := a.bot.GetUpdatesChan(u)
	a.logger.Info("got channel for updates")
	err := a.longPolling(ctx, updates)
	if err != nil {
		return fmt.Errorf("running bot: %w", err)
	}
	return nil
}

func (a *App) longPolling(ctx context.Context, updates tgbotapi.UpdatesChannel) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("bot long polling: %w", ctx.Err())
		case update, ok := <-updates:
			if !ok {
				return nil
			}
			if update.Message == nil {
				continue
			}
			var msg tgbotapi.MessageConfig
			var command string
			chatID := update.Message.Chat.ID
			a.logger.Info("recieved update", "chatID", chatID)
			if !update.Message.IsCommand() {
				a.logger.Info("recieved text", "chatID", chatID)
				msg, command = a.computeText(ctx, chatID, update.Message.Text)
			} else {
				command = update.Message.Command()
				msg, command = a.moderateCommand(ctx, command, update)
				a.logger.Info("recieved command", "chatID", chatID, "command", command)
			}
			_, err := a.bot.Send(msg)
			if err != nil {
				a.logger.Error("error to send message", "error", err, "chatID", chatID, "command", command)
			}
			a.logger.Info("reply sent", "chatID", chatID, "command", command)
		}
	}
}

func (a *App) computeText(ctx context.Context, chatID int64, text string) (tgbotapi.MessageConfig, string) {
	var msg tgbotapi.MessageConfig
	command := NotCommand
	switch a.stMachine.State(chatID).State {
	case NothingGot:
		msg = tgbotapi.NewMessage(chatID, UnknownText)
	case TrackCommandGot:
		msg, command = a.computeGotLink(ctx, chatID, text)
	case LinkGot:
		msg, command = a.computeGotTags(ctx, chatID, text)
	default:
		msg = tgbotapi.NewMessage(chatID, UnknownText)
	}
	return msg, command
}

func (a *App) computeGotLink(ctx context.Context, chatID int64, text string) (tgbotapi.MessageConfig, string) {
	link, err := validateTrackedLink(text)
	switch {
	case errors.Is(err, errUnsupportedTrackedLink):
		return tgbotapi.NewMessage(chatID, NotSupportedURL), TextNotSupportedLinkGot
	case err != nil:
		return tgbotapi.NewMessage(chatID, InvalidURL), TextInvalidLinkGot
	}
	if a.repo.IsLinkPresent(ctx, chatID, link) {
		return tgbotapi.NewMessage(chatID, TrackExistedURLNoReset), URLExists
	}
	a.stMachine.State(chatID).URL = link
	a.stMachine.State(chatID).State = LinkGot

	return tgbotapi.NewMessage(chatID, ValidURL), TextValidLinkGot
}

func (a *App) computeGotTags(ctx context.Context, chatID int64, text string) (tgbotapi.MessageConfig, string) {
	var msg tgbotapi.MessageConfig
	var command string
	if strings.TrimSpace(text) == "-" {
		text = ""
	}
	t := strings.Split(text, ",")
	tags := make([]string, 0)
	for i := range t {
		tag := strings.TrimSpace(t[i])
		if tag != "" {
			if !isValidTag(tag) {
				return tgbotapi.NewMessage(chatID, InvalidTags), TextInvalidTagGot
			}
			tags = append(tags, tag)
		}
	}
	switch err := a.repo.TrackLink(ctx, chatID, a.stMachine.State(chatID).URL, tags); {
	case errors.Is(err, ports.ErrLinkAlreadyExists):
		msg = tgbotapi.NewMessage(chatID, TrackExistedURLWithReset)
		command = URLExists
	case errors.Is(err, ports.ErrChatNotFound):
		msg = tgbotapi.NewMessage(chatID, ChatIDNotFound)
		command = ChatIDNotFound
	case err != nil:
		msg = tgbotapi.NewMessage(chatID, TrackErrorToAddLink)
		command = UnexpectedErrorToAddLink
	default:
		msg = tgbotapi.NewMessage(chatID, TrackNotExistedURL)
		command = URLAdded
	}

	a.stMachine.Reset(chatID)

	return msg, command
}

func (a *App) moderateCommand(ctx context.Context, command string, update tgbotapi.Update) (tgbotapi.MessageConfig, string) {
	chatID := update.Message.Chat.ID
	var msg tgbotapi.MessageConfig
	switch command {
	case "help":
		msg = a.computeHelp(chatID)
	case "start":
		msg = a.computeStart(ctx, chatID)
	case "track":
		msg = a.computeTrack(update)
	case "untrack":
		msg = a.computeUnTrack(ctx, update)
	case "list":
		msg = a.computeList(ctx, update)
	case "add_tag":
		msg = a.computeAddTag(ctx, update)
	case "get_tags":
		msg = a.computeGetTags(ctx, update)
	case "delete_tag":
		msg = a.computeDeleteTag(ctx, update)
	case "cancel":
		msg = a.computeCancel(chatID)
	default:
		msg, command = a.computeUnknownCommand(chatID, command)
	}
	return msg, command
}

func (a *App) computeHelp(chatID int64) tgbotapi.MessageConfig {
	if a.stMachine.HasActiveState(chatID) {
		a.stMachine.Reset(chatID)
	}
	return tgbotapi.NewMessage(chatID, HelpCommand)
}

func (a *App) computeStart(ctx context.Context, chatID int64) tgbotapi.MessageConfig {
	var msg tgbotapi.MessageConfig
	if a.stMachine.HasActiveState(chatID) {
		a.stMachine.Reset(chatID)
	}
	err := a.repo.AddChat(ctx, chatID)
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

	if a.stMachine.State(chatID).State == NothingGot {
		msg = tgbotapi.NewMessage(chatID, AddingToTrackSeqStarted)
	} else {
		msg = tgbotapi.NewMessage(chatID, AddingToTrackSeqRestarted)
	}
	a.stMachine.StartTrack(chatID)
	return msg
}

func (a *App) computeUnTrack(ctx context.Context, update tgbotapi.Update) tgbotapi.MessageConfig {
	args := strings.Fields(update.Message.CommandArguments())
	var msg tgbotapi.MessageConfig
	var url string
	chatID := update.Message.Chat.ID
	if a.stMachine.HasActiveState(chatID) {
		a.stMachine.Reset(chatID)
	}
	if len(args) == 1 {
		url = args[0]
	} else {
		msg = tgbotapi.NewMessage(chatID, UntrackNotOneURL)
	}

	if url != "" {
		switch err := a.repo.UnTrackLink(ctx, chatID, url); {
		case errors.Is(err, ports.ErrLinkNotFound):
			msg = tgbotapi.NewMessage(chatID, UntrackNotExistedURL)
		case errors.Is(err, ports.ErrChatNotFound):
			msg = tgbotapi.NewMessage(chatID, ChatIDNotFound)
		case err != nil:
			msg = tgbotapi.NewMessage(chatID, UntrackErrToDeleteLink)
		default:
			msg = tgbotapi.NewMessage(chatID, UntrackExistedURL)
		}
	}
	return msg
}

func (a *App) computeList(ctx context.Context, update tgbotapi.Update) tgbotapi.MessageConfig {
	tags := strings.Fields(update.Message.CommandArguments())
	var msg tgbotapi.MessageConfig
	chatID := update.Message.Chat.ID
	if a.stMachine.HasActiveState(chatID) {
		a.stMachine.Reset(chatID)
	}
	links, err := a.repo.ListLinks(ctx, chatID, tags)
	switch {
	case errors.Is(err, ports.ErrChatNotFound):
		msg = tgbotapi.NewMessage(chatID, ChatIDNotFound)
	case err != nil:
		msg = tgbotapi.NewMessage(chatID, LinksListFailed)
	case len(links) == 0:
		msg = tgbotapi.NewMessage(chatID, LinksNotExist)
	default:
		msg = tgbotapi.NewMessage(chatID, linksOutput(links))
	}

	return msg
}

func (a *App) computeCancel(chatID int64) tgbotapi.MessageConfig {
	if a.stMachine.State(chatID).State == NothingGot {
		return tgbotapi.NewMessage(chatID, CancelCommandNothingGot)
	}
	a.stMachine.Reset(chatID)
	return tgbotapi.NewMessage(chatID, CancelCommand)
}

func (a *App) computeAddTag(ctx context.Context, update tgbotapi.Update) tgbotapi.MessageConfig {
	return a.computeTagMutation(
		ctx,
		update,
		AddTagUsage,
		a.repo.AddTag,
		ports.ErrTagAlreadyExists,
		AddTagAlreadyExists,
		AddTagFailed,
		AddTagSucceeded,
	)
}

func (a *App) computeGetTags(ctx context.Context, update tgbotapi.Update) tgbotapi.MessageConfig {
	chatID := update.Message.Chat.ID
	if a.stMachine.HasActiveState(chatID) {
		a.stMachine.Reset(chatID)
	}

	urlArg, ok := parseURLArg(update.Message.CommandArguments())
	if !ok {
		return tgbotapi.NewMessage(chatID, GetTagsUsage)
	}

	tags, err := a.repo.GetTags(ctx, chatID, urlArg)
	switch {
	case errors.Is(err, ports.ErrChatNotFound):
		return tgbotapi.NewMessage(chatID, ChatIDNotFound)
	case errors.Is(err, ports.ErrLinkNotFound):
		return tgbotapi.NewMessage(chatID, UntrackNotExistedURL)
	case err != nil:
		return tgbotapi.NewMessage(chatID, GetTagsFailed)
	case len(tags) == 0:
		return tgbotapi.NewMessage(chatID, GetTagsEmpty)
	default:
		return tgbotapi.NewMessage(chatID, tagsOutput(urlArg, tags))
	}
}

func (a *App) computeDeleteTag(ctx context.Context, update tgbotapi.Update) tgbotapi.MessageConfig {
	return a.computeTagMutation(
		ctx,
		update,
		DeleteTagUsage,
		a.repo.DeleteTag,
		ports.ErrTagNotFound,
		DeleteTagNotFound,
		DeleteTagFailed,
		DeleteTagSucceeded,
	)
}

func (a *App) computeTagMutation(
	ctx context.Context,
	update tgbotapi.Update,
	usageText string,
	operation func(context.Context, int64, string, string) error,
	expectedErr error,
	expectedText string,
	failedText string,
	successText string,
) tgbotapi.MessageConfig {
	chatID := update.Message.Chat.ID
	if a.stMachine.HasActiveState(chatID) {
		a.stMachine.Reset(chatID)
	}

	urlArg, tagArg, ok := parseURLAndTagArgs(update.Message.CommandArguments())
	if !ok {
		return tgbotapi.NewMessage(chatID, usageText)
	}

	switch err := operation(ctx, chatID, urlArg, tagArg); {
	case errors.Is(err, ports.ErrChatNotFound):
		return tgbotapi.NewMessage(chatID, ChatIDNotFound)
	case errors.Is(err, ports.ErrLinkNotFound):
		return tgbotapi.NewMessage(chatID, UntrackNotExistedURL)
	case errors.Is(err, expectedErr):
		return tgbotapi.NewMessage(chatID, expectedText)
	case err != nil:
		return tgbotapi.NewMessage(chatID, failedText)
	default:
		return tgbotapi.NewMessage(chatID, successText)
	}
}

func (a *App) computeUnknownCommand(chatID int64, command string) (tgbotapi.MessageConfig, string) {
	if a.stMachine.HasActiveState(chatID) {
		a.stMachine.Reset(chatID)
	}
	msg := tgbotapi.NewMessage(chatID, UnknownCommand)

	if len(command) > LenOfToLongCommand {
		command = Unknown
	}
	return msg, command
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
		text.WriteString(". ")
		text.WriteString(v.URL)
		text.WriteString("\n")
	}

	return text.String()
}

func tagsOutput(url string, tags []string) string {
	var text strings.Builder
	text.WriteString("Теги для ссылки:\n\n")
	text.WriteString(url)
	text.WriteString("\n\n")
	for i, tag := range tags {
		text.WriteString(strconv.Itoa(i + 1))
		text.WriteString(". ")
		text.WriteString(tag)
		text.WriteString("\n")
	}

	return text.String()
}

func parseURLArg(raw string) (string, bool) {
	args := strings.Fields(raw)
	if len(args) != 1 {
		return "", false
	}

	return args[0], true
}

func parseURLAndTagArgs(raw string) (string, string, bool) {
	args := strings.Fields(raw)
	if len(args) != URLAndTagArgsCount {
		return "", "", false
	}
	if !isValidTag(args[1]) {
		return "", "", false
	}

	return args[0], args[1], true
}

func isValidTag(tag string) bool {
	if strings.TrimSpace(tag) == "" {
		return false
	}

	return !strings.ContainsFunc(tag, unicode.IsSpace)
}

func validateTrackedLink(raw string) (string, error) {
	u, err := url.ParseRequestURI(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", errors.New("invalid url")
	}

	switch u.Host {
	case "github.com":
		if !isValidGitHubRepoURL(u) {
			return "", errors.New("invalid github repo url")
		}
	case "stackoverflow.com":
		if !isValidStackOverflowQuestionURL(u) {
			return "", errors.New("invalid stackoverflow question url")
		}
	default:
		return "", errUnsupportedTrackedLink
	}

	u.Fragment = ""

	return u.String(), nil
}

func isValidGitHubRepoURL(u *url.URL) bool {
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	return len(parts) == 2 && parts[0] != "" && parts[1] != ""
}

func isValidStackOverflowQuestionURL(u *url.URL) bool {
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 || parts[0] != "questions" || parts[1] == "" {
		return false
	}

	_, err := strconv.ParseInt(parts[1], 10, 64)
	return err == nil
}
