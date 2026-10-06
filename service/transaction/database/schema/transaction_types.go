package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type CreateTransactionParams struct {
	MerchantID    int32  `json:"merchant_id" db:"merchant_id"`
	PaymentMethod string `json:"payment_method" db:"payment_method"`
	Amount        int32  `json:"amount" db:"amount"`
	ChangeAmount  *int32 `json:"change_amount" db:"change_amount"`
	PaymentStatus string `json:"payment_status" db:"payment_status"`
	OrderID       int32  `json:"order_id" db:"order_id"`
}

type GetMonthlyAmountTransactionFailedParams struct {
	Column1 time.Time `json:"column_1" db:"column_1"`
	Column2 time.Time `json:"column_2" db:"column_2"`
	Column3 time.Time `json:"column_3" db:"column_3"`
	Column4 time.Time `json:"column_4" db:"column_4"`
}

type GetMonthlyAmountTransactionFailedRow struct {
	Year        string `json:"year" db:"year"`
	Month       string `json:"month" db:"month"`
	TotalFailed int64  `json:"total_failed" db:"total_failed"`
	TotalAmount int32  `json:"total_amount" db:"total_amount"`
}

type GetMonthlyAmountTransactionFailedByMerchantParams struct {
	Column1    time.Time `json:"column_1" db:"column_1"`
	Column2    time.Time `json:"column_2" db:"column_2"`
	Column3    time.Time `json:"column_3" db:"column_3"`
	Column4    time.Time `json:"column_4" db:"column_4"`
	MerchantID int32     `json:"merchant_id" db:"merchant_id"`
}

type GetMonthlyAmountTransactionFailedByMerchantRow struct {
	Year        string `json:"year" db:"year"`
	Month       string `json:"month" db:"month"`
	TotalFailed int64  `json:"total_failed" db:"total_failed"`
	TotalAmount int32  `json:"total_amount" db:"total_amount"`
}

type GetMonthlyAmountTransactionSuccessParams struct {
	Column1 time.Time `json:"column_1" db:"column_1"`
	Column2 time.Time `json:"column_2" db:"column_2"`
	Column3 time.Time `json:"column_3" db:"column_3"`
	Column4 time.Time `json:"column_4" db:"column_4"`
}

type GetMonthlyAmountTransactionSuccessRow struct {
	Year         string `json:"year" db:"year"`
	Month        string `json:"month" db:"month"`
	TotalSuccess int64  `json:"total_success" db:"total_success"`
	TotalAmount  int32  `json:"total_amount" db:"total_amount"`
}

type GetMonthlyAmountTransactionSuccessByMerchantParams struct {
	Column1    time.Time `json:"column_1" db:"column_1"`
	Column2    time.Time `json:"column_2" db:"column_2"`
	Column3    time.Time `json:"column_3" db:"column_3"`
	Column4    time.Time `json:"column_4" db:"column_4"`
	MerchantID int32     `json:"merchant_id" db:"merchant_id"`
}

type GetMonthlyAmountTransactionSuccessByMerchantRow struct {
	Year         string `json:"year" db:"year"`
	Month        string `json:"month" db:"month"`
	TotalSuccess int64  `json:"total_success" db:"total_success"`
	TotalAmount  int32  `json:"total_amount" db:"total_amount"`
}

type GetMonthlyTransactionMethodsByMerchantFailedParams struct {
	Column1    time.Time `json:"column_1" db:"column_1"`
	Column2    time.Time `json:"column_2" db:"column_2"`
	Column3    time.Time `json:"column_3" db:"column_3"`
	Column4    time.Time `json:"column_4" db:"column_4"`
	MerchantID int32     `json:"merchant_id" db:"merchant_id"`
}

type GetMonthlyTransactionMethodsByMerchantFailedRow struct {
	Month             string  `json:"month" db:"month"`
	PaymentMethod     string  `json:"payment_method" db:"payment_method"`
	TotalTransactions int64   `json:"total_transactions" db:"total_transactions"`
	TotalAmount       float64 `json:"total_amount" db:"total_amount"`
}

type GetMonthlyTransactionMethodsByMerchantSuccessParams struct {
	Column1    time.Time `json:"column_1" db:"column_1"`
	Column2    time.Time `json:"column_2" db:"column_2"`
	Column3    time.Time `json:"column_3" db:"column_3"`
	Column4    time.Time `json:"column_4" db:"column_4"`
	MerchantID int32     `json:"merchant_id" db:"merchant_id"`
}

type GetMonthlyTransactionMethodsByMerchantSuccessRow struct {
	Month             string  `json:"month" db:"month"`
	PaymentMethod     string  `json:"payment_method" db:"payment_method"`
	TotalTransactions int64   `json:"total_transactions" db:"total_transactions"`
	TotalAmount       float64 `json:"total_amount" db:"total_amount"`
}

type GetMonthlyTransactionMethodsFailedParams struct {
	Column1 time.Time `json:"column_1" db:"column_1"`
	Column2 time.Time `json:"column_2" db:"column_2"`
	Column3 time.Time `json:"column_3" db:"column_3"`
	Column4 time.Time `json:"column_4" db:"column_4"`
}

type GetMonthlyTransactionMethodsFailedRow struct {
	Month             string  `json:"month" db:"month"`
	PaymentMethod     string  `json:"payment_method" db:"payment_method"`
	TotalTransactions int64   `json:"total_transactions" db:"total_transactions"`
	TotalAmount       float64 `json:"total_amount" db:"total_amount"`
}

type GetMonthlyTransactionMethodsSuccessParams struct {
	Column1 time.Time `json:"column_1" db:"column_1"`
	Column2 time.Time `json:"column_2" db:"column_2"`
	Column3 time.Time `json:"column_3" db:"column_3"`
	Column4 time.Time `json:"column_4" db:"column_4"`
}

