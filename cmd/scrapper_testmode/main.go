package main

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	scrapper "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper"
	scrapperinterfaces "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/interfaces"
	sourcedummy "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/e2e"
	botgrpc "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/clients/bot_grpc"
	grpcscrapper "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/grpc/scrapper"
	logs "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/logger"
	postgresrepo "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/infrastructure/postgres"
	settings "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/settings/scrapper"
	pb "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	logger := logs.NewLogger()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	err := run(ctx, logger)
	stop()

	if err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, logger *logs.Logger) error {

	cfg, err := settings.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	logger.Info("scrapper testmode config loaded", "bot_addr", cfg.BotURL.HostPort(), "grpc_port", cfg.GRPCPort, "batch_size", cfg.BatchSize)

	repository, err := postgresrepo.NewRepository(ctx, logger, cfg.DB)
	if err != nil {
		return fmt.Errorf("create postgres repository: %w", err)
	}
	defer repository.Close()

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

	scrapperService := scrapper.NewScrapper(logger, repository, botClient, githubUpdates, stackOverflowUpdates, cfg.BatchSize)
	interval := loadCheckInterval(logger)
	scrapperService.RunCron(ctx, interval)
	logger.Info("scrapper cron started", "interval", interval)

	var listenerConfig net.ListenConfig
	grpcListener, err := listenerConfig.Listen(ctx, "tcp", fmt.Sprintf(":%d", cfg.GRPCPort))
	if err != nil {
		return fmt.Errorf("listen grpc: %w", err)
	}
	logger.Info("scrapper grpc server starting", "port", cfg.GRPCPort)

	grpcSrv := grpc.NewServer()
	go func() {
		<-ctx.Done()
		logger.Info("stopping scrapper_dummy grpc server", "error", ctx.Err())
		grpcSrv.GracefulStop()
	}()

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
