package bot

import (
	"context"
	"fmt"

	logger "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
	pc "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/proto/gen"
	"google.golang.org/grpc"
)

type GRPCClient struct {
	client pc.BotServiceClient
	logger logger.Logger
}

func NewGRPCBotClient(logger logger.Logger, conn grpc.ClientConnInterface) *GRPCClient {
	return &GRPCClient{
		client: pc.NewBotServiceClient(conn),
		logger: logger,
	}
}

func (b *GRPCClient) SendUpdates(ctx context.Context, chatIDs []int64, url, description string) error {
	updates := &pc.SendUpdatesRequest{
		Url:         url,
		TgChatIds:   chatIDs,
		Description: description,
	}

	if _, err := b.client.SendUpdates(ctx, updates); err != nil {
		return fmt.Errorf("updates send: %w", err)
	}
	return nil
}
