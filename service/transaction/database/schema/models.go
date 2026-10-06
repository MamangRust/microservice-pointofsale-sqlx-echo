package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type Cashier struct {
	CashierID  int32            `json:"cashier_id" db:"cashier_id"`
	MerchantID int32            `json:"merchant_id" db:"merchant_id"`
	UserID     int32            `json:"user_id" db:"user_id"`
	Name       string           `json:"name" db:"name"`
	CreatedAt  pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt  pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt  pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}

type Merchant struct {
	MerchantID   int32            `json:"merchant_id" db:"merchant_id"`
	UserID       int32            `json:"user_id" db:"user_id"`
	Name         string           `json:"name" db:"name"`
	Description  *string          `json:"description" db:"description"`
	Address      *string          `json:"address" db:"address"`
	ContactEmail *string          `json:"contact_email" db:"contact_email"`
	ContactPhone *string          `json:"contact_phone" db:"contact_phone"`
	Status       string           `json:"status" db:"status"`
	CreatedAt    pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt    pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt    pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}

type Order struct {
	OrderID    int32            `json:"order_id" db:"order_id"`
	MerchantID int32            `json:"merchant_id" db:"merchant_id"`
	CashierID  int32            `json:"cashier_id" db:"cashier_id"`
	TotalPrice int64            `json:"total_price" db:"total_price"`
	CreatedAt  pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt  pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt  pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}

type OrderItem struct {
	OrderItemID int32            `json:"order_item_id" db:"order_item_id"`
	OrderID     int32            `json:"order_id" db:"order_id"`
	ProductID   int32            `json:"product_id" db:"product_id"`
	Quantity    int32            `json:"quantity" db:"quantity"`
	Price       int32            `json:"price" db:"price"`
	CreatedAt   pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt   pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt   pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}

type Transaction struct {
	TransactionID int32            `json:"transaction_id" db:"transaction_id"`
	OrderID       int32            `json:"order_id" db:"order_id"`
	MerchantID    int32            `json:"merchant_id" db:"merchant_id"`
	PaymentMethod string           `json:"payment_method" db:"payment_method"`
	Amount        int32            `json:"amount" db:"amount"`
	ChangeAmount  *int32           `json:"change_amount" db:"change_amount"`
	PaymentStatus string           `json:"payment_status" db:"payment_status"`
	CreatedAt     pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt     pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt     pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}
