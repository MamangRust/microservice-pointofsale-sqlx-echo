package product_test

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
)

// stubCategoryRepo implements categoryadapter.Repository. The product tests
// that use it only exercise the SQL-backed product repositories, so the
// category dependency is never queried.
type stubCategoryRepo struct{}

func (s *stubCategoryRepo) FindById(ctx context.Context, id int) (*models.Category, error) {
	return &models.Category{CategoryID: int32(id)}, nil
}

func (s *stubCategoryRepo) FindByIds(ctx context.Context, ids []int) ([]*models.Category, error) {
	return nil, nil
}

func (s *stubCategoryRepo) FindByName(ctx context.Context, name string) (*models.Category, error) {
	return &models.Category{Name: name}, nil
}

// stubMerchantRepo implements merchantadapter.Repository. Like stubCategoryRepo
// it is never queried by the SQL-only product tests.
type stubMerchantRepo struct{}

func (s *stubMerchantRepo) FindById(ctx context.Context, id int) (*models.Merchant, error) {
	return &models.Merchant{MerchantID: int32(id)}, nil
}
