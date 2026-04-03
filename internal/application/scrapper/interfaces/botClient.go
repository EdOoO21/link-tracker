package scrapperinterfaces

import "context"

type BotClient interface {
	SendUpdates(ctx context.Context, chatIDs []int64, url, description string) error
}
