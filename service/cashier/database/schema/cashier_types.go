// Code migrated from sqlc-generated sources; domain types kept for repository layer.
// source: cashiers.sql

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateCashierParams struct {
	MerchantID int32  `db:"merchant_id" json:"merchant_id"`
	UserID     int32  `db:"user_id" json:"user_id"`
	Name       string `db:"name" json:"name"`
}

type UpdateCashierParams struct {
	CashierID int32  `db:"cashier_id" json:"cashier_id"`
	Name      string `db:"name" json:"name"`
}

type GetCashiersParams struct {
	Column1 string `db:"column1" json:"column_1"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetCashiersRow struct {
	CashierID  int32            `db:"cashier_id" json:"cashier_id"`
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	UserID     int32            `db:"user_id" json:"user_id"`
	Name       string           `db:"name" json:"name"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetCashiersActiveParams struct {
	Column1 string `db:"column1" json:"column_1"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetCashiersActiveRow struct {
	CashierID  int32            `db:"cashier_id" json:"cashier_id"`
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	UserID     int32            `db:"user_id" json:"user_id"`
	Name       string           `db:"name" json:"name"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetCashiersTrashedParams struct {
	Column1 string `db:"column1" json:"column_1"`
	Limit   int32  `db:"limit" json:"limit"`
	Offset  int32  `db:"offset" json:"offset"`
}

type GetCashiersTrashedRow struct {
	CashierID  int32            `db:"cashier_id" json:"cashier_id"`
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	UserID     int32            `db:"user_id" json:"user_id"`
	Name       string           `db:"name" json:"name"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetCashiersByMerchantParams struct {
	MerchantID int32  `db:"merchant_id" json:"merchant_id"`
	Column2    string `db:"column2" json:"column_2"`
	Limit      int32  `db:"limit" json:"limit"`
	Offset     int32  `db:"offset" json:"offset"`
}

type GetCashiersByMerchantRow struct {
	CashierID  int32            `db:"cashier_id" json:"cashier_id"`
	MerchantID int32            `db:"merchant_id" json:"merchant_id"`
	UserID     int32            `db:"user_id" json:"user_id"`
	Name       string           `db:"name" json:"name"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}
