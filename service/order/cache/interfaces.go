package mencache

import (
	"context"

	db "github.com/MamangRust/microservice-point-of-sale-order/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type OrderQueryCache interface {
	GetOrderAllCache(ctx context.Context, req *requests.FindAllOrders) ([]*db.GetOrdersRow, *int, bool)
	SetOrderAllCache(ctx context.Context, req *requests.FindAllOrders, data []*db.GetOrdersRow, total *int)

	GetCachedOrderCache(ctx context.Context, orderID int) (*db.Order, bool)
	SetCachedOrderCache(ctx context.Context, data *db.Order)

	GetCachedOrderMerchant(ctx context.Context, req *requests.FindAllOrderMerchant) ([]*db.GetOrdersByMerchantRow, *int, bool)
	SetCachedOrderMerchant(ctx context.Context, req *requests.FindAllOrderMerchant, res []*db.GetOrdersByMerchantRow, total *int)

	GetOrderActiveCache(ctx context.Context, req *requests.FindAllOrders) ([]*db.GetOrdersActiveRow, *int, bool)
	SetOrderActiveCache(ctx context.Context, req *requests.FindAllOrders, data []*db.GetOrdersActiveRow, total *int)

	GetOrderTrashedCache(ctx context.Context, req *requests.FindAllOrders) ([]*db.GetOrdersTrashedRow, *int, bool)
	SetOrderTrashedCache(ctx context.Context, req *requests.FindAllOrders, data []*db.GetOrdersTrashedRow, total *int)
}

type OrderCommandCache interface {
	DeleteOrderCache(ctx context.Context, id int)
	DeleteOrderAllCache(ctx context.Context)
}
