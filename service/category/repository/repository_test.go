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

func setupTestDB(t *testing.T) (*Repositories, *sqlx.DB, func()) {
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

	repos := NewRepositories(pool)

	cleanup := func() {
		pool.Close()
		pgContainer.Terminate(ctx)
	}

	return repos, pool, cleanup
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

func TestCategoryCommand_Create(t *testing.T) {
	repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	req := &requests.CreateCategoryRequest{
		Name:        "Electronics",
		Description: "Electronic items and gadgets",
	}

	category, err := repos.CategoryCommand.CreateCategory(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, category)
	assert.Equal(t, "Electronics", category.Name)
	assert.Equal(t, "Electronic items and gadgets", *category.Description)
	assert.NotZero(t, category.CategoryID)
}

func TestCategoryQuery_FindByID(t *testing.T) {
	repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, err := repos.CategoryCommand.CreateCategory(ctx, &requests.CreateCategoryRequest{
		Name:        "Books",
		Description: "All kinds of books",
	})
	require.NoError(t, err)

	found, err := repos.CategoryQuery.FindById(ctx, int(created.CategoryID))
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.CategoryID, found.CategoryID)
	assert.Equal(t, "Books", found.Name)
}

func TestCategoryCommand_Update(t *testing.T) {
	repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, err := repos.CategoryCommand.CreateCategory(ctx, &requests.CreateCategoryRequest{
		Name:        "OldName",
		Description: "Old description",
	})
	require.NoError(t, err)

	updated, err := repos.CategoryCommand.UpdateCategory(ctx, &requests.UpdateCategoryRequest{
		CategoryID:  ptr(int(created.CategoryID)),
		Name:        "NewName",
		Description: "New description",
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "NewName", updated.Name)
	assert.Equal(t, "New description", *updated.Description)
}

func TestCategoryQuery_FindAll(t *testing.T) {
	repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	_, err := repos.CategoryCommand.CreateCategory(ctx, &requests.CreateCategoryRequest{
		Name:        "Cat1",
		Description: "Desc1",
	})
	require.NoError(t, err)

	_, err = repos.CategoryCommand.CreateCategory(ctx, &requests.CreateCategoryRequest{
		Name:        "Cat2",
		Description: "Desc2",
	})
	require.NoError(t, err)

	results, total, err := repos.CategoryQuery.FindAllCategory(ctx, &requests.FindAllCategory{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 2)
	assert.GreaterOrEqual(t, len(results), 2)
}

func TestCategoryCommand_Trash(t *testing.T) {
	repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, err := repos.CategoryCommand.CreateCategory(ctx, &requests.CreateCategoryRequest{
		Name:        "TrashMe",
		Description: "Will be trashed",
	})
	require.NoError(t, err)

	trashed, err := repos.CategoryCommand.TrashedCategory(ctx, int(created.CategoryID))
	require.NoError(t, err)
	require.NotNil(t, trashed)
	assert.True(t, trashed.DeletedAt.Valid)
}

func TestCategoryQuery_FindByTrashed(t *testing.T) {
	repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, err := repos.CategoryCommand.CreateCategory(ctx, &requests.CreateCategoryRequest{
		Name:        "TrashFind",
		Description: "Find in trash",
	})
	require.NoError(t, err)

	_, err = repos.CategoryCommand.TrashedCategory(ctx, int(created.CategoryID))
	require.NoError(t, err)

	results, total, err := repos.CategoryQuery.FindByTrashed(ctx, &requests.FindAllCategory{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 1)
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestCategoryCommand_Restore(t *testing.T) {
	repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, err := repos.CategoryCommand.CreateCategory(ctx, &requests.CreateCategoryRequest{
		Name:        "RestoreMe",
		Description: "Will be restored",
	})
	require.NoError(t, err)

	_, err = repos.CategoryCommand.TrashedCategory(ctx, int(created.CategoryID))
	require.NoError(t, err)

	restored, err := repos.CategoryCommand.RestoreCategory(ctx, int(created.CategoryID))
	require.NoError(t, err)
	require.NotNil(t, restored)
	assert.False(t, restored.DeletedAt.Valid)
}

func TestCategoryQuery_FindByActive(t *testing.T) {
	repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, err := repos.CategoryCommand.CreateCategory(ctx, &requests.CreateCategoryRequest{
		Name:        "ActiveCat",
		Description: "Active category",
	})
	require.NoError(t, err)

	results, total, err := repos.CategoryQuery.FindByActive(ctx, &requests.FindAllCategory{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 1)

	found := false
	for _, r := range results {
		if int(r.CategoryID) == int(created.CategoryID) {
			found = true
			break
		}
	}
	assert.True(t, found, "created category should be in active results")
}

func TestCategoryCommand_DeletePermanent(t *testing.T) {
	repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, err := repos.CategoryCommand.CreateCategory(ctx, &requests.CreateCategoryRequest{
		Name:        "PermanentDelete",
		Description: "Will be permanently deleted",
	})
	require.NoError(t, err)

	_, err = repos.CategoryCommand.TrashedCategory(ctx, int(created.CategoryID))
	require.NoError(t, err)

	deleted, err := repos.CategoryCommand.DeleteCategoryPermanently(ctx, int(created.CategoryID))
	require.NoError(t, err)
	assert.True(t, deleted)

	_, err = repos.CategoryQuery.FindById(ctx, int(created.CategoryID))
	assert.Error(t, err)
}
