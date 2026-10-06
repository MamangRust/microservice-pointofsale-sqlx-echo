package repository

import (
	"context"

	db "github.com/MamangRust/microservice-point-of-sale-order-item/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	orderitem_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/order_item_errors"
	"github.com/jmoiron/sqlx"
)

const createOrderItem = `INSERT INTO order_items (order_id, product_id, quantity, price)
VALUES ($1, $2, $3, $4)
RETURNING *`

const updateOrderItem = `UPDATE order_items
SET order_id = $2,
    product_id = $3,
    quantity = $4,
    price = $5,
    updated_at = CURRENT_TIMESTAMP
WHERE order_item_id = $1
  AND deleted_at IS NULL
RETURNING *`

const deleteOrderItem = `DELETE FROM order_items
WHERE order_item_id = $1
  AND deleted_at IS NULL`

type orderItemCommandRepository struct {
	db *sqlx.DB
}

func NewOrderItemCommandRepository(db *sqlx.DB) OrderItemCommandRepository {
	return &orderItemCommandRepository{
		db: db,
	}
}

func (r *orderItemCommandRepository) CreateOrderItem(ctx context.Context, req *requests.CreateOrderItemRecordRequest) (*db.OrderItem, error) {
	var orderItem db.OrderItem
	err := r.db.GetContext(ctx, &orderItem, createOrderItem,
		int32(req.OrderID),
		int32(req.ProductID),
		int32(req.Quantity),
		int32(req.Price),
	)
	if err != nil {
		return nil, orderitem_errors.ErrCreateOrderItem
	}

	return &orderItem, nil
}

func (r *orderItemCommandRepository) UpdateOrderItem(ctx context.Context, req *requests.UpdateOrderItemRecordRequest) (*db.OrderItem, error) {
	var orderItem db.OrderItem
	err := r.db.GetContext(ctx, &orderItem, updateOrderItem,
		int32(req.OrderItemID),
		int32(req.OrderID),
		int32(req.ProductID),
		int32(req.Quantity),
		int32(req.Price),
	)
	if err != nil {
		return nil, orderitem_errors.ErrUpdateOrderItem
	}

	return &orderItem, nil
}

func (r *orderItemCommandRepository) DeleteOrderItem(ctx context.Context, orderItemID int) error {
	if _, err := r.db.ExecContext(ctx, deleteOrderItem, int32(orderItemID)); err != nil {
		return orderitem_errors.ErrDeleteOrderItemPermanent
	}

	return nil
}
