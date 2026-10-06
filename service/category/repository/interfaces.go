package repository

import (
	"context"

	db "github.com/MamangRust/microservice-point-of-sale-category/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type CategoryQueryRepository interface {
	FindAllCategory(ctx context.Context, req *requests.FindAllCategory) ([]*db.GetCategoriesRow, *int, error)
	FindById(ctx context.Context, category_id int) (*db.Category, error)
	FindByIds(ctx context.Context, ids []int) ([]*db.Category, error)
	FindByNameAndId(ctx context.Context, req *requests.CategoryNameAndId) (*db.Category, error)
	FindByName(ctx context.Context, name string) (*db.Category, error)

	FindByIdTrashed(ctx context.Context, category_id int) (*db.Category, error)

	FindByActive(ctx context.Context, req *requests.FindAllCategory) ([]*db.GetCategoriesActiveRow, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllCategory) ([]*db.GetCategoriesTrashedRow, *int, error)
}

type CategoryCommandRepository interface {
	CreateCategory(ctx context.Context, request *requests.CreateCategoryRequest) (*db.Category, error)
	UpdateCategory(ctx context.Context, request *requests.UpdateCategoryRequest) (*db.Category, error)
	TrashedCategory(ctx context.Context, category_id int) (*db.Category, error)
	RestoreCategory(ctx context.Context, category_id int) (*db.Category, error)
	DeleteCategoryPermanently(ctx context.Context, category_id int) (bool, error)
	RestoreAllCategories(ctx context.Context) (bool, error)
	DeleteAllPermanentCategories(ctx context.Context) (bool, error)
}
