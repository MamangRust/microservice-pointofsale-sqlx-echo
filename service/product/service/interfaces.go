package service

import (
	"context"

	db "github.com/MamangRust/microservice-point-of-sale-product/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type ProductQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllProducts) ([]*db.GetProductsRow, *int, error)
	FindByMerchant(ctx context.Context, req *requests.ProductByMerchantRequest) ([]*db.GetProductsByMerchantRow, *int, error)
	FindByCategory(ctx context.Context, req *requests.ProductByCategoryRequest) ([]*db.GetProductsByCategoryNameRow, *int, error)
	FindById(ctx context.Context, productID int) (*db.Product, error)
	FindByActive(ctx context.Context, req *requests.FindAllProducts) ([]*db.GetProductsActiveRow, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllProducts) ([]*db.GetProductsTrashedRow, *int, error)
}

type ProductCommandService interface {
	CreateProduct(ctx context.Context, req *requests.CreateProductRequest) (*db.Product, error)
	UpdateProduct(ctx context.Context, req *requests.UpdateProductRequest) (*db.Product, error)
	IncrementProductCountStock(ctx context.Context, productID int, quantity int) (*db.Product, error)
	DecrementProductCountStock(ctx context.Context, productID int, quantity int) (*db.Product, error)
	TrashProduct(ctx context.Context, productID int) (*db.Product, error)
	RestoreProduct(ctx context.Context, productID int) (*db.Product, error)
	DeleteProductPermanent(ctx context.Context, productID int) (bool, error)
	RestoreAllProducts(ctx context.Context) (bool, error)
	DeleteAllProductsPermanent(ctx context.Context) (bool, error)
}
