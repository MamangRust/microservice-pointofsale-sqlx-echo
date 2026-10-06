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
	repos := NewRepositories(queries)

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

func TestRoleCommand_Create(t *testing.T) {
	_, repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	role, err := repos.RoleCommand.CreateRole(ctx, &requests.CreateRoleRequest{
		Name: "admin",
	})
	require.NoError(t, err)
	require.NotNil(t, role)
	assert.Equal(t, "admin", role.RoleName)
	assert.NotZero(t, role.RoleID)
}

func TestRoleQuery_FindByID(t *testing.T) {
	_, repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, err := repos.RoleCommand.CreateRole(ctx, &requests.CreateRoleRequest{
		Name: "editor",
	})
	require.NoError(t, err)

	found, err := repos.RoleQuery.FindById(ctx, int(created.RoleID))
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.RoleID, found.RoleID)
	assert.Equal(t, "editor", found.RoleName)
}

func TestRoleQuery_FindByName(t *testing.T) {
	_, repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	_, err := repos.RoleCommand.CreateRole(ctx, &requests.CreateRoleRequest{
		Name: "viewer",
	})
	require.NoError(t, err)

	found, err := repos.RoleQuery.FindByName(ctx, "viewer")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "viewer", found.RoleName)
}

func TestRoleCommand_Update(t *testing.T) {
	_, repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, err := repos.RoleCommand.CreateRole(ctx, &requests.CreateRoleRequest{
		Name: "oldrole",
	})
	require.NoError(t, err)

	updated, err := repos.RoleCommand.UpdateRole(ctx, &requests.UpdateRoleRequest{
		ID:   ptr(int(created.RoleID)),
		Name: "newrole",
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "newrole", updated.RoleName)
}

func TestRoleQuery_FindAll(t *testing.T) {
	_, repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	_, err := repos.RoleCommand.CreateRole(ctx, &requests.CreateRoleRequest{Name: "role1"})
	require.NoError(t, err)
	_, err = repos.RoleCommand.CreateRole(ctx, &requests.CreateRoleRequest{Name: "role2"})
	require.NoError(t, err)

	results, total, err := repos.RoleQuery.FindAllRoles(ctx, &requests.FindAllRoles{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 2)
	assert.GreaterOrEqual(t, len(results), 2)
}

func TestRoleCommand_Trash(t *testing.T) {
	_, repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, err := repos.RoleCommand.CreateRole(ctx, &requests.CreateRoleRequest{Name: "temporary"})
	require.NoError(t, err)

	trashed, err := repos.RoleCommand.TrashedRole(ctx, int(created.RoleID))
	require.NoError(t, err)
	require.NotNil(t, trashed)
	assert.True(t, trashed.DeletedAt.Valid)
}

func TestRoleQuery_FindByTrashed(t *testing.T) {
	_, repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, err := repos.RoleCommand.CreateRole(ctx, &requests.CreateRoleRequest{Name: "trashrole"})
	require.NoError(t, err)

	_, err = repos.RoleCommand.TrashedRole(ctx, int(created.RoleID))
	require.NoError(t, err)

	results, total, err := repos.RoleQuery.FindByTrashedRole(ctx, &requests.FindAllRoles{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 1)
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestRoleCommand_Restore(t *testing.T) {
	_, repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, err := repos.RoleCommand.CreateRole(ctx, &requests.CreateRoleRequest{Name: "restorerole"})
	require.NoError(t, err)

	_, err = repos.RoleCommand.TrashedRole(ctx, int(created.RoleID))
	require.NoError(t, err)

	restored, err := repos.RoleCommand.RestoreRole(ctx, int(created.RoleID))
	require.NoError(t, err)
	require.NotNil(t, restored)
	assert.False(t, restored.DeletedAt.Valid)
}

func TestRoleQuery_FindByActive(t *testing.T) {
	_, repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, err := repos.RoleCommand.CreateRole(ctx, &requests.CreateRoleRequest{Name: "activerole"})
	require.NoError(t, err)

	results, total, err := repos.RoleQuery.FindByActiveRole(ctx, &requests.FindAllRoles{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 1)

	found := false
	for _, r := range results {
		if r.RoleID == created.RoleID {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestRoleCommand_DeletePermanent(t *testing.T) {
	_, repos, _, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	created, err := repos.RoleCommand.CreateRole(ctx, &requests.CreateRoleRequest{Name: "deletepermanent"})
	require.NoError(t, err)

	_, err = repos.RoleCommand.TrashedRole(ctx, int(created.RoleID))
	require.NoError(t, err)

	deleted, err := repos.RoleCommand.DeleteRolePermanent(ctx, int(created.RoleID))
	require.NoError(t, err)
	assert.True(t, deleted)

	_, err = repos.RoleQuery.FindById(ctx, int(created.RoleID))
	assert.Error(t, err)
}

func TestRoleQuery_FindByUserId(t *testing.T) {
	_, repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	// Seed a user and a role, then assign the role
	var userID int
	err := pool.QueryRowContext(ctx, `INSERT INTO users (firstname, lastname, email, password, verification_code)
		VALUES ('Test', 'User', 'testuser@example.com', 'password', 'CODE') RETURNING user_id`).Scan(&userID)
	require.NoError(t, err)

	created, err := repos.RoleCommand.CreateRole(ctx, &requests.CreateRoleRequest{Name: "userole"})
	require.NoError(t, err)

	_, err = pool.ExecContext(ctx, `INSERT INTO user_roles (user_id, role_id) VALUES ($1, $2)`, userID, created.RoleID)
	require.NoError(t, err)

	roles, err := repos.RoleQuery.FindByUserId(ctx, userID)
	require.NoError(t, err)
	require.NotNil(t, roles)
	assert.GreaterOrEqual(t, len(roles), 1)
	assert.Equal(t, "userole", roles[0].RoleName)
}
