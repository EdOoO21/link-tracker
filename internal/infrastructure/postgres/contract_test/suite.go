package contracttest

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	scrapperinterfaces "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/interfaces"
	models "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/application/scrapper/models"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/domain"
	"gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/ports"
	settings "gitlab.education.tbank.ru/backend-academy-go-2025/homeworks/link-tracker/internal/settings/scrapper"
)

type Repository interface {
	scrapperinterfaces.Repository
	Close()
}

type Constructor func(context.Context, ports.Logger, settings.DBConfig) (Repository, error)

type testEnv struct {
	cfg         settings.DBConfig
	adminPool   *pgxpool.Pool
	constructor Constructor
	container   testcontainers.Container
}

type noopLogger struct{}

func (noopLogger) Info(string, ...any)  {}
func (noopLogger) Warn(string, ...any)  {}
func (noopLogger) Error(string, ...any) {}

const (
	defaultChatID         int64 = 101
	firstChatID           int64 = 1
	secondChatID          int64 = 2
	missingLinkID         int64 = 999_999
	trackedLinksBatchSize       = 100
	dbReadyLogCount             = 2
	expectedTwoLinks            = 2
	postgresStartTimeout        = 60 * time.Second
)

func RunRepositorySuite(t *testing.T, constructor Constructor) {
	t.Helper()

	if os.Getenv("RUN_TESTCONTAINERS") != "1" {
		t.Skip("set RUN_TESTCONTAINERS=1 to run postgres integration tests")
	}

	ctx := context.Background()
	env := newTestEnv(ctx, t, constructor)
	t.Cleanup(func() {
		env.close(ctx)
	})

	t.Run("chat presence and duplicates", func(t *testing.T) {
		runChatPresenceAndDuplicates(t, env)
	})

	t.Run("track list filter and untrack", func(t *testing.T) {
		runTrackListFilterAndUntrack(t, env)
	})

	t.Run("tag crud", func(t *testing.T) {
		runTagCRUD(t, env)
	})

	t.Run("list all links groups chat ids and updates timestamp", func(t *testing.T) {
		runListAllLinksAndUpdateTimestamp(t, env)
	})

	t.Run("delete chat cleans subscriptions and shared links", func(t *testing.T) {
		runDeleteChatCleanup(t, env)
	})
}

func runChatPresenceAndDuplicates(t *testing.T, env *testEnv) {
	t.Helper()

	env.run(t, func(ctx context.Context, repo Repository) {
		present, err := repo.IsPresent(ctx, defaultChatID)
		if err != nil {
			t.Fatalf("initial IsPresent: %v", err)
		}
		if present {
			t.Fatal("missing chat reported as present")
		}

		addErr := repo.AddChat(ctx, defaultChatID)
		if addErr != nil {
			t.Fatalf("AddChat: %v", addErr)
		}

		present, err = repo.IsPresent(ctx, defaultChatID)
		if err != nil {
			t.Fatalf("IsPresent after add: %v", err)
		}
		if !present {
			t.Fatal("added chat reported as missing")
		}

		err = repo.AddChat(ctx, defaultChatID)
		assertErrorIs(t, err, ports.ErrChatAlreadyExists)

		deleteErr := repo.DeleteChat(ctx, defaultChatID)
		if deleteErr != nil {
			t.Fatalf("DeleteChat: %v", deleteErr)
		}

		present, err = repo.IsPresent(ctx, defaultChatID)
		if err != nil {
			t.Fatalf("IsPresent after delete: %v", err)
		}
		if present {
			t.Fatal("deleted chat reported as present")
		}

		err = repo.DeleteChat(ctx, defaultChatID)
		assertErrorIs(t, err, ports.ErrChatNotFound)
	})
}

