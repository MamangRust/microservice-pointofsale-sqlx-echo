package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

func setupTestDB(t *testing.T) (*sqlx.DB, *Repositories, *sqlx.DB, func()) {
	t.Helper()

	ctx := context.Background()

	pgContainer, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
	)
	require.NoError(t, err)

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	pool, err := sqlx.Connect("pgx", connStr)
	require.NoError(t, err)

	waitForDatabase(t, ctx, pool)

	runMigrations(t, ctx, connStr, "../../../service/*/database/migration/*.sql")

	queries := pool
	repos := NewRepositories(queries, nil, nil)

	cleanup := func() {
		pool.Close()
		pgContainer.Terminate(ctx)
	}

	return queries, repos, pool, cleanup
}

func runMigrations(t *testing.T, ctx context.Context, dsn string, migrationDir string) {
	t.Helper()

	// Migrations live per-service (service/*/database/migration) in goose
	// format. Collect every *.sql into one directory so goose applies only the
	// +goose Up sections in timestamp order on the shared test database — raw
	// per-file Exec would also run the +goose Down sections and drop what was
	// just created.
	matches, err := filepath.Glob(migrationDir)
	require.NoError(t, err)
	if len(matches) == 0 {
		t.Fatalf("no migration files found matching %s", migrationDir)
	}

	dest := t.TempDir()
	for _, file := range matches {
		data, err := os.ReadFile(file)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(dest, filepath.Base(file)), data, 0o644))
	}

	db, err := goose.OpenDBWithDriver("pgx", dsn)
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, goose.UpContext(ctx, db, dest))
}

// waitForDatabase retries until Postgres answers queries. RunContainer's log-based
// readiness can fire while the entrypoint's temporary init server is still up,
// and every connection during the restart gap is refused with a TCP reset.
func waitForDatabase(t *testing.T, ctx context.Context, pool *sqlx.DB) {
	t.Helper()

	deadline := time.Now().Add(60 * time.Second)
	for {
		var one int
		err := pool.QueryRowContext(ctx, `SELECT 1`).Scan(&one)
		if err == nil {
			return
		}
		if time.Now().After(deadline) {
			t.Fatalf("database not ready after 60s: %v", err)
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func ptr[T any](v T) *T {
	return &v
}

func TestUserCommand_Create(t *testing.T) {
	_, repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	user, err := repos.UserCommand.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName:       "Alice",
		LastName:        "Johnson",
		Email:           "alice@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	})
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, "Alice", user.Firstname)
	assert.Equal(t, "Johnson", user.Lastname)
	assert.Equal(t, "alice@example.com", user.Email)
	assert.NotZero(t, user.UserID)
}

func TestUserQuery_FindByID(t *testing.T) {
	_, repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, err := repos.UserCommand.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName:       "Bob",
		LastName:        "Smith",
		Email:           "bob@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	})
	require.NoError(t, err)

	found, err := repos.UserQuery.FindById(ctx, int(created.UserID))
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.UserID, found.UserID)
	assert.Equal(t, "Bob", found.Firstname)
}

func TestUserQuery_FindByEmail(t *testing.T) {
	_, repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	_, err := repos.UserCommand.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName:       "Email",
		LastName:        "Test",
		Email:           "emailtest@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	})
	require.NoError(t, err)

	found, err := repos.UserQuery.FindByEmail(ctx, "emailtest@example.com")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "emailtest@example.com", found.Email)
}

func TestUserCommand_Update(t *testing.T) {
	_, repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, err := repos.UserCommand.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName:       "Update",
		LastName:        "Me",
		Email:           "updateme@example.com",
		Password:        "oldpassword",
		ConfirmPassword: "oldpassword",
	})
	require.NoError(t, err)

	updated, err := repos.UserCommand.UpdateUser(ctx, &requests.UpdateUserRequest{
		UserID:          ptr(int(created.UserID)),
		FirstName:       "Updated",
		LastName:        "User",
		Email:           "updated@example.com",
		Password:        "newpassword",
		ConfirmPassword: "newpassword",
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "Updated", updated.Firstname)
	assert.Equal(t, "updated@example.com", updated.Email)
}

func TestUserQuery_FindAll(t *testing.T) {
	_, repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	_, err := repos.UserCommand.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "User1", LastName: "Test1", Email: "user1@example.com", Password: "password123", ConfirmPassword: "password123",
	})
	require.NoError(t, err)
	_, err = repos.UserCommand.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "User2", LastName: "Test2", Email: "user2@example.com", Password: "password123", ConfirmPassword: "password123",
	})
	require.NoError(t, err)

	results, total, err := repos.UserQuery.FindAllUsers(ctx, &requests.FindAllUsers{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 2)
	assert.GreaterOrEqual(t, len(results), 2)
}

func TestUserCommand_Trash(t *testing.T) {
	_, repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, err := repos.UserCommand.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "Trash", LastName: "User", Email: "trash@example.com", Password: "password123", ConfirmPassword: "password123",
	})
	require.NoError(t, err)

	trashed, err := repos.UserCommand.TrashedUser(ctx, int(created.UserID))
	require.NoError(t, err)
	require.NotNil(t, trashed)
	assert.True(t, trashed.DeletedAt.Valid)
}

func TestUserQuery_FindByTrashed(t *testing.T) {
	_, repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, err := repos.UserCommand.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "TrashFind", LastName: "User", Email: "trashfind@example.com", Password: "password123", ConfirmPassword: "password123",
	})
	require.NoError(t, err)

	_, err = repos.UserCommand.TrashedUser(ctx, int(created.UserID))
	require.NoError(t, err)

	results, total, err := repos.UserQuery.FindByTrashed(ctx, &requests.FindAllUsers{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 1)
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestUserCommand_Restore(t *testing.T) {
	_, repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, err := repos.UserCommand.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "Restore", LastName: "User", Email: "restore@example.com", Password: "password123", ConfirmPassword: "password123",
	})
	require.NoError(t, err)

	_, err = repos.UserCommand.TrashedUser(ctx, int(created.UserID))
	require.NoError(t, err)

	restored, err := repos.UserCommand.RestoreUser(ctx, int(created.UserID))
	require.NoError(t, err)
	require.NotNil(t, restored)
	assert.False(t, restored.DeletedAt.Valid)
}

func TestUserQuery_FindByActive(t *testing.T) {
	_, repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, err := repos.UserCommand.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "Active", LastName: "User", Email: "active@example.com", Password: "password123", ConfirmPassword: "password123",
	})
	require.NoError(t, err)

	results, total, err := repos.UserQuery.FindByActive(ctx, &requests.FindAllUsers{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 1)

	found := false
	for _, r := range results {
		if int(r.UserID) == int(created.UserID) {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestUserCommand_DeletePermanent(t *testing.T) {
	_, repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, err := repos.UserCommand.CreateUser(ctx, &requests.CreateUserRequest{
		FirstName: "PermDelete", LastName: "User", Email: "permdelete@example.com", Password: "password123", ConfirmPassword: "password123",
	})
	require.NoError(t, err)

	_, err = repos.UserCommand.TrashedUser(ctx, int(created.UserID))
	require.NoError(t, err)

	deleted, err := repos.UserCommand.DeleteUserPermanent(ctx, int(created.UserID))
	require.NoError(t, err)
	assert.True(t, deleted)

	_, err = repos.UserQuery.FindById(ctx, int(created.UserID))
	assert.Error(t, err)
}
