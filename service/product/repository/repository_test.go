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

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

// stubCategoryQuery is a test double for categoryadapter.Repository that reads
// categories straight from the test database, so these repository tests can
// still resolve category names without standing up the category gRPC service.
type stubCategoryQuery struct{ pool *sqlx.DB }

func (s *stubCategoryQuery) FindById(ctx context.Context, id int) (*models.Category, error) {
	var c models.Category
	err := s.pool.QueryRowContext(ctx, `SELECT category_id, name FROM categories WHERE category_id = $1`, id).
		Scan(&c.CategoryID, &c.Name)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (s *stubCategoryQuery) FindByIds(ctx context.Context, ids []int) ([]*models.Category, error) {
	out := make([]*models.Category, 0, len(ids))
	for _, id := range ids {
		c, err := s.FindById(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}

func (s *stubCategoryQuery) FindByName(ctx context.Context, name string) (*models.Category, error) {
	var c models.Category
	err := s.pool.QueryRowContext(ctx, `SELECT category_id, name FROM categories WHERE name = $1`, name).
		Scan(&c.CategoryID, &c.Name)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// stubMerchantQuery is a test double for merchantadapter.Repository.
type stubMerchantQuery struct{ pool *sqlx.DB }

func (s *stubMerchantQuery) FindById(ctx context.Context, id int) (*models.Merchant, error) {
	var m models.Merchant
	err := s.pool.QueryRowContext(ctx, `SELECT merchant_id, name FROM merchants WHERE merchant_id = $1`, id).
		Scan(&m.MerchantID, &m.Name)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

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

	repos := &Repositories{
		ProductQuery:   NewProductQueryRepository(pool),
		ProductCommand: NewProductCommandRepository(pool),
		CategoryQuery:  &stubCategoryQuery{pool: pool},
		MerchantQuery:  &stubMerchantQuery{pool: pool},
	}

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

func seedProductDependencies(t *testing.T, ctx context.Context, pool *sqlx.DB) (int, int) {
	t.Helper()

	var userID int
	err := pool.QueryRowContext(ctx, `INSERT INTO users (firstname, lastname, email, password, verification_code)
		VALUES ('Merchant', 'User', 'merchant@example.com', 'password', 'CODE') RETURNING user_id`).Scan(&userID)
	require.NoError(t, err)

	var merchantID int
	err = pool.QueryRowContext(ctx, `INSERT INTO merchants (user_id, name, status) VALUES ($1, 'Test Merchant', 'active') RETURNING merchant_id`, userID).Scan(&merchantID)
	require.NoError(t, err)

	var categoryID int
	err = pool.QueryRowContext(ctx, `INSERT INTO categories (name) VALUES ('Test Category') RETURNING category_id`).Scan(&categoryID)
	require.NoError(t, err)

	return merchantID, categoryID
}

func TestProductCommand_Create(t *testing.T) {
	repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, categoryID := seedProductDependencies(t, ctx, pool)

	product, err := repos.ProductCommand.CreateProduct(ctx, &requests.CreateProductRequest{
		MerchantID:   merchantID,
		CategoryID:   categoryID,
		Name:         "Smartphone",
		Description:  "High-end smartphone",
		Price:        5000000,
		CountInStock: 100,
		Brand:        "TechBrand",
		Weight:       250,
		ImageProduct: "https://example.com/image.jpg",
	})
	require.NoError(t, err)
	require.NotNil(t, product)
	assert.Equal(t, "Smartphone", product.Name)
	assert.Equal(t, int32(merchantID), product.MerchantID)
	assert.Equal(t, int32(categoryID), product.CategoryID)
	assert.Equal(t, int32(5000000), product.Price)
	assert.NotZero(t, product.ProductID)
}

func TestProductQuery_FindByID(t *testing.T) {
	repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, categoryID := seedProductDependencies(t, ctx, pool)

	created, err := repos.ProductCommand.CreateProduct(ctx, &requests.CreateProductRequest{
		MerchantID:   merchantID,
		CategoryID:   categoryID,
		Name:         "Laptop",
		Description:  "Gaming laptop",
		Price:        15000000,
		CountInStock: 50,
		Brand:        "GameTech",
		Weight:       2000,
		ImageProduct: "https://example.com/laptop.jpg",
	})
	require.NoError(t, err)

	found, err := repos.ProductQuery.FindById(ctx, int(created.ProductID))
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, created.ProductID, found.ProductID)
	assert.Equal(t, "Laptop", found.Name)
}

func TestProductCommand_Update(t *testing.T) {
	repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, categoryID := seedProductDependencies(t, ctx, pool)

	created, err := repos.ProductCommand.CreateProduct(ctx, &requests.CreateProductRequest{
		MerchantID:   merchantID,
		CategoryID:   categoryID,
		Name:         "Old Product",
		Description:  "Old description",
		Price:        100000,
		CountInStock: 10,
		Brand:        "OldBrand",
		Weight:       100,
		ImageProduct: "https://example.com/old.jpg",
	})
	require.NoError(t, err)

	updated, err := repos.ProductCommand.UpdateProduct(ctx, &requests.UpdateProductRequest{
		ProductID:    ptr(int(created.ProductID)),
		MerchantID:   merchantID,
		CategoryID:   categoryID,
		Name:         "New Product",
		Description:  "New description",
		Price:        200000,
		CountInStock: 20,
		Brand:        "NewBrand",
		Weight:       200,
		ImageProduct: "https://example.com/new.jpg",
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "New Product", updated.Name)
	assert.Equal(t, int32(200000), updated.Price)
}

func TestProductCommand_UpdateCountStock(t *testing.T) {
	repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, categoryID := seedProductDependencies(t, ctx, pool)

	created, err := repos.ProductCommand.CreateProduct(ctx, &requests.CreateProductRequest{
		MerchantID:   merchantID,
		CategoryID:   categoryID,
		Name:         "StockTest",
		Description:  "Testing stock",
		Price:        50000,
		CountInStock: 10,
		Brand:        "Brand",
		Weight:       100,
		ImageProduct: "https://example.com/stock.jpg",
	})
	require.NoError(t, err)

	updated, err := repos.ProductCommand.UpdateProductCountStock(ctx, int(created.ProductID), 25)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, int32(25), updated.CountInStock)
}

func TestProductQuery_FindAll(t *testing.T) {
	repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, categoryID := seedProductDependencies(t, ctx, pool)

	_, err := repos.ProductCommand.CreateProduct(ctx, &requests.CreateProductRequest{
		MerchantID: merchantID, CategoryID: categoryID, Name: "Product1",
		Description: "Desc1", Price: 1000, CountInStock: 10, Brand: "B1", Weight: 100,
		ImageProduct: "https://example.com/1.jpg",
	})
	require.NoError(t, err)
	_, err = repos.ProductCommand.CreateProduct(ctx, &requests.CreateProductRequest{
		MerchantID: merchantID, CategoryID: categoryID, Name: "Product2",
		Description: "Desc2", Price: 2000, CountInStock: 20, Brand: "B2", Weight: 200,
		ImageProduct: "https://example.com/2.jpg",
	})
	require.NoError(t, err)

	results, total, err := repos.ProductQuery.FindAllProducts(ctx, &requests.FindAllProducts{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 2)
	assert.GreaterOrEqual(t, len(results), 2)
}

func TestProductCommand_Trash(t *testing.T) {
	repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, categoryID := seedProductDependencies(t, ctx, pool)
	created, err := repos.ProductCommand.CreateProduct(ctx, &requests.CreateProductRequest{
		MerchantID: merchantID, CategoryID: categoryID, Name: "TrashMe",
		Description: "Desc", Price: 1000, CountInStock: 10, Brand: "B", Weight: 100,
		ImageProduct: "https://example.com/t.jpg",
	})
	require.NoError(t, err)

	trashed, err := repos.ProductCommand.TrashedProduct(ctx, int(created.ProductID))
	require.NoError(t, err)
	require.NotNil(t, trashed)
	assert.True(t, trashed.DeletedAt.Valid)
}

func TestProductQuery_FindByTrashed(t *testing.T) {
	repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, categoryID := seedProductDependencies(t, ctx, pool)
	created, err := repos.ProductCommand.CreateProduct(ctx, &requests.CreateProductRequest{
		MerchantID: merchantID, CategoryID: categoryID, Name: "TrashFind",
		Description: "Desc", Price: 1000, CountInStock: 10, Brand: "B", Weight: 100,
		ImageProduct: "https://example.com/tf.jpg",
	})
	require.NoError(t, err)

	_, err = repos.ProductCommand.TrashedProduct(ctx, int(created.ProductID))
	require.NoError(t, err)

	results, total, err := repos.ProductQuery.FindByTrashed(ctx, &requests.FindAllProducts{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 1)
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestProductCommand_Restore(t *testing.T) {
	repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, categoryID := seedProductDependencies(t, ctx, pool)
	created, err := repos.ProductCommand.CreateProduct(ctx, &requests.CreateProductRequest{
		MerchantID: merchantID, CategoryID: categoryID, Name: "RestoreMe",
		Description: "Desc", Price: 1000, CountInStock: 10, Brand: "B", Weight: 100,
		ImageProduct: "https://example.com/r.jpg",
	})
	require.NoError(t, err)

	_, err = repos.ProductCommand.TrashedProduct(ctx, int(created.ProductID))
	require.NoError(t, err)

	restored, err := repos.ProductCommand.RestoreProduct(ctx, int(created.ProductID))
	require.NoError(t, err)
	require.NotNil(t, restored)
	assert.False(t, restored.DeletedAt.Valid)
}

func TestProductQuery_FindByActive(t *testing.T) {
	repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, categoryID := seedProductDependencies(t, ctx, pool)
	created, err := repos.ProductCommand.CreateProduct(ctx, &requests.CreateProductRequest{
		MerchantID: merchantID, CategoryID: categoryID, Name: "ActiveProd",
		Description: "Desc", Price: 1000, CountInStock: 10, Brand: "B", Weight: 100,
		ImageProduct: "https://example.com/a.jpg",
	})
	require.NoError(t, err)

	results, total, err := repos.ProductQuery.FindByActive(ctx, &requests.FindAllProducts{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 1)

	found := false
	for _, r := range results {
		if r.ProductID == created.ProductID {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestProductCommand_DeletePermanent(t *testing.T) {
	repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, categoryID := seedProductDependencies(t, ctx, pool)
	created, err := repos.ProductCommand.CreateProduct(ctx, &requests.CreateProductRequest{
		MerchantID: merchantID, CategoryID: categoryID, Name: "PermDelete",
		Description: "Desc", Price: 1000, CountInStock: 10, Brand: "B", Weight: 100,
		ImageProduct: "https://example.com/p.jpg",
	})
	require.NoError(t, err)

	_, err = repos.ProductCommand.TrashedProduct(ctx, int(created.ProductID))
	require.NoError(t, err)

	deleted, err := repos.ProductCommand.DeleteProductPermanent(ctx, int(created.ProductID))
	require.NoError(t, err)
	assert.True(t, deleted)

	_, err = repos.ProductQuery.FindById(ctx, int(created.ProductID))
	assert.Error(t, err)
}

func TestProductQuery_FindByMerchant(t *testing.T) {
	repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, categoryID := seedProductDependencies(t, ctx, pool)

	_, err := repos.ProductCommand.CreateProduct(ctx, &requests.CreateProductRequest{
		MerchantID: merchantID, CategoryID: categoryID, Name: "MProd",
		Description: "Desc", Price: 1000, CountInStock: 10, Brand: "B", Weight: 100,
		ImageProduct: "https://example.com/m.jpg",
	})
	require.NoError(t, err)

	results, total, err := repos.ProductQuery.FindByMerchant(ctx, &requests.ProductByMerchantRequest{
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

func TestProductQuery_FindByCategory(t *testing.T) {
	repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	merchantID, categoryID := seedProductDependencies(t, ctx, pool)

	_, err := repos.ProductCommand.CreateProduct(ctx, &requests.CreateProductRequest{
		MerchantID: merchantID, CategoryID: categoryID, Name: "CProd",
		Description: "Desc", Price: 1000, CountInStock: 10, Brand: "B", Weight: 100,
		ImageProduct: "https://example.com/c.jpg",
	})
	require.NoError(t, err)

	results, total, err := repos.ProductQuery.FindByCategoryID(ctx, int32(categoryID), &requests.ProductByCategoryRequest{
		CategoryName: "Test Category",
		Search:       "",
		Page:         1,
		PageSize:     10,
	})
	require.NoError(t, err)
	require.NotNil(t, total)
	assert.GreaterOrEqual(t, *total, 1)
	assert.GreaterOrEqual(t, len(results), 1)
}

func TestCategoryQuery_FindById(t *testing.T) {
	repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	var categoryID int
	err := pool.QueryRowContext(ctx, `INSERT INTO categories (name) VALUES ('TestCat') RETURNING category_id`).Scan(&categoryID)
	require.NoError(t, err)

	found, err := repos.CategoryQuery.FindById(ctx, categoryID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "TestCat", found.Name)
}

func TestMerchantQuery_FindById(t *testing.T) {
	repos, pool, cleanup := setupTestDB(t)
	defer cleanup()
	ctx := context.Background()

	var userID int
	err := pool.QueryRowContext(ctx, `INSERT INTO users (firstname, lastname, email, password, verification_code)
		VALUES ('M', 'User', 'm@example.com', 'pass', 'CODE') RETURNING user_id`).Scan(&userID)
	require.NoError(t, err)

	var merchantID int
	err = pool.QueryRowContext(ctx, `INSERT INTO merchants (user_id, name, status) VALUES ($1, 'TestM', 'active') RETURNING merchant_id`, userID).Scan(&merchantID)
	require.NoError(t, err)

	found, err := repos.MerchantQuery.FindById(ctx, merchantID)
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "TestM", found.Name)
}
