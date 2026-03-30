package bot

import (
	"context"

	empty "github.com/golang/protobuf/ptypes/empty"
	botinterfaces "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/interfaces"
	serviceModels "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/models"
	ports "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
	pb "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/proto/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type BotServiceServer struct {
	pb.UnimplementedBotServiceServer
	botService botinterfaces.BotService
	logger     ports.Logger
}

func NewBotServiceServer(logger ports.Logger, botService botinterfaces.BotService) *BotServiceServer {
	return &BotServiceServer{
		logger:     logger,
		botService: botService,
	}
}

func (b *BotServiceServer) SendUpdates(ctx context.Context, req *pb.SendUpdatesRequest) (*empty.Empty, error) {
	updates := serviceModels.SendUpdates{
		URL:         req.GetUrl(),
		ChatIDS:     req.GetTgChatIds(),
		Description: req.GetDescription(),
	}

	if err := b.botService.SendUpdateMessages(updates); err != nil {
		b.logger.Error("failed to send updates via grpc", "error", err)
		return nil, toStatusError(err)
	}

	return &empty.Empty{}, nil
}

func toStatusError(err error) error {
	return status.Error(codes.Internal, err.Error())
}
