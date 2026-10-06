// Code migrated from sqlc-generated sources; domain types kept for repository layer.

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

type User struct {
	UserID           int32            `db:"user_id" json:"user_id"`
	Firstname        string           `db:"firstname" json:"firstname"`
	Lastname         string           `db:"lastname" json:"lastname"`
	Email            string           `db:"email" json:"email"`
	Password         string           `db:"password" json:"password"`
	VerificationCode string           `db:"verification_code" json:"verification_code"`
	IsVerified       *bool            `db:"is_verified" json:"is_verified"`
	CreatedAt        pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt        pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt        pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}
