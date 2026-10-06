package service

import (
	"context"

	db "github.com/MamangRust/microservice-point-of-sale-category/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type CategoryQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllCategory) ([]*db.GetCategoriesRow, *int, error)
	FindById(ctx context.Context, category_id int) (*db.Category, error)
	FindByIds(ctx context.Context, ids []int) ([]*db.Category, error)
	FindByName(ctx context.Context, name string) (*db.Category, error)
	FindByActive(ctx context.Context, req *requests.FindAllCategory) ([]*db.GetCategoriesActiveRow, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllCategory) ([]*db.GetCategoriesTrashedRow, *int, error)
}

type CategoryCommandService interface {
	CreateCategory(ctx context.Context, req *requests.CreateCategoryRequest) (*db.Category, error)
	UpdateCategory(ctx context.Context, req *requests.UpdateCategoryRequest) (*db.Category, error)
	TrashedCategory(ctx context.Context, category_id int) (*db.Category, error)
	RestoreCategory(ctx context.Context, categoryID int) (*db.Category, error)
	DeleteCategoryPermanent(ctx context.Context, categoryID int) (bool, error)
	RestoreAllCategories(ctx context.Context) (bool, error)
	DeleteAllCategoriesPermanent(ctx context.Context) (bool, error)
}
