package service

import (
	"context"

	db "github.com/MamangRust/microservice-point-of-sale-order-item/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type OrderItemQueryService interface {
	FindAllOrderItems(ctx context.Context, req *requests.FindAllOrderItems) ([]*db.GetOrderItemsRow, *int, error)
	FindByActive(ctx context.Context, req *requests.FindAllOrderItems) ([]*db.GetOrderItemsActiveRow, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllOrderItems) ([]*db.GetOrderItemsTrashedRow, *int, error)
	FindOrderItemByOrder(ctx context.Context, orderID int) ([]*db.OrderItem, error)
}

type OrderItemCommandService interface {
	CreateOrderItem(ctx context.Context, req *requests.CreateOrderItemRecordRequest) (*db.OrderItem, error)
	UpdateOrderItem(ctx context.Context, req *requests.UpdateOrderItemRecordRequest) (*db.OrderItem, error)
	DeleteOrderItem(ctx context.Context, orderItemID int) error
}
