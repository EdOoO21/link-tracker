package main

import (
	"fmt"
	"net"
	"os"

	app "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot"
	botinterfaces "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/interfaces"
	scrappergrpc "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/clients/scrapper_grpc"
	botinit "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/clients/telegram"
	grpcbot "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/grpc/bot"
	logs "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/logger"
	settings "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/settings/bot"
	pb "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger := logs.NewLogger()

	config, err := settings.LoadConfig()
	if err != nil {
		logger.Error("fatal error", "error", err)
		os.Exit(1)
	}

	bot, err := botinit.NewTGBot(logger, config)
	if err != nil {
		logger.Error("fatal error", "error", err)
		os.Exit(1)
	}

	scrapperClient, scrapperConn, err := newScrapperGRPCClient(logger, config.ScrapperURL.HostPort())
	if err != nil {
		logger.Error("failed to create scrapper grpc client", "error", err)
		os.Exit(1)
	}
	defer func() {
		if err := scrapperConn.Close(); err != nil {
			logger.Error("failed to close scrapper grpc connection", "error", err)
		}
	}()

	botService := app.NewApp(logger, config, scrapperClient, bot)

	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", config.GRPCPort))
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

	botGRPCServer := grpcbot.NewBotServiceServer(logger, botService)
	pb.RegisterBotServiceServer(grpcSrv, botGRPCServer)

	go func() {
		if err := grpcSrv.Serve(grpcListener); err != nil {
			logger.Error("grpc server died", "error", err)
		}
	}()

	if err = botService.Run(); err != nil {
		logger.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func newScrapperGRPCClient(logger *logs.Logger, addr string) (botinterfaces.Repository, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, err
	}

	client := scrappergrpc.NewGRPCScrapperClient(logger, conn)
	return client, conn, nil
}