type GetMonthlyTransactionMethodsSuccessRow struct {
	Month             string  `json:"month" db:"month"`
	PaymentMethod     string  `json:"payment_method" db:"payment_method"`
	TotalTransactions int64   `json:"total_transactions" db:"total_transactions"`
	TotalAmount       float64 `json:"total_amount" db:"total_amount"`
}

type GetTransactionByMerchantParams struct {
	Column1 string `json:"column_1" db:"column_1"`
	Column2 int32  `json:"column_2" db:"column_2"`
	Limit   int32  `json:"limit" db:"limit"`
	Offset  int32  `json:"offset" db:"offset"`
}

type GetTransactionByMerchantRow struct {
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
	TotalCount    int64            `json:"total_count" db:"total_count"`
}

type GetTransactionsParams struct {
	Column1 string `json:"column_1" db:"column_1"`
	Limit   int32  `json:"limit" db:"limit"`
	Offset  int32  `json:"offset" db:"offset"`
}

type GetTransactionsRow struct {
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
	TotalCount    int64            `json:"total_count" db:"total_count"`
}

type GetTransactionsActiveParams struct {
	Column1 string `json:"column_1" db:"column_1"`
	Limit   int32  `json:"limit" db:"limit"`
	Offset  int32  `json:"offset" db:"offset"`
}

type GetTransactionsActiveRow struct {
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
	TotalCount    int64            `json:"total_count" db:"total_count"`
}

type GetTransactionsTrashedParams struct {
	Column1 string `json:"column_1" db:"column_1"`
	Limit   int32  `json:"limit" db:"limit"`
	Offset  int32  `json:"offset" db:"offset"`
}

type GetTransactionsTrashedRow struct {
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
	TotalCount    int64            `json:"total_count" db:"total_count"`
}

type GetYearlyAmountTransactionFailedRow struct {
	Year        string `json:"year" db:"year"`
	TotalFailed int32  `json:"total_failed" db:"total_failed"`
	TotalAmount int32  `json:"total_amount" db:"total_amount"`
}

type GetYearlyAmountTransactionFailedByMerchantParams struct {
	Column1    int32 `json:"column_1" db:"column_1"`
	MerchantID int32 `json:"merchant_id" db:"merchant_id"`
}

type GetYearlyAmountTransactionFailedByMerchantRow struct {
	Year        string `json:"year" db:"year"`
	TotalFailed int32  `json:"total_failed" db:"total_failed"`
	TotalAmount int32  `json:"total_amount" db:"total_amount"`
}

type GetYearlyAmountTransactionSuccessRow struct {
	Year         string `json:"year" db:"year"`
	TotalSuccess int32  `json:"total_success" db:"total_success"`
	TotalAmount  int32  `json:"total_amount" db:"total_amount"`
}

type GetYearlyAmountTransactionSuccessByMerchantParams struct {
	Column1    int32 `json:"column_1" db:"column_1"`
	MerchantID int32 `json:"merchant_id" db:"merchant_id"`
}

type GetYearlyAmountTransactionSuccessByMerchantRow struct {
	Year         string `json:"year" db:"year"`
	TotalSuccess int32  `json:"total_success" db:"total_success"`
	TotalAmount  int32  `json:"total_amount" db:"total_amount"`
}

type GetYearlyTransactionMethodsByMerchantFailedParams struct {
	Column1    time.Time `json:"column_1" db:"column_1"`
	MerchantID int32     `json:"merchant_id" db:"merchant_id"`
}

type GetYearlyTransactionMethodsByMerchantFailedRow struct {
	Year              string  `json:"year" db:"year"`
	PaymentMethod     string  `json:"payment_method" db:"payment_method"`
	TotalTransactions int64   `json:"total_transactions" db:"total_transactions"`
	TotalAmount       float64 `json:"total_amount" db:"total_amount"`
}

type GetYearlyTransactionMethodsByMerchantSuccessParams struct {
	Column1    time.Time `json:"column_1" db:"column_1"`
	MerchantID int32     `json:"merchant_id" db:"merchant_id"`
}

type GetYearlyTransactionMethodsByMerchantSuccessRow struct {
	Year              string  `json:"year" db:"year"`
	PaymentMethod     string  `json:"payment_method" db:"payment_method"`
	TotalTransactions int64   `json:"total_transactions" db:"total_transactions"`
	TotalAmount       float64 `json:"total_amount" db:"total_amount"`
}

type GetYearlyTransactionMethodsFailedRow struct {
	Year              string  `json:"year" db:"year"`
	PaymentMethod     string  `json:"payment_method" db:"payment_method"`
	TotalTransactions int64   `json:"total_transactions" db:"total_transactions"`
	TotalAmount       float64 `json:"total_amount" db:"total_amount"`
}

type GetYearlyTransactionMethodsSuccessRow struct {
	Year              string  `json:"year" db:"year"`
	PaymentMethod     string  `json:"payment_method" db:"payment_method"`
	TotalTransactions int64   `json:"total_transactions" db:"total_transactions"`
	TotalAmount       float64 `json:"total_amount" db:"total_amount"`
}

type UpdateTransactionParams struct {
	TransactionID int32  `json:"transaction_id" db:"transaction_id"`
	MerchantID    int32  `json:"merchant_id" db:"merchant_id"`
	PaymentMethod string `json:"payment_method" db:"payment_method"`
	Amount        int32  `json:"amount" db:"amount"`
	ChangeAmount  *int32 `json:"change_amount" db:"change_amount"`
	PaymentStatus string `json:"payment_status" db:"payment_status"`
	OrderID       int32  `json:"order_id" db:"order_id"`
}
