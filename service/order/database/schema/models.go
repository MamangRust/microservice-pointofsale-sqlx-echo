// Hand-written models migrated from sqlc-generated source.
package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Cashier struct {
	CashierID  int32            `db:"cashier_id" json:"cashier_id"`
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	UserID     int32            `db:"user_id" json:"user_id"`
	Name       string           `db:"name" json:"name"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}

type Category struct {
	CategoryID   int32            `db:"category_id" json:"category_id"`
	Name         string           `db:"name" json:"name"`
	Description  *string          `db:"description" json:"description"`
	SlugCategory *string          `db:"slug_category" json:"slug_category"`
	CreatedAt    pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt    pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}

type Merchant struct {
	MerchantID   int32            `db:"merchant_id" json:"merchant_id"`
	UserID       int32            `db:"user_id" json:"user_id"`
	Name         string           `db:"name" json:"name"`
	Description  *string          `db:"description" json:"description"`
	Address      *string          `db:"address" json:"address"`
	ContactEmail *string          `db:"contact_email" json:"contact_email"`
	ContactPhone *string          `db:"contact_phone" json:"contact_phone"`
	Status       string           `db:"status" json:"status"`
	CreatedAt    pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt    pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}

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

type Transaction struct {
	TransactionID int32            `db:"transaction_id" json:"transaction_id"`
	OrderID       int32            `db:"order_id" json:"order_id"`
	MerchantID    int32            `db:"merchant_id" json:"merchant_id"`
	PaymentMethod string           `db:"payment_method" json:"payment_method"`
	Amount        int32            `db:"amount" json:"amount"`
	ChangeAmount  *int32           `db:"change_amount" json:"change_amount"`
	PaymentStatus string           `db:"payment_status" json:"payment_status"`
	CreatedAt     pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt     pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt     pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}
