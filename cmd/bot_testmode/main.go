package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	app "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot"
	botinterfaces "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/bot/interfaces"
	botinit "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/e2e"
	scrappergrpc "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/clients/scrapper_grpc"
	grpcbot "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/grpc/bot"
	logs "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/logger"
	settings "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/settings/bot"
	pb "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger := logs.NewLogger()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	err := run(ctx, stop, logger)

	if err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, stop context.CancelFunc, logger *logs.Logger) error {
	defer stop()
	config, err := settings.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	bot := botinit.NewDummyBot(logger)

	scrapperClient, scrapperConn, err := newScrapperGRPCClient(logger, config.ScrapperURL.HostPort())
	if err != nil {
		return fmt.Errorf("create scrapper grpc client: %w", err)
	}
	defer func() {
		if closeErr := scrapperConn.Close(); closeErr != nil {
			logger.Error("failed to close scrapper grpc connection", "error", closeErr)
		}
	}()

	botService := app.NewApp(logger, config, scrapperClient, bot)

	var listenerConfig net.ListenConfig
	grpcListener, err := listenerConfig.Listen(ctx, "tcp", fmt.Sprintf(":%d", config.GRPCPort))
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}

	grpcSrv := grpc.NewServer()
	go func() {
		<-ctx.Done()
		logger.Info("stopping bot_dummy grpc server", "error", ctx.Err())
		grpcSrv.GracefulStop()
	}()

	botGRPCServer := grpcbot.NewBotServiceServer(logger, botService)
	pb.RegisterBotServiceServer(grpcSrv, botGRPCServer)

	go func() {
		if serveErr := grpcSrv.Serve(grpcListener); serveErr != nil {
			logger.Error("grpc server died", "error", serveErr)
			stop()
		}
	}()

	if err = botService.Run(ctx); err != nil {
		return fmt.Errorf("run bot service: %w", err)
	}
	return nil
}

func newScrapperGRPCClient(logger *logs.Logger, addr string) (botinterfaces.Repository, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create grpc client connection: %w", err)
	}

	client := scrappergrpc.NewGRPCScrapperClient(logger, conn)
	return client, conn, nil
}
