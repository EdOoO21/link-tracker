package main

import (
	"fmt"
	"net"
	"os"
	"time"

	scrapper "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper"
	scrapperinterfaces "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/interfaces"
	botgrpc "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/clients/bot_grpc"
	github "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/clients/github"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/clients/stackoverflow"
	grpcscrapper "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/grpc/scrapper"
	logs "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/logger"
	repo "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/repository"
	settings "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/settings/scrapper"
	pb "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger := logs.NewLogger()
	repo := repo.NewRepo()
	cfg, err := settings.LoadConfig()
	if err != nil {
		logger.Error("fatal error", "error", err)
		os.Exit(1)
	}

	githubClient := github.NewGitHubClient()
	stackOverflowClient := stackoverflow.NewStackOverflowClient()
	botClient, botConn, err := newBotGRPCClient(logger, cfg.BotURL.HostPort())
	if err != nil {
		logger.Error("failed to create bot grpc client", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := botConn.Close(); err != nil {
			logger.Error("failed to close bot grpc connection", "error", err)
		}
	}()

	scrapperService := scrapper.NewScrapper(logger, repo, botClient, githubClient, stackOverflowClient)
	scrapperService.RunCron(1 * time.Minute)

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		logger.Error("failed to listen grpc", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := grpcListener.Close(); err != nil {
			logger.Error("failed to close grpc listener", "error", err)
		}
	}()

	grpcSrv := grpc.NewServer()
	defer grpcSrv.GracefulStop()

	scrapperGRPCServer := grpcscrapper.NewScrapperServiceServer(logger, scrapperService)
	pb.RegisterScrapperServiceServer(grpcSrv, scrapperGRPCServer)

	if err := grpcSrv.Serve(grpcListener); err != nil {
		logger.Error("grpc server died", "error", err)
	}
}

func newBotGRPCClient(logger *logs.Logger, addr string) (scrapperinterfaces.BotClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, err
	}

	client := botgrpc.NewGRPCBotClient(logger, conn)
	return client, conn, nil
}
