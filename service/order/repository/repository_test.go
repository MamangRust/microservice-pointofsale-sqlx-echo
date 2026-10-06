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

	repos := NewRepositories(pool, nil, nil, nil, nil)

	cleanup := func() {
		pool.Close()
		pgContainer.Terminate(ctx)
	}

	return pool, repos, pool, cleanup
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

func ptr[T any](v T) *T {
	return &v
}

func seedOrderDeps(t *testing.T, ctx context.Context, pool *sqlx.DB) (int, int) {
	t.Helper()

	var userID int
	err := pool.QueryRowContext(ctx, `INSERT INTO users (firstname, lastname, email, password, verification_code)
		VALUES ('Order', 'User', 'orderuser@example.com', 'password', 'CODE') RETURNING user_id`).Scan(&userID)
	require.NoError(t, err)

	var merchantID int
	err = pool.QueryRowContext(ctx, `INSERT INTO merchants (user_id, name, status) VALUES ($1, 'Order Merchant', 'active') RETURNING merchant_id`, userID).Scan(&merchantID)
	require.NoError(t, err)

	var cashierID int
	err = pool.QueryRowContext(ctx, `INSERT INTO cashiers (merchant_id, user_id, name) VALUES ($1, $2, 'Order Cashier') RETURNING cashier_id`, merchantID, userID).Scan(&cashierID)
	require.NoError(t, err)

	return merchantID, cashierID
}

func TestOrderCommand_Create(t *testing.T) {
	_, repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, cashierID := seedOrderDeps(t, ctx, pool)

	order, err := repos.OrderCommand.CreateOrder(ctx, &requests.CreateOrderRecordRequest{
		MerchantID: merchantID,
		CashierID:  cashierID,
		TotalPrice: 50000,
	})
	require.NoError(t, err)
	require.NotNil(t, order)
	assert.Equal(t, int32(merchantID), order.MerchantID)
	assert.Equal(t, int32(cashierID), order.CashierID)
	assert.Equal(t, int64(50000), order.TotalPrice)
	assert.NotZero(t, order.OrderID)
}

func TestOrderQuery_FindByID(t *testing.T) {
	_, repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, cashierID := seedOrderDeps(t, ctx, pool)

	created, err := repos.OrderCommand.CreateOrder(ctx, &requests.CreateOrderRecordRequest{
		MerchantID: merchantID, CashierID: cashierID, TotalPrice: 75000,
	})
	require.NoError(t, err)

	found, err := repos.OrderQuery.FindById(ctx, int(created.OrderID))
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.OrderID, found.OrderID)
	assert.Equal(t, int64(75000), found.TotalPrice)
}