func runTrackListFilterAndUntrack(t *testing.T, env *testEnv) {
	t.Helper()

	env.run(t, func(ctx context.Context, repo Repository) {
		addErr := repo.AddChat(ctx, firstChatID)
		if addErr != nil {
			t.Fatalf("AddChat: %v", addErr)
		}

		trackErr := repo.TrackLink(ctx, firstChatID, "https://github.com/acme/one", []string{"go", "backend"})
		if trackErr != nil {
			t.Fatalf("TrackLink one: %v", trackErr)
		}
		trackErr = repo.TrackLink(ctx, firstChatID, "https://github.com/acme/two", []string{"go"})
		if trackErr != nil {
			t.Fatalf("TrackLink two: %v", trackErr)
		}

		err := repo.TrackLink(ctx, firstChatID, "https://github.com/acme/one", []string{"go"})
		assertErrorIs(t, err, ports.ErrLinkAlreadyExists)

		links, err := repo.ListLinks(ctx, firstChatID, nil)
		if err != nil {
			t.Fatalf("ListLinks without filter: %v", err)
		}
		if len(links) != expectedTwoLinks {
			t.Fatalf("ListLinks without filter len = %d, want %d", len(links), expectedTwoLinks)
		}
		assertHasLinkWithTags(t, links, "https://github.com/acme/one", "backend", "go")
		assertHasLinkWithTags(t, links, "https://github.com/acme/two", "go")

		filteredLinks, err := repo.ListLinks(ctx, firstChatID, []string{"go", "backend"})
		if err != nil {
			t.Fatalf("ListLinks with filter: %v", err)
		}
		if len(filteredLinks) != 1 {
			t.Fatalf("filtered links len = %d, want 1", len(filteredLinks))
		}
		assertHasLinkWithTags(t, filteredLinks, "https://github.com/acme/one", "backend", "go")

		untrackErr := repo.UnTrackLink(ctx, firstChatID, "https://github.com/acme/one")
		if untrackErr != nil {
			t.Fatalf("UnTrackLink: %v", untrackErr)
		}

		err = repo.UnTrackLink(ctx, firstChatID, "https://github.com/acme/one")
		assertErrorIs(t, err, ports.ErrLinkNotFound)

		links, err = repo.ListLinks(ctx, firstChatID, nil)
		if err != nil {
			t.Fatalf("ListLinks after untrack: %v", err)
		}
		if len(links) != 1 {
			t.Fatalf("links after untrack len = %d, want 1", len(links))
		}
		assertHasLinkWithTags(t, links, "https://github.com/acme/two", "go")
	})
}

func runTagCRUD(t *testing.T, env *testEnv) {
	t.Helper()

	env.run(t, func(ctx context.Context, repo Repository) {
		addErr := repo.AddChat(ctx, firstChatID)
		if addErr != nil {
			t.Fatalf("AddChat: %v", addErr)
		}
		trackErr := repo.TrackLink(ctx, firstChatID, "https://github.com/acme/one", []string{"go"})
		if trackErr != nil {
			t.Fatalf("TrackLink: %v", trackErr)
		}

		tags, err := repo.GetTags(ctx, firstChatID, "https://github.com/acme/one")
		if err != nil {
			t.Fatalf("GetTags initial: %v", err)
		}
		assertSliceEqual(t, tags, []string{"go"})

		addTagErr := repo.AddTag(ctx, firstChatID, "https://github.com/acme/one", "news")
		if addTagErr != nil {
			t.Fatalf("AddTag: %v", addTagErr)
		}

		err = repo.AddTag(ctx, firstChatID, "https://github.com/acme/one", "news")
		assertErrorIs(t, err, ports.ErrTagAlreadyExists)

		tags, err = repo.GetTags(ctx, firstChatID, "https://github.com/acme/one")
		if err != nil {
			t.Fatalf("GetTags after add: %v", err)
		}
		assertSliceEqual(t, tags, []string{"go", "news"})

		deleteTagErr := repo.DeleteTag(ctx, firstChatID, "https://github.com/acme/one", "news")
		if deleteTagErr != nil {
			t.Fatalf("DeleteTag: %v", deleteTagErr)
		}

		err = repo.DeleteTag(ctx, firstChatID, "https://github.com/acme/one", "news")
		assertErrorIs(t, err, ports.ErrTagNotFound)

		tags, err = repo.GetTags(ctx, firstChatID, "https://github.com/acme/one")
		if err != nil {
			t.Fatalf("GetTags final: %v", err)
		}
		assertSliceEqual(t, tags, []string{"go"})
	})
}

