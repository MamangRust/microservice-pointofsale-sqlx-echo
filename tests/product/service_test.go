package product_test

import (
	"context"
	"testing"

	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	prod_cache "github.com/MamangRust/microservice-point-of-sale-product/cache"
	"github.com/MamangRust/microservice-point-of-sale-product/repository"
	"github.com/MamangRust/microservice-point-of-sale-product/service"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
)

type ProductServiceTestSuite struct {
	suite.Suite
	ts        *tests.TestSuite
	svc       *service.Service
	productID int
}

func (s *ProductServiceTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	pool, err := sqlx.ConnectContext(s.ts.Ctx, "pgx", s.ts.DBURL)
	s.Require().NoError(err)

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	redisClient := redis.NewClient(opts)

	productQueries := pool

	log, _ := logger.NewLogger("test", nil)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)
	mencache := prod_cache.NewMencache(cacheStore)

	obs, _ := observability.NewObservability("test", log)

	repos := &repository.Repositories{
		ProductQuery:   repository.NewProductQueryRepository(productQueries),
		ProductCommand: repository.NewProductCommandRepository(productQueries),
		CategoryQuery:  &stubCategoryRepo{},
		MerchantQuery:  &stubMerchantRepo{},
	}

	s.svc = service.NewService(&service.Deps{
		Repositories:  repos,
		Logger:        log,
		Mencache:      mencache,
		Observability: obs,
	})

	// Seed a merchant and category for product tests
	var userID, merchantID, categoryID int
	err = pool.QueryRowContext(s.ts.Ctx,
		`INSERT INTO users (firstname, lastname, email, password, verification_code, is_verified) VALUES ($1, $2, $3, $4, 'test-verify', true) RETURNING user_id`,
		"Prod", "Svc", "prod.svc@example.com", "password123",
	).Scan(&userID)
	s.Require().NoError(err)

	err = pool.QueryRowContext(s.ts.Ctx,
		`INSERT INTO categories (name, description) VALUES ($1, $2) RETURNING category_id`,
		"Svc Category", "Category for service tests",
	).Scan(&categoryID)
	s.Require().NoError(err)

	err = pool.QueryRowContext(s.ts.Ctx,
		`INSERT INTO merchants (user_id, name, description, address, contact_email, contact_phone, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING merchant_id`,
		userID, "Svc Merchant", "Desc", "Addr", "ps@example.com", "123", "active",
	).Scan(&merchantID)
	s.Require().NoError(err)
}

func (s *ProductServiceTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *ProductServiceTestSuite) TestProductLifecycle() {
	ctx := context.Background()

	// 1. Create
	req := &requests.CreateProductRequest{
		MerchantID:   1,
		CategoryID:   1,
		Name:         "Test Product",
		Description:  "Product description",
		Price:        10000,
		CountInStock: 100,
		Brand:        "Test Brand",
		Weight:       1000,
		ImageProduct: "test.jpg",
	}
	created, err := s.svc.ProductCommand.CreateProduct(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(created)
	productID := int(created.ProductID)

	// 2. FindByID
	found, err := s.svc.ProductQuery.FindById(ctx, productID)
	s.Require().NoError(err)
	s.Equal(req.Name, found.Name)

	// 3. Update
	updateReq := &requests.UpdateProductRequest{
		ProductID:    &productID,
		MerchantID:   1,
		CategoryID:   1,
		Name:         "Updated Product Name",
		Description:  "Updated description",
		Price:        20000,
		CountInStock: 90,
		Brand:        "Updated Brand",
		Weight:       1000,
		ImageProduct: "updated.jpg",
	}
	updated, err := s.svc.ProductCommand.UpdateProduct(ctx, updateReq)
	s.Require().NoError(err)
	s.Equal(updateReq.Name, updated.Name)

	// 4. FindAll
	_, total, err := s.svc.ProductQuery.FindAll(ctx, &requests.FindAllProducts{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*total, 1)

	// 5. Trash
	_, err = s.svc.ProductCommand.TrashProduct(ctx, productID)
	s.Require().NoError(err)

	// 6. FindTrashed
	_, totalTrashed, err := s.svc.ProductQuery.FindByTrashed(ctx, &requests.FindAllProducts{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*totalTrashed, 1)

	// 7. FindActive
	active, _, err := s.svc.ProductQuery.FindByActive(ctx, &requests.FindAllProducts{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	for _, p := range active {
		s.NotEqual(productID, int(p.ProductID))
	}

	// 8. Restore
	_, err = s.svc.ProductCommand.RestoreProduct(ctx, productID)
	s.Require().NoError(err)

	// 9. DeletePermanent
	_, err = s.svc.ProductCommand.TrashProduct(ctx, productID)
	s.Require().NoError(err)
	success, err := s.svc.ProductCommand.DeleteProductPermanent(ctx, productID)
	s.Require().NoError(err)
	s.True(success)

	// 10. RestoreAll & DeleteAll
	p1, _ := s.svc.ProductCommand.CreateProduct(ctx, &requests.CreateProductRequest{
		MerchantID: 1, CategoryID: 1, Name: "P1", Description: "D1", Price: 100,
		CountInStock: 10, Brand: "B", Weight: 1, ImageProduct: "i1.jpg",
	})
	p2, _ := s.svc.ProductCommand.CreateProduct(ctx, &requests.CreateProductRequest{
		MerchantID: 1, CategoryID: 1, Name: "P2", Description: "D2", Price: 200,
		CountInStock: 20, Brand: "B", Weight: 2, ImageProduct: "i2.jpg",
	})

	s.svc.ProductCommand.TrashProduct(ctx, int(p1.ProductID))
	s.svc.ProductCommand.TrashProduct(ctx, int(p2.ProductID))

	resRestoreAll, err := s.svc.ProductCommand.RestoreAllProducts(ctx)
	s.Require().NoError(err)
	s.True(resRestoreAll)

	s.svc.ProductCommand.TrashProduct(ctx, int(p1.ProductID))
	s.svc.ProductCommand.TrashProduct(ctx, int(p2.ProductID))

	resDeleteAll, err := s.svc.ProductCommand.DeleteAllProductsPermanent(ctx)
	s.Require().NoError(err)
	s.True(resDeleteAll)
}

func TestProductServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(ProductServiceTestSuite))
}
