package scrapperinterfaces

type BotClient interface {
	SendUpdates(chatIDS []int64, url, description string) error
}
