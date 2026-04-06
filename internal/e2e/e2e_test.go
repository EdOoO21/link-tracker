package e2e

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
	pb "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/proto/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestBotAndScrapperE2EOverGRPC(t *testing.T) {
	if os.Getenv("RUN_TESTCONTAINERS") != "1" {
		t.Skip("set RUN_TESTCONTAINERS=1 to run container e2e tests")
	}

	ctx := context.Background()
	repoRoot := filepath.Join("..", "..")

	dockerConfigDir := t.TempDir()
	configPath := filepath.Join(dockerConfigDir, "config.json")
	if err := os.WriteFile(configPath, []byte(`{}`), 0o644); err != nil {
		t.Fatalf("write docker config: %v", err)
	}
	t.Setenv("DOCKER_CONFIG", dockerConfigDir)

	buildImage(t, repoRoot, "linktracker-bot-e2e:local", "bot_testmode")
	buildImage(t, repoRoot, "linktracker-scrapper-e2e:local", "scrapper_testmode")

	nw, err := network.New(ctx)
	if err != nil {
		t.Fatalf("create network: %v", err)
	}
	defer func() {
		_ = nw.Remove(ctx)
	}()

	const (
		postgresImage        = "postgres:16-alpine"
		postgresDBName       = "linktracker_e2e"
		postgresUser         = "postgres"
		postgresPassword     = "postgres"
		postgresAlias        = "postgres"
		postgresInternalPort = "5432"
		botPort              = "9090/tcp"
		scrapperPort         = "9080/tcp"
		startupTimeout       = 60 * time.Second
	)

	postgresContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        postgresImage,
			ExposedPorts: []string{postgresInternalPort + "/tcp"},
			Env: map[string]string{
				"POSTGRES_DB":       postgresDBName,
				"POSTGRES_USER":     postgresUser,
				"POSTGRES_PASSWORD": postgresPassword,
			},
			Networks: []string{nw.Name},
			NetworkAliases: map[string][]string{
				nw.Name: {postgresAlias},
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(startupTimeout),
		},
	})
	if err != nil {
		t.Fatalf("start postgres container: %v", err)
	}
	defer func() {
		_ = postgresContainer.Terminate(ctx)
	}()

	applyPostgresMigration(ctx, t, postgresContainer, postgresDBName, postgresUser, postgresPassword)

	botContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "linktracker-bot-e2e:local",
			ExposedPorts: []string{botPort},
			Env: map[string]string{
				"APP_TELEGRAM_TOKEN":         "dummy-token",
				"APP_TELEGRAM_COMMANDS_PATH": "/app/commands.json",
				"APP_SCRAPPER_BASE_URL":      "http://scrapper:9080",
				"APP_BOT_BASE_URL":           "http://bot:9090",
			},
			Networks: []string{nw.Name},
			NetworkAliases: map[string][]string{
				nw.Name: {"bot"},
			},
			WaitingFor: wait.ForListeningPort(botPort).WithStartupTimeout(startupTimeout),
		},
	})
	if err != nil {
		t.Fatalf("start bot container: %v", err)
	}
	defer func() {
		_ = botContainer.Terminate(ctx)
	}()

	scrapperContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "linktracker-scrapper-e2e:local",
			ExposedPorts: []string{scrapperPort},
			Env: map[string]string{
				"APP_SCRAPPER_CHECK_INTERVAL": "1s",
				"APP_SCRAPPER_BASE_URL":       "http://scrapper:9080",
				"APP_BOT_BASE_URL":            "http://bot:9090",
				"APP_DATABASE_URL":            fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", postgresUser, postgresPassword, postgresAlias, postgresInternalPort, postgresDBName),
				"APP_DATABASE_USER":           postgresUser,
				"APP_DATABASE_PASSWORD":       postgresPassword,
				"APP_DATABASE_ACCESS_TYPE":    "sql",
			},
			Networks: []string{nw.Name},
			NetworkAliases: map[string][]string{
				nw.Name: {"scrapper"},
			},
			WaitingFor: wait.ForListeningPort(scrapperPort).WithStartupTimeout(startupTimeout),
		},
	})
	if err != nil {
		t.Fatalf("start scrapper container: %v", err)
	}
	defer func() {
		_ = scrapperContainer.Terminate(ctx)
	}()

	host, err := scrapperContainer.Host(ctx)
	if err != nil {
		t.Fatalf("scrapper host: %v", err)
	}
	port, err := scrapperContainer.MappedPort(ctx, scrapperPort)
	if err != nil {
		t.Fatalf("scrapper mapped port: %v", err)
	}

	conn, err := grpc.NewClient(
		fmt.Sprintf("%s:%s", host, port.Port()),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("grpc client conn: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	scrapperClient := pb.NewScrapperServiceClient(conn)

	if _, addChatErr := scrapperClient.AddChat(ctx, &pb.ChatRequest{ChatId: 1}); addChatErr != nil {
		t.Fatalf("add chat: %v", addChatErr)
	}
	if _, addLinkErr := scrapperClient.AddLink(ctx, &pb.AddLinkRequest{
		ChatId: 1,
		Url:    "https://github.com/user/repo",
		Tags:   []string{"go"},
	}); addLinkErr != nil {
		t.Fatalf("add link: %v", addLinkErr)
	}

	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		logs, logsErr := botContainer.Logs(ctx)
		if logsErr == nil {
			body, readErr := io.ReadAll(logs)
			_ = logs.Close()
			if readErr == nil && strings.Contains(string(body), "update sent") {
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}

	logs, _ := botContainer.Logs(ctx)
	body, _ := io.ReadAll(logs)
	if logs != nil {
		_ = logs.Close()
	}
	t.Fatalf("did not observe update delivery in bot logs:\n%s", string(body))
}

func buildImage(t *testing.T, repoRoot, tag, service string) {
	t.Helper()

	cmd := exec.Command("docker", "build", "-t", tag, "--build-arg", "SERVICE="+service, ".")
	cmd.Dir = repoRoot
	cmd.Env = os.Environ()
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("build image %s: %v\n%s", tag, err, string(output))
	}
}

func applyPostgresMigration(
	ctx context.Context,
	t *testing.T,
	postgresContainer testcontainers.Container,
	databaseName string,
	user string,
	password string,
) {
	t.Helper()

	host, err := postgresContainer.Host(ctx)
	if err != nil {
		t.Fatalf("postgres host: %v", err)
	}
	port, err := postgresContainer.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("postgres mapped port: %v", err)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port.Port(), databaseName)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create postgres pool: %v", err)
	}
	defer pool.Close()

	pingErr := pool.Ping(ctx)
	if pingErr != nil {
		t.Fatalf("ping postgres: %v", pingErr)
	}

	migrationPath := e2eMigrationPath(t)
	script, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration %s: %v", migrationPath, err)
	}

	_, execErr := pool.Exec(ctx, string(script))
	if execErr != nil {
		t.Fatalf("apply migration %s: %v", migrationPath, execErr)
	}
}

func e2eMigrationPath(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve e2e test file path")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "migrations", "V1__init.sql"))
}
