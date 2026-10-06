package tests

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	goredis "github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"net"
	"os"
	"path/filepath"
)

type TestSuite struct {
	PGContainer    *postgres.PostgresContainer
	RedisContainer *redis.RedisContainer
	// CHContainer is started lazily by stats suites (see stats_env.go), so
	// non-stats suites never pay for a ClickHouse container.
	CHContainer testcontainers.Container
	CHAddr      string
	DBURL       string
	RedisURL    string
	Ctx         context.Context
}

func SetupTestSuite() (*TestSuite, error) {
	ctx := context.Background()

	// Setup PostgreSQL
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start postgres container: %w", err)
	}

	dbURL, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to get postgres connection string: %w", err)
	}

	// Setup Redis
	redisContainer, err := redis.Run(ctx, "redis:7-alpine")
	if err != nil {
		return nil, fmt.Errorf("failed to start redis container: %w", err)
	}

	redisURL, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get redis connection string: %w", err)
	}

	ts := &TestSuite{
		PGContainer:    pgContainer,
		RedisContainer: redisContainer,
		DBURL:          dbURL,
		RedisURL:       redisURL,
		Ctx:            ctx,
	}

	// Find project root
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get cwd: %w", err)
	}

	root := cwd
	for {
		if _, err := os.Stat(filepath.Join(root, "justfile")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			return nil, fmt.Errorf("could not find justfile in any parent directory")
		}
		root = parent
	}

	// F1: the consolidated pkg/database/migrations was removed; the single source
	// of truth is now the per-service migrations (service/*/database/migration). The test
	// DB is a single shared instance, so collect every per-service migration file
	// into a temp dir and run them in timestamp order (the file-name prefix
	// preserves the global chronology).
	migrationsDir, err := os.MkdirTemp("", "pos-test-migrations-")
	if err != nil {
		ts.Teardown()
		return nil, fmt.Errorf("failed to create temp migrations dir: %w", err)
	}
	defer os.RemoveAll(migrationsDir)
	if err := collectServiceMigrations(root, migrationsDir); err != nil {
		ts.Teardown()
		return nil, fmt.Errorf("failed to collect per-service migrations: %w", err)
	}
	if err := ts.RunMigrations(migrationsDir); err != nil {
		ts.Teardown()
		return nil, fmt.Errorf("failed to run migrations in %s: %w", migrationsDir, err)
	}

	if err := ts.SeedMinimalRoles(); err != nil {
		ts.Teardown()
		return nil, fmt.Errorf("failed to seed minimal roles: %w", err)
	}

	return ts, nil
}

// SeedMinimalRoles inserts the default roles required by the auth service
// (register assigns ROLE_ADMIN by default). Migration only creates the
// schema — fixture roles must exist for integration tests.
func (ts *TestSuite) SeedMinimalRoles() error {
	pool := ts.SQLxDB()
	defer pool.Close()

	_, err := pool.ExecContext(ts.Ctx, `
		INSERT INTO roles (role_name)
		VALUES ('ROLE_ADMIN'), ('ROLE_CASHIER')
		ON CONFLICT (role_name) DO NOTHING
	`)
	return err
}

// collectServiceMigrations copies every service/*/database/migration/*.sql file
// into dest, preserving the original file names so goose orders them by their
// timestamp prefix. Identical names across services overwrite each other with
// the same content, which is safe because the definitions are identical.
func collectServiceMigrations(root, dest string) error {
	matches, err := filepath.Glob(filepath.Join(root, "service", "*", "database", "migration", "*.sql"))
	if err != nil {
		return err
	}
	if len(matches) == 0 {
		return fmt.Errorf("no per-service migration files found under %s", filepath.Join(root, "service"))
	}
	for _, m := range matches {
		data, err := os.ReadFile(m)
		if err != nil {
			return fmt.Errorf("read %s: %w", m, err)
		}
		if err := os.WriteFile(filepath.Join(dest, filepath.Base(m)), data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", filepath.Base(m), err)
		}
	}
	return nil
}

func (ts *TestSuite) RunMigrations(migrationsDir string) error {
	db, err := goose.OpenDBWithDriver("pgx", ts.DBURL)
	if err != nil {
		return fmt.Errorf("failed to open db: %w", err)
	}
	defer db.Close()

	if err := goose.Up(db, migrationsDir); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	return nil
}

func (ts *TestSuite) SQLxDB() *sqlx.DB {
	db, err := sqlx.Connect("pgx", ts.DBURL)
	if err != nil {
		panic(fmt.Sprintf("test suite: failed to connect via sqlx: %v", err))
	}
	return db
}

func (ts *TestSuite) RedisClient() *goredis.Client {
	opts, _ := goredis.ParseURL(ts.RedisURL)
	return goredis.NewClient(opts)
}

func (ts *TestSuite) Teardown() {
	if ts.PGContainer != nil {
		if err := ts.PGContainer.Terminate(ts.Ctx); err != nil {
			log.Printf("failed to terminate postgres container: %v", err)
		}
	}
	if ts.RedisContainer != nil {
		if err := ts.RedisContainer.Terminate(ts.Ctx); err != nil {
			log.Printf("failed to terminate redis container: %v", err)
		}
	}
	if ts.CHContainer != nil {
		if err := ts.CHContainer.Terminate(ts.Ctx); err != nil {
			log.Printf("failed to terminate clickhouse container: %v", err)
		}
	}
}

type GRPCServerRunner interface {
	Serve(lis net.Listener) error
	Stop()
	GracefulStop()
}

func RunGRPCServer(server *grpc.Server) (string, error) {
	lis, err := net.Listen("tcp", "localhost:0")
	if err != nil {
		return "", err
	}
	go func() {
		if err := server.Serve(lis); err != nil {
			log.Printf("grpc server error: %v", err)
		}
	}()
	return lis.Addr().String(), nil
}
