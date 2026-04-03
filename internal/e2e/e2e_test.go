package e2e

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

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

	botContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "linktracker-bot-e2e:local",
			ExposedPorts: []string{"9090/tcp"},
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
			WaitingFor: wait.ForListeningPort("9090/tcp").WithStartupTimeout(60 * time.Second),
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
			ExposedPorts: []string{"9080/tcp"},
			Env: map[string]string{
				"APP_SCRAPPER_CHECK_INTERVAL": "1s",
				"APP_SCRAPPER_BASE_URL":       "http://scrapper:9080",
				"APP_BOT_BASE_URL":            "http://bot:9090",
			},
			Networks: []string{nw.Name},
			NetworkAliases: map[string][]string{
				nw.Name: {"scrapper"},
			},
			WaitingFor: wait.ForListeningPort("9080/tcp").WithStartupTimeout(60 * time.Second),
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
	port, err := scrapperContainer.MappedPort(ctx, "9080/tcp")
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