func runListAllLinksAndUpdateTimestamp(t *testing.T, env *testEnv) {
	t.Helper()

	env.run(t, func(ctx context.Context, repo Repository) {
		addErr := repo.AddChat(ctx, firstChatID)
		if addErr != nil {
			t.Fatalf("AddChat chat1: %v", addErr)
		}
		addErr = repo.AddChat(ctx, secondChatID)
		if addErr != nil {
			t.Fatalf("AddChat chat2: %v", addErr)
		}
		trackErr := repo.TrackLink(ctx, firstChatID, "https://github.com/acme/shared", []string{"go"})
		if trackErr != nil {
			t.Fatalf("TrackLink chat1: %v", trackErr)
		}
		trackErr = repo.TrackLink(ctx, secondChatID, "https://github.com/acme/shared", []string{"infra"})
		if trackErr != nil {
			t.Fatalf("TrackLink chat2: %v", trackErr)
		}

		tracked, err := listAllTrackedLinks(ctx, repo)
		if err != nil {
			t.Fatalf("ListAllLinksBatch initial: %v", err)
		}
		if len(tracked) != 1 {
			t.Fatalf("tracked links len = %d, want 1", len(tracked))
		}
		assertTrackedLink(t, tracked[0], "https://github.com/acme/shared", []int64{firstChatID, secondChatID})

		newLastUpdate := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
		updateErr := repo.UpdateLinksLastUpdate(ctx, tracked[0].LinkID, newLastUpdate)
		if updateErr != nil {
			t.Fatalf("UpdateLinksLastUpdate: %v", updateErr)
		}

		tracked, err = listAllTrackedLinks(ctx, repo)
		if err != nil {
			t.Fatalf("ListAllLinksBatch after update: %v", err)
		}
		if len(tracked) != 1 {
			t.Fatalf("tracked links after update len = %d, want 1", len(tracked))
		}
		if !tracked[0].LastUpdate.Equal(newLastUpdate) {
			t.Fatalf("last update = %s, want %s", tracked[0].LastUpdate, newLastUpdate)
		}

		err = repo.UpdateLinksLastUpdate(ctx, missingLinkID, newLastUpdate)
		assertErrorIs(t, err, ports.ErrLinkNotFound)
	})
}

func runDeleteChatCleanup(t *testing.T, env *testEnv) {
	t.Helper()

	env.run(t, func(ctx context.Context, repo Repository) {
		addErr := repo.AddChat(ctx, firstChatID)
		if addErr != nil {
			t.Fatalf("AddChat chat1: %v", addErr)
		}
		addErr = repo.AddChat(ctx, secondChatID)
		if addErr != nil {
			t.Fatalf("AddChat chat2: %v", addErr)
		}
		trackErr := repo.TrackLink(ctx, firstChatID, "https://github.com/acme/shared", []string{"go"})
		if trackErr != nil {
			t.Fatalf("TrackLink shared chat1: %v", trackErr)
		}
		trackErr = repo.TrackLink(ctx, secondChatID, "https://github.com/acme/shared", []string{"infra"})
		if trackErr != nil {
			t.Fatalf("TrackLink shared chat2: %v", trackErr)
		}
		trackErr = repo.TrackLink(ctx, firstChatID, "https://github.com/acme/private", []string{"private"})
		if trackErr != nil {
			t.Fatalf("TrackLink private chat1: %v", trackErr)
		}

		deleteErr := repo.DeleteChat(ctx, firstChatID)
		if deleteErr != nil {
			t.Fatalf("DeleteChat chat1: %v", deleteErr)
		}

		present, err := repo.IsPresent(ctx, firstChatID)
		if err != nil {
			t.Fatalf("IsPresent chat1 after delete: %v", err)
		}
		if present {
			t.Fatal("deleted chat1 reported as present")
		}

		tracked, err := listAllTrackedLinks(ctx, repo)
		if err != nil {
			t.Fatalf("ListAllLinksBatch after deleting chat1: %v", err)
		}
		if len(tracked) != 1 {
			t.Fatalf("tracked links after deleting chat1 len = %d, want 1", len(tracked))
		}
		assertTrackedLink(t, tracked[0], "https://github.com/acme/shared", []int64{secondChatID})

		deleteErr = repo.DeleteChat(ctx, secondChatID)
		if deleteErr != nil {
			t.Fatalf("DeleteChat chat2: %v", deleteErr)
		}

		tracked, err = listAllTrackedLinks(ctx, repo)
		if err != nil {
			t.Fatalf("ListAllLinksBatch after deleting all chats: %v", err)
		}
		if len(tracked) != 0 {
			t.Fatalf("tracked links after deleting all chats len = %d, want 0", len(tracked))
		}

		err = repo.DeleteChat(ctx, firstChatID)
		assertErrorIs(t, err, ports.ErrChatNotFound)
	})
}

