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

type ServiceServer struct {
	pb.UnimplementedBotServiceServer
	botService botinterfaces.BotService
	logger     ports.Logger
}

func NewBotServiceServer(logger ports.Logger, botService botinterfaces.BotService) *ServiceServer {
	return &ServiceServer{
		logger:     logger,
		botService: botService,
	}
}

func (b *ServiceServer) SendUpdates(_ context.Context, req *pb.SendUpdatesRequest) (*empty.Empty, error) {
	updates := serviceModels.SendUpdates{
		URL:         req.GetUrl(),
		ChatIDs:     req.GetTgChatIds(),
		Description: req.GetDescription(),
	}

	if err := b.botService.SendUpdateMessages(updates); err != nil {
		b.logger.Error("failed to send updates via grpc", "error", err)
		return nil, toStatusError(err)
	}

	return &empty.Empty{}, nil
}

//nolint:wrapcheck // gRPC handlers must return raw status errors to preserve status codes.
func toStatusError(err error) error {
	return status.Error(codes.Internal, err.Error())
}
