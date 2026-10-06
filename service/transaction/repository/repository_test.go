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

func setupTestDB(t *testing.T) (*sqlx.DB, *Repositories, func()) {
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

	db, err := sqlx.Connect("pgx", connStr)
	require.NoError(t, err)

	waitForDatabase(t, ctx, db)

	runMigrations(t, ctx, connStr, "../../../service/*/database/migration/*.sql")

	repos := NewRepositories(db, nil, nil, nil, nil)

	cleanup := func() {
		db.Close()
		pgContainer.Terminate(ctx)
	}

	return db, repos, cleanup
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
func waitForDatabase(t *testing.T, ctx context.Context, db *sqlx.DB) {
	t.Helper()

	deadline := time.Now().Add(60 * time.Second)
	for {
		var one int
		err := db.QueryRowContext(ctx, `SELECT 1`).Scan(&one)
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

func seedTransactionDeps(t *testing.T, ctx context.Context, db *sqlx.DB) (int, int, int) {
	t.Helper()

	var userID int
	err := db.QueryRowContext(ctx, `INSERT INTO users (firstname, lastname, email, password, verification_code)
		VALUES ('Trans', 'User', 'trans@example.com', 'password', 'CODE') RETURNING user_id`).Scan(&userID)
	require.NoError(t, err)

	var merchantID int
	err = db.QueryRowContext(ctx, `INSERT INTO merchants (user_id, name, status) VALUES ($1, 'Trans Merchant', 'active') RETURNING merchant_id`, userID).Scan(&merchantID)
	require.NoError(t, err)

	var cashierID int
	err = db.QueryRowContext(ctx, `INSERT INTO cashiers (merchant_id, user_id, name) VALUES ($1, $2, 'Trans Cashier') RETURNING cashier_id`, merchantID, userID).Scan(&cashierID)
	require.NoError(t, err)

	var orderID int
	err = db.QueryRowContext(ctx, `INSERT INTO orders (merchant_id, cashier_id, total_price) VALUES ($1, $2, 100000) RETURNING order_id`, merchantID, cashierID).Scan(&orderID)
	require.NoError(t, err)

	return merchantID, cashierID, orderID
}

func TestTransactionCommand_Create(t *testing.T) {
	db, repos, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, _, orderID := seedTransactionDeps(t, ctx, db)
	paymentStatus := "completed"

	txn, err := repos.TransactionCommandRepository.CreateTransaction(ctx, &requests.CreateTransactionRequest{
		OrderID:       orderID,
		CashierID:     0, // Not used in create
		MerchantID:    merchantID,
		PaymentMethod: "cash",
		Amount:        100000,
		ChangeAmount:  ptr(0),
		PaymentStatus: &paymentStatus,
	})
	require.NoError(t, err)
	require.NotNil(t, txn)
	assert.Equal(t, int32(orderID), txn.OrderID)
	assert.Equal(t, int32(merchantID), txn.MerchantID)
	assert.Equal(t, "cash", txn.PaymentMethod)
	assert.Equal(t, int32(100000), txn.Amount)
	assert.Equal(t, "completed", txn.PaymentStatus)
	assert.NotZero(t, txn.TransactionID)
}

func TestTransactionQuery_FindByID(t *testing.T) {
	db, repos, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, _, orderID := seedTransactionDeps(t, ctx, db)
	paymentStatus := "completed"

	created, err := repos.TransactionCommandRepository.CreateTransaction(ctx, &requests.CreateTransactionRequest{
		OrderID: orderID, MerchantID: merchantID, PaymentMethod: "card",
		Amount: 50000, PaymentStatus: &paymentStatus,
	})
	require.NoError(t, err)

	found, err := repos.TransactionQueryRepository.FindById(ctx, int(created.TransactionID))
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.TransactionID, found.TransactionID)
	assert.Equal(t, "card", found.PaymentMethod)
}

func TestTransactionQuery_FindByOrderID(t *testing.T) {
	db, repos, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, _, orderID := seedTransactionDeps(t, ctx, db)
	paymentStatus := "completed"

	_, err := repos.TransactionCommandRepository.CreateTransaction(ctx, &requests.CreateTransactionRequest{
		OrderID: orderID, MerchantID: merchantID, PaymentMethod: "qris",
		Amount: 75000, PaymentStatus: &paymentStatus,
	})
	require.NoError(t, err)

	found, err := repos.TransactionQueryRepository.FindByOrderId(ctx, orderID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, int32(orderID), found.OrderID)
}

func TestTransactionCommand_Update(t *testing.T) {
	db, repos, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, cashierID, orderID := seedTransactionDeps(t, ctx, db)
	paymentStatus := "completed"

	created, err := repos.TransactionCommandRepository.CreateTransaction(ctx, &requests.CreateTransactionRequest{
		OrderID: orderID, MerchantID: merchantID, PaymentMethod: "cash",
		Amount: 50000, PaymentStatus: &paymentStatus,
	})
	require.NoError(t, err)

	newStatus := "refunded"
	updated, err := repos.TransactionCommandRepository.UpdateTransaction(ctx, &requests.UpdateTransactionRequest{
		TransactionID: ptr(int(created.TransactionID)),
		OrderID:       orderID,
		CashierID:     cashierID,
		MerchantID:    merchantID,
		PaymentMethod: "transfer",
		Amount:        60000,
		PaymentStatus: &newStatus,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "transfer", updated.PaymentMethod)
	assert.Equal(t, "refunded", updated.PaymentStatus)
	assert.Equal(t, int32(60000), updated.Amount)
}

func TestTransactionQuery_FindAll(t *testing.T) {
	db, repos, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, _, orderID := seedTransactionDeps(t, ctx, db)
	paymentStatus := "completed"

	_, err := repos.TransactionCommandRepository.CreateTransaction(ctx, &requests.CreateTransactionRequest{
		OrderID: orderID, MerchantID: merchantID, PaymentMethod: "cash",
		Amount: 10000, PaymentStatus: &paymentStatus,
	})
	require.NoError(t, err)
	_, err = repos.TransactionCommandRepository.CreateTransaction(ctx, &requests.CreateTransactionRequest{
		OrderID: orderID, MerchantID: merchantID, PaymentMethod: "card",
		Amount: 20000, PaymentStatus: &paymentStatus,
	})
	require.NoError(t, err)

	results, total, err := repos.TransactionQueryRepository.FindAllTransactions(ctx, &requests.FindAllTransaction{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 2)
	assert.GreaterOrEqual(t, len(results), 2)
}

func TestTransactionCommand_Trash(t *testing.T) {
	db, repos, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, _, orderID := seedTransactionDeps(t, ctx, db)
	paymentStatus := "completed"

	created, err := repos.TransactionCommandRepository.CreateTransaction(ctx, &requests.CreateTransactionRequest{
		OrderID: orderID, MerchantID: merchantID, PaymentMethod: "cash",
		Amount: 30000, PaymentStatus: &paymentStatus,
	})
	require.NoError(t, err)

	trashed, err := repos.TransactionCommandRepository.TrashTransaction(ctx, int(created.TransactionID))
	require.NoError(t, err)
	require.NotNil(t, trashed)
	assert.True(t, trashed.DeletedAt.Valid)
}

func TestTransactionQuery_FindByTrashed(t *testing.T) {
	db, repos, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, _, orderID := seedTransactionDeps(t, ctx, db)
	paymentStatus := "completed"

	created, err := repos.TransactionCommandRepository.CreateTransaction(ctx, &requests.CreateTransactionRequest{
		OrderID: orderID, MerchantID: merchantID, PaymentMethod: "cash",
		Amount: 40000, PaymentStatus: &paymentStatus,
	})
	require.NoError(t, err)

	_, err = repos.TransactionCommandRepository.TrashTransaction(ctx, int(created.TransactionID))
	require.NoError(t, err)

	results, total, err := repos.TransactionQueryRepository.FindByTrashed(ctx, &requests.FindAllTransaction{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 1)
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestTransactionCommand_Restore(t *testing.T) {
	db, repos, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, _, orderID := seedTransactionDeps(t, ctx, db)
	paymentStatus := "completed"

	created, err := repos.TransactionCommandRepository.CreateTransaction(ctx, &requests.CreateTransactionRequest{
		OrderID: orderID, MerchantID: merchantID, PaymentMethod: "cash",
		Amount: 50000, PaymentStatus: &paymentStatus,
	})
	require.NoError(t, err)

	_, err = repos.TransactionCommandRepository.TrashTransaction(ctx, int(created.TransactionID))
	require.NoError(t, err)

	restored, err := repos.TransactionCommandRepository.RestoreTransaction(ctx, int(created.TransactionID))
	require.NoError(t, err)
	require.NotNil(t, restored)
	assert.False(t, restored.DeletedAt.Valid)
}

func TestTransactionQuery_FindByActive(t *testing.T) {
	db, repos, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, _, orderID := seedTransactionDeps(t, ctx, db)
	paymentStatus := "completed"

	created, err := repos.TransactionCommandRepository.CreateTransaction(ctx, &requests.CreateTransactionRequest{
		OrderID: orderID, MerchantID: merchantID, PaymentMethod: "cash",
		Amount: 60000, PaymentStatus: &paymentStatus,
	})
	require.NoError(t, err)

	results, total, err := repos.TransactionQueryRepository.FindByActive(ctx, &requests.FindAllTransaction{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 1)

	found := false
	for _, r := range results {
		if int(r.TransactionID) == int(created.TransactionID) {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestTransactionCommand_DeletePermanent(t *testing.T) {
	db, repos, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, _, orderID := seedTransactionDeps(t, ctx, db)
	paymentStatus := "completed"

	created, err := repos.TransactionCommandRepository.CreateTransaction(ctx, &requests.CreateTransactionRequest{
		OrderID: orderID, MerchantID: merchantID, PaymentMethod: "cash",
		Amount: 70000, PaymentStatus: &paymentStatus,
	})
	require.NoError(t, err)

	_, err = repos.TransactionCommandRepository.TrashTransaction(ctx, int(created.TransactionID))
	require.NoError(t, err)

	deleted, err := repos.TransactionCommandRepository.DeleteTransactionPermanently(ctx, int(created.TransactionID))
	require.NoError(t, err)
	assert.True(t, deleted)

	_, err = repos.TransactionQueryRepository.FindById(ctx, int(created.TransactionID))
	assert.Error(t, err)
}

func TestTransactionQuery_FindByMerchant(t *testing.T) {
	db, repos, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, _, orderID := seedTransactionDeps(t, ctx, db)
	paymentStatus := "completed"

	_, err := repos.TransactionCommandRepository.CreateTransaction(ctx, &requests.CreateTransactionRequest{
		OrderID: orderID, MerchantID: merchantID, PaymentMethod: "cash",
		Amount: 80000, PaymentStatus: &paymentStatus,
	})
	require.NoError(t, err)

	results, total, err := repos.TransactionQueryRepository.FindByMerchant(ctx, &requests.FindAllTransactionByMerchant{
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
