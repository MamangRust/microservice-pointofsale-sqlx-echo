package repository

import (
	"context"

	db "github.com/MamangRust/microservice-point-of-sale-order/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type OrderQueryRepository interface {
	FindAllOrders(ctx context.Context, req *requests.FindAllOrders) ([]*db.GetOrdersRow, *int, error)
	FindByActive(ctx context.Context, req *requests.FindAllOrders) ([]*db.GetOrdersActiveRow, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllOrders) ([]*db.GetOrdersTrashedRow, *int, error)
	FindByMerchant(ctx context.Context, req *requests.FindAllOrderMerchant) ([]*db.GetOrdersByMerchantRow, *int, error)
	FindById(ctx context.Context, orderID int) (*db.Order, error)
	FindByTrashedId(ctx context.Context, orderID int) (*db.Order, error)
}

type OrderCommandRepository interface {
	DeleteOrder(ctx context.Context, orderID int) error
	CreateOrder(ctx context.Context, request *requests.CreateOrderRecordRequest) (*db.Order, error)
	UpdateOrder(ctx context.Context, request *requests.UpdateOrderRecordRequest) (*db.Order, error)
	FindAllTrashed(ctx context.Context) ([]*db.Order, error)
	TrashedOrder(ctx context.Context, orderID int) (*db.Order, error)
	RestoreOrder(ctx context.Context, orderID int) (*db.Order, error)
	DeleteOrderPermanent(ctx context.Context, orderID int) (bool, error)
	DeleteAllOrderPermanent(ctx context.Context) (bool, error)
}
