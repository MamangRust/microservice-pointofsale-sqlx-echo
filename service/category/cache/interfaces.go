package mencache

import (
	"context"

	db "github.com/MamangRust/microservice-point-of-sale-category/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type CategoryQueryCache interface {
	GetCachedCategoriesCache(ctx context.Context, req *requests.FindAllCategory) ([]*db.GetCategoriesRow, *int, bool)
	SetCachedCategoriesCache(ctx context.Context, req *requests.FindAllCategory, data []*db.GetCategoriesRow, total *int)

	GetCachedCategoryActiveCache(ctx context.Context, req *requests.FindAllCategory) ([]*db.GetCategoriesActiveRow, *int, bool)
	SetCachedCategoryActiveCache(ctx context.Context, req *requests.FindAllCategory, data []*db.GetCategoriesActiveRow, total *int)

	GetCachedCategoryTrashedCache(ctx context.Context, req *requests.FindAllCategory) ([]*db.GetCategoriesTrashedRow, *int, bool)
	SetCachedCategoryTrashedCache(ctx context.Context, req *requests.FindAllCategory, data []*db.GetCategoriesTrashedRow, total *int)

	GetCachedCategoryCache(ctx context.Context, id int) (*db.Category, bool)
	SetCachedCategoryCache(ctx context.Context, data *db.Category)
}

type CategoryCommandCache interface {
	DeleteCachedCategoryCache(ctx context.Context, id int)
	DeleteCachedCategoryAllCache(ctx context.Context)
}