func TestOrderCommand_Update(t *testing.T) {
	_, repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, cashierID := seedOrderDeps(t, ctx, pool)

	created, err := repos.OrderCommand.CreateOrder(ctx, &requests.CreateOrderRecordRequest{
		MerchantID: merchantID, CashierID: cashierID, TotalPrice: 30000,
	})
	require.NoError(t, err)

	updated, err := repos.OrderCommand.UpdateOrder(ctx, &requests.UpdateOrderRecordRequest{
		OrderID:    int(created.OrderID),
		TotalPrice: 60000,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, int64(60000), updated.TotalPrice)
}

func TestOrderQuery_FindAll(t *testing.T) {
	_, repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, cashierID := seedOrderDeps(t, ctx, pool)

	_, err := repos.OrderCommand.CreateOrder(ctx, &requests.CreateOrderRecordRequest{
		MerchantID: merchantID, CashierID: cashierID, TotalPrice: 10000,
	})
	require.NoError(t, err)
	_, err = repos.OrderCommand.CreateOrder(ctx, &requests.CreateOrderRecordRequest{
		MerchantID: merchantID, CashierID: cashierID, TotalPrice: 20000,
	})
	require.NoError(t, err)

	results, total, err := repos.OrderQuery.FindAllOrders(ctx, &requests.FindAllOrders{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 2)
	assert.GreaterOrEqual(t, len(results), 2)
}

func TestOrderCommand_Trash(t *testing.T) {
	_, repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, cashierID := seedOrderDeps(t, ctx, pool)
	created, err := repos.OrderCommand.CreateOrder(ctx, &requests.CreateOrderRecordRequest{
		MerchantID: merchantID, CashierID: cashierID, TotalPrice: 50000,
	})
	require.NoError(t, err)

	trashed, err := repos.OrderCommand.TrashedOrder(ctx, int(created.OrderID))
	require.NoError(t, err)
	require.NotNil(t, trashed)
	assert.True(t, trashed.DeletedAt.Valid)
}

func TestOrderQuery_FindByTrashed(t *testing.T) {
	_, repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, cashierID := seedOrderDeps(t, ctx, pool)
	created, err := repos.OrderCommand.CreateOrder(ctx, &requests.CreateOrderRecordRequest{
		MerchantID: merchantID, CashierID: cashierID, TotalPrice: 50000,
	})
	require.NoError(t, err)

	_, err = repos.OrderCommand.TrashedOrder(ctx, int(created.OrderID))
	require.NoError(t, err)

	results, total, err := repos.OrderQuery.FindByTrashed(ctx, &requests.FindAllOrders{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 1)
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestOrderCommand_Restore(t *testing.T) {
	_, repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, cashierID := seedOrderDeps(t, ctx, pool)
	created, err := repos.OrderCommand.CreateOrder(ctx, &requests.CreateOrderRecordRequest{
		MerchantID: merchantID, CashierID: cashierID, TotalPrice: 50000,
	})
	require.NoError(t, err)

	_, err = repos.OrderCommand.TrashedOrder(ctx, int(created.OrderID))
	require.NoError(t, err)

	restored, err := repos.OrderCommand.RestoreOrder(ctx, int(created.OrderID))
	require.NoError(t, err)
	require.NotNil(t, restored)
	assert.False(t, restored.DeletedAt.Valid)
}

func TestOrderQuery_FindByActive(t *testing.T) {
	_, repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, cashierID := seedOrderDeps(t, ctx, pool)
	created, err := repos.OrderCommand.CreateOrder(ctx, &requests.CreateOrderRecordRequest{
		MerchantID: merchantID, CashierID: cashierID, TotalPrice: 50000,
	})
	require.NoError(t, err)

	results, total, err := repos.OrderQuery.FindByActive(ctx, &requests.FindAllOrders{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 1)

	found := false
	for _, r := range results {
		if int(r.OrderID) == int(created.OrderID) {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestOrderCommand_DeletePermanent(t *testing.T) {
	_, repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, cashierID := seedOrderDeps(t, ctx, pool)
	created, err := repos.OrderCommand.CreateOrder(ctx, &requests.CreateOrderRecordRequest{
		MerchantID: merchantID, CashierID: cashierID, TotalPrice: 50000,
	})
	require.NoError(t, err)

	_, err = repos.OrderCommand.TrashedOrder(ctx, int(created.OrderID))
	require.NoError(t, err)

	deleted, err := repos.OrderCommand.DeleteOrderPermanent(ctx, int(created.OrderID))
	require.NoError(t, err)
	assert.True(t, deleted)

	_, err = repos.OrderQuery.FindById(ctx, int(created.OrderID))
	assert.Error(t, err)
}

func TestOrderQuery_FindByMerchant(t *testing.T) {
	_, repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, cashierID := seedOrderDeps(t, ctx, pool)
	_, err := repos.OrderCommand.CreateOrder(ctx, &requests.CreateOrderRecordRequest{
		MerchantID: merchantID, CashierID: cashierID, TotalPrice: 50000,
	})
	require.NoError(t, err)

	results, total, err := repos.OrderQuery.FindByMerchant(ctx, &requests.FindAllOrderMerchant{
		MerchantID: merchantID,
		Search:     "",
		Page:       1,
		PageSize:   10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 1)
	assert.GreaterOrEqual(t, len(results), 1)
}

// ----- OrderItemCommand Tests (DB-backed in order service) -----

func TestOrderItemCommand_Create(t *testing.T) {
	_, repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, cashierID := seedOrderDeps(t, ctx, pool)

	// Seed product dependencies
	var categoryID int
	err := pool.QueryRowContext(ctx, `INSERT INTO categories (name) VALUES ('Test Cat') RETURNING category_id`).Scan(&categoryID)
	require.NoError(t, err)

	var productID int
	err = pool.QueryRowContext(ctx, `INSERT INTO products (merchant_id, category_id, name, price, count_in_stock)
		VALUES ($1, $2, 'Test Product', 10000, 50) RETURNING product_id`, merchantID, categoryID).Scan(&productID)
	require.NoError(t, err)

	created, err := repos.OrderCommand.CreateOrder(ctx, &requests.CreateOrderRecordRequest{
		MerchantID: merchantID, CashierID: cashierID, TotalPrice: 20000,
	})
	require.NoError(t, err)

	item, err := repos.OrderItemCommand.CreateOrderItem(ctx, &requests.CreateOrderItemRecordRequest{
		OrderID:   int(created.OrderID),
		ProductID: productID,
		Quantity:  2,
		Price:     10000,
	})
	require.NoError(t, err)
	require.NotNil(t, item)
	assert.Equal(t, int32(created.OrderID), item.OrderID)
	assert.Equal(t, int32(productID), item.ProductID)
	assert.Equal(t, int32(2), item.Quantity)
	assert.NotZero(t, item.OrderItemID)
}

func TestOrderItemCommand_Update(t *testing.T) {
	_, repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, cashierID := seedOrderDeps(t, ctx, pool)

	var categoryID int
	err := pool.QueryRowContext(ctx, `INSERT INTO categories (name) VALUES ('Cat') RETURNING category_id`).Scan(&categoryID)
	require.NoError(t, err)

	var productID int
	err = pool.QueryRowContext(ctx, `INSERT INTO products (merchant_id, category_id, name, price, count_in_stock)
		VALUES ($1, $2, 'P', 5000, 10) RETURNING product_id`, merchantID, categoryID).Scan(&productID)
	require.NoError(t, err)

	created, err := repos.OrderCommand.CreateOrder(ctx, &requests.CreateOrderRecordRequest{
		MerchantID: merchantID, CashierID: cashierID, TotalPrice: 10000,
	})
	require.NoError(t, err)

	item, err := repos.OrderItemCommand.CreateOrderItem(ctx, &requests.CreateOrderItemRecordRequest{
		OrderID: int(created.OrderID), ProductID: productID, Quantity: 1, Price: 5000,
	})
	require.NoError(t, err)

	updated, err := repos.OrderItemCommand.UpdateOrderItem(ctx, &requests.UpdateOrderItemRecordRequest{
		OrderItemID: int(item.OrderItemID),
		OrderID:     int(created.OrderID),
		ProductID:   productID,
		Quantity:    3,
		Price:       6000,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, int32(3), updated.Quantity)
	assert.Equal(t, int32(6000), updated.Price)
}
