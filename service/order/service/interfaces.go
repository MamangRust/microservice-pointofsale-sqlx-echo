package service

import (
	"context"

	db "github.com/MamangRust/microservice-point-of-sale-order/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type OrderQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllOrders) ([]*db.GetOrdersRow, *int, error)
	FindById(ctx context.Context, orderID int) (*db.Order, error)
	FindByActive(ctx context.Context, req *requests.FindAllOrders) ([]*db.GetOrdersActiveRow, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllOrders) ([]*db.GetOrdersTrashedRow, *int, error)
	FindByMerchant(ctx context.Context, req *requests.FindAllOrderMerchant) ([]*db.GetOrdersByMerchantRow, *int, error)
}

type OrderCommandService interface {
	CreateOrder(ctx context.Context, req *requests.CreateOrderRequest) (*db.Order, error)
	UpdateOrder(ctx context.Context, req *requests.UpdateOrderRequest) (*db.Order, error)
	TrashedOrder(ctx context.Context, orderID int) (*db.Order, error)
	RestoreOrder(ctx context.Context, orderID int) (*db.Order, error)
	DeleteOrderPermanent(ctx context.Context, orderID int) (bool, error)
	RestoreAllOrder(ctx context.Context) (bool, error)
	DeleteAllOrderPermanent(ctx context.Context) (bool, error)
}
