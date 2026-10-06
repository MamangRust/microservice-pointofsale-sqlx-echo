// Hand-written models migrated from sqlc-generated source.
package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Order struct {
	OrderID    int32            `db:"order_id" json:"order_id"`
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	CashierID  int32            `db:"cashier_id" json:"cashier_id"`
	TotalPrice int64            `db:"total_price" json:"total_price"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}

type OrderItem struct {
	OrderItemID int32            `db:"order_item_id" json:"order_item_id"`
	OrderID     int32            `db:"order_id" json:"order_id"`
	ProductID   int32            `db:"product_id" json:"product_id"`
	Quantity    int32            `db:"quantity" json:"quantity"`
	Price       int32            `db:"price" json:"price"`
	CreatedAt   pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt   pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt   pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}

type Product struct {
	ProductID    int32            `db:"product_id" json:"product_id"`
	MerchantID   int32            `db:"merchant_id" json:"merchant_id"`
	CategoryID   int32            `db:"category_id" json:"category_id"`
	Name         string           `db:"name" json:"name"`
	Description  *string          `db:"description" json:"description"`
	Price        int32            `db:"price" json:"price"`
	CountInStock int32            `db:"count_in_stock" json:"count_in_stock"`
	Brand        *string          `db:"brand" json:"brand"`
	Weight       *int32           `db:"weight" json:"weight"`
	SlugProduct  *string          `db:"slug_product" json:"slug_product"`
	ImageProduct *string          `db:"image_product" json:"image_product"`
	Barcode      *string          `db:"barcode" json:"barcode"`
	CreatedAt    pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt    pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}
