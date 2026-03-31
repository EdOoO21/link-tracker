package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	scrapper "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper"
	scrapperinterfaces "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/interfaces"
	botgrpc "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/clients/bot_grpc"
	sourcedummy "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/clients/source_dummy"
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
	if err := run(logger); err != nil {
		logger.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run(logger *logs.Logger) error {
	repository := repo.NewRepo()
	cfg, err := settings.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	logger.Info("scrapper testmode config loaded", "bot_addr", cfg.BotURL.HostPort(), "grpc_port", cfg.GRPCPort)

	githubUpdates := sourcedummy.NewGitHubClient()
	stackOverflowUpdates := sourcedummy.NewStackOverflowClient()
	logger.Info("scrapper dummy source clients initialized")
	botClient, botConn, err := newBotGRPCClient(logger, cfg.BotURL.HostPort())
	if err != nil {
		return fmt.Errorf("create bot grpc client: %w", err)
	}
	logger.Info("bot grpc client created", "target", cfg.BotURL.HostPort())
	defer func() {
		if closeErr := botConn.Close(); closeErr != nil {
			logger.Error("failed to close bot grpc connection", "error", closeErr)
		}
	}()

	scrapperService := scrapper.NewScrapper(logger, repository, botClient, githubUpdates, stackOverflowUpdates)
	interval := loadCheckInterval(logger)
	scrapperService.RunCron(interval)
	logger.Info("scrapper cron started", "interval", interval)

	var listenerConfig net.ListenConfig
	grpcListener, err := listenerConfig.Listen(context.Background(), "tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}
	logger.Info("scrapper grpc server starting", "port", cfg.GRPCPort)
	defer func() {
		if closeErr := grpcListener.Close(); closeErr != nil {
			logger.Error("failed to close grpc listener", "error", closeErr)
		}
	}()

	grpcSrv := grpc.NewServer()
	defer grpcSrv.GracefulStop()

	scrapperGRPCServer := grpcscrapper.NewScrapperServiceServer(logger, scrapperService)
	pb.RegisterScrapperServiceServer(grpcSrv, scrapperGRPCServer)

	if serveErr := grpcSrv.Serve(grpcListener); serveErr != nil {
		return fmt.Errorf("serve grpc: %w", serveErr)
	}
	return nil
}

func loadCheckInterval(logger *logs.Logger) time.Duration {
	raw := os.Getenv("APP_SCRAPPER_CHECK_INTERVAL")
	if raw == "" {
		return time.Minute
	}

	duration, err := time.ParseDuration(raw)
	if err != nil {
		logger.Warn("invalid scrapper check interval, using default", "raw", raw, "error", err)
		return time.Minute
	}
	return duration
}

func newBotGRPCClient(logger *logs.Logger, addr string) (scrapperinterfaces.BotClient, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create grpc client connection: %w", err)
	}

	client := botgrpc.NewGRPCBotClient(logger, conn)
	return client, conn, nil
}
