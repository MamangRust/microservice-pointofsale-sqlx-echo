package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
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

	repos := NewRepositories(pool)

	cleanup := func() {
		pool.Close()
		pgContainer.Terminate(ctx)
	}

	return pool, repos, pool, cleanup
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

func seedOrderItemData(t *testing.T, ctx context.Context, pool *sqlx.DB) (int, int) {
	t.Helper()

	// Seed user
	var userID int
	err := pool.QueryRowContext(ctx, `INSERT INTO users (firstname, lastname, email, password, verification_code)
		VALUES ('Order', 'User', 'orderuser@example.com', 'password', 'CODE') RETURNING user_id`).Scan(&userID)
	require.NoError(t, err)

	// Seed merchant
	var merchantID int
	err = pool.QueryRowContext(ctx, `INSERT INTO merchants (user_id, name, status) VALUES ($1, 'Test Merchant', 'active') RETURNING merchant_id`, userID).Scan(&merchantID)
	require.NoError(t, err)

	// Seed category
	var categoryID int
	err = pool.QueryRowContext(ctx, `INSERT INTO categories (name) VALUES ('Test Category') RETURNING category_id`).Scan(&categoryID)
	require.NoError(t, err)

	// Seed product
	var productID int
	err = pool.QueryRowContext(ctx, `INSERT INTO products (merchant_id, category_id, name, price, count_in_stock)
		VALUES ($1, $2, 'Test Product', 10000, 50) RETURNING product_id`, merchantID, categoryID).Scan(&productID)
	require.NoError(t, err)

	// Seed cashier
	var cashierID int
	err = pool.QueryRowContext(ctx, `INSERT INTO cashiers (merchant_id, user_id, name)
		VALUES ($1, $2, 'Test Cashier') RETURNING cashier_id`, merchantID, userID).Scan(&cashierID)
	require.NoError(t, err)

	// Seed order
	var orderID int
	err = pool.QueryRowContext(ctx, `INSERT INTO orders (merchant_id, cashier_id, total_price)
		VALUES ($1, $2, 10000) RETURNING order_id`, merchantID, cashierID).Scan(&orderID)
	require.NoError(t, err)

	// Seed order item
	_, err = pool.ExecContext(ctx, `INSERT INTO order_items (order_id, product_id, quantity, price) VALUES ($1, $2, 2, 10000)`, orderID, productID)
	require.NoError(t, err)

	return orderID, productID
}

func TestOrderItemQuery_FindAll(t *testing.T) {
	_, repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	seedOrderItemData(t, ctx, pool)

	results, total, err := repos.OrderItemQuery.FindAllOrderItems(ctx, &requests.FindAllOrderItems{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 1)
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestOrderItemQuery_FindByActive(t *testing.T) {
	_, repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	seedOrderItemData(t, ctx, pool)

	results, total, err := repos.OrderItemQuery.FindByActive(ctx, &requests.FindAllOrderItems{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 1)
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestOrderItemQuery_FindByTrashed(t *testing.T) {
	_, repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	orderID, _ := seedOrderItemData(t, ctx, pool)

	// First trash the order item
	_, err := pool.ExecContext(ctx, `UPDATE order_items SET deleted_at = NOW() WHERE order_id = $1`, orderID)
	require.NoError(t, err)

	results, total, err := repos.OrderItemQuery.FindByTrashed(ctx, &requests.FindAllOrderItems{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 1)
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestOrderItemQuery_FindOrderItemByOrder(t *testing.T) {
	_, repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	orderID, _ := seedOrderItemData(t, ctx, pool)

	items, err := repos.OrderItemQuery.FindOrderItemByOrder(ctx, orderID)
	require.NoError(t, err)
	require.NotNil(t, items)
	assert.GreaterOrEqual(t, len(items), 1)
	assert.Equal(t, int32(orderID), items[0].OrderID)
}
