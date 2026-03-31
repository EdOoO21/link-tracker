package scrapperinterfaces

type BotClient interface {
	SendUpdates(chatIDs []int64, url, description string) error
}
