package mencache

import (
	"context"

	db "github.com/MamangRust/microservice-point-of-sale-order-item/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type OrderItemQueryCache interface {
	GetCachedOrderItemsAll(ctx context.Context, req *requests.FindAllOrderItems) ([]*db.GetOrderItemsRow, *int, bool)
	SetCachedOrderItemsAll(ctx context.Context, req *requests.FindAllOrderItems, data []*db.GetOrderItemsRow, total *int)

	GetCachedOrderItemActive(ctx context.Context, req *requests.FindAllOrderItems) ([]*db.GetOrderItemsActiveRow, *int, bool)
	SetCachedOrderItemActive(ctx context.Context, req *requests.FindAllOrderItems, data []*db.GetOrderItemsActiveRow, total *int)

	GetCachedOrderItemTrashed(ctx context.Context, req *requests.FindAllOrderItems) ([]*db.GetOrderItemsTrashedRow, *int, bool)
	SetCachedOrderItemTrashed(ctx context.Context, req *requests.FindAllOrderItems, data []*db.GetOrderItemsTrashedRow, total *int)

	GetCachedOrderItems(ctx context.Context, orderID int) ([]*db.OrderItem, bool)
	SetCachedOrderItems(ctx context.Context, data []*db.OrderItem)
}

type OrderItemCommandCache interface {
	DeleteOrderItemCache(ctx context.Context, orderID int)
	DeleteOrderItemAllCache(ctx context.Context)
}