func newTestEnv(ctx context.Context, t *testing.T, constructor Constructor) *testEnv {
	t.Helper()

	const (
		image    = "postgres:16-alpine"
		dbName   = "linktracker_test"
		user     = "postgres"
		password = "postgres"
	)

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image: image,
			ExposedPorts: []string{
				"5432/tcp",
			},
			Env: map[string]string{
				"POSTGRES_DB":       dbName,
				"POSTGRES_USER":     user,
				"POSTGRES_PASSWORD": password,
			},
			WaitingFor: wait.ForLog("database system is ready to accept connections").
				WithOccurrence(dbReadyLogCount).
				WithStartupTimeout(postgresStartTimeout),
		},
	})
	if err != nil {
		t.Fatalf("start postgres test container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("postgres host: %v", err)
	}
	port, err := container.MappedPort(ctx, "5432/tcp")
	if err != nil {
		t.Fatalf("postgres mapped port: %v", err)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port.Port(), dbName)
	adminPool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Fatalf("create postgres admin pool: %v", err)
	}

	pingErr := adminPool.Ping(ctx)
	if pingErr != nil {
		t.Fatalf("ping postgres admin pool: %v", pingErr)
	}

	applyMigrations(ctx, t, adminPool)

	return &testEnv{
		cfg: settings.DBConfig{
			DatabaseURL:      dsn,
			DatabaseUser:     user,
			DatabasePassword: password,
		},
		adminPool:   adminPool,
		constructor: constructor,
		container:   container,
	}
}

func listAllTrackedLinks(ctx context.Context, repo Repository) ([]models.TrackedLink, error) {
	res := make([]models.TrackedLink, 0)
	var afterLinkID int64

	for {
		batch, err := repo.ListAllLinksBatch(ctx, afterLinkID, trackedLinksBatchSize)
		if err != nil {
			return nil, fmt.Errorf("list all links: %w", err)
		}

		if len(batch) == 0 {
			return res, nil
		}

		res = append(res, batch...)
		afterLinkID = batch[len(batch)-1].LinkID
	}
}

func (e *testEnv) close(ctx context.Context) {
	if e.adminPool != nil {
		e.adminPool.Close()
	}
	if e.container != nil {
		_ = e.container.Terminate(ctx)
	}
}

func (e *testEnv) run(t *testing.T, fn func(context.Context, Repository)) {
	t.Helper()

	ctx := context.Background()
	e.reset(ctx, t)

	repo, err := e.constructor(ctx, noopLogger{}, e.cfg)
	if err != nil {
		t.Fatalf("construct repository: %v", err)
	}
	t.Cleanup(repo.Close)

	fn(ctx, repo)
}

func (e *testEnv) reset(ctx context.Context, t *testing.T) {
	t.Helper()

	const query = `
		TRUNCATE TABLE link_tags, chat_links, tags, links, chats RESTART IDENTITY CASCADE
	`

	if _, err := e.adminPool.Exec(ctx, query); err != nil {
		t.Fatalf("reset database: %v", err)
	}
}

func applyMigrations(ctx context.Context, t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	scriptPath := migrationPath(t)
	script, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read migration %s: %v", scriptPath, err)
	}

	_, execErr := pool.Exec(ctx, string(script))
	if execErr != nil {
		t.Fatalf("apply migration %s: %v", scriptPath, execErr)
	}
}

func migrationPath(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve testsuite file path")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", "..", "migrations", "V1__init.sql"))
}

func assertErrorIs(t *testing.T, err, want error) {
	t.Helper()

	if !errors.Is(err, want) {
		t.Fatalf("error = %v, want %v", err, want)
	}
}

func assertHasLinkWithTags(t *testing.T, links []domain.Link, wantURL string, wantTags ...string) {
	t.Helper()

	for _, link := range links {
		if link.URL != wantURL {
			continue
		}
		if len(link.Tags) != len(wantTags) {
			t.Fatalf("link %s tags len = %d, want %d", wantURL, len(link.Tags), len(wantTags))
		}
		for _, tag := range wantTags {
			if _, ok := link.Tags[tag]; !ok {
				t.Fatalf("link %s missing tag %q", wantURL, tag)
			}
		}
		return
	}

	t.Fatalf("link %s not found", wantURL)
}

func assertTrackedLink(t *testing.T, tracked models.TrackedLink, wantURL string, wantChatIDs []int64) {
	t.Helper()

	if tracked.URL != wantURL {
		t.Fatalf("tracked url = %q, want %q", tracked.URL, wantURL)
	}
	if len(tracked.ChatIDs) != len(wantChatIDs) {
		t.Fatalf("tracked chat ids len = %d, want %d", len(tracked.ChatIDs), len(wantChatIDs))
	}
	for i := range wantChatIDs {
		if tracked.ChatIDs[i] != wantChatIDs[i] {
			t.Fatalf("tracked chat ids = %v, want %v", tracked.ChatIDs, wantChatIDs)
		}
	}
}

func assertSliceEqual[T comparable](t *testing.T, got, want []T) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("slice len = %d, want %d; got=%v want=%v", len(got), len(want), got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("slice = %v, want %v", got, want)
		}
	}
}
