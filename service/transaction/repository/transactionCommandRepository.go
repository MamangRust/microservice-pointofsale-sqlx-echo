package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/transaction_errors"
	db "github.com/MamangRust/microservice-point-of-sale-transacton/database/schema"
	"github.com/jmoiron/sqlx"
)

const createTransaction = `INSERT INTO transactions (
    merchant_id,
    payment_method,
    amount,
    change_amount,
    payment_status,
    order_id,
    created_at,
    updated_at,
    deleted_at
) VALUES (
    $1, 
    $2, 
    $3, 
    $4, 
    $5, 
    $6,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP,
    NULL
)
RETURNING transaction_id, order_id, merchant_id, payment_method, amount, change_amount, payment_status, created_at, updated_at, deleted_at
`

const deleteAllPermanentTransactions = `DELETE FROM transactions WHERE deleted_at IS NOT NULL
`

const deleteTransactionPermanently = `DELETE FROM transactions
WHERE
    transaction_id = $1
    AND deleted_at IS NOT NULL
`

const restoreAllTransactions = `UPDATE transactions
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL
`

const restoreTransaction = `UPDATE transactions
SET
    deleted_at = NULL
WHERE
    transaction_id = $1
    AND deleted_at IS NOT NULL
RETURNING
    transaction_id, order_id, merchant_id, payment_method, amount, change_amount, payment_status, created_at, updated_at, deleted_at
`

const trashTransaction = `UPDATE transactions
SET
    deleted_at = current_timestamp
WHERE
    transaction_id = $1
    AND deleted_at IS NULL
RETURNING
    transaction_id, order_id, merchant_id, payment_method, amount, change_amount, payment_status, created_at, updated_at, deleted_at
`

const updateTransaction = `UPDATE transactions
SET
    merchant_id = $2,
    payment_method = $3,
    amount = $4,
    change_amount = $5,
    payment_status = $6,
    order_id = $7,
    updated_at = CURRENT_TIMESTAMP
WHERE
    transaction_id = $1
    AND deleted_at IS NULL
RETURNING
    transaction_id, order_id, merchant_id, payment_method, amount, change_amount, payment_status, created_at, updated_at, deleted_at
`

type transactionCommandRepository struct {
	db *sqlx.DB
}

func NewTransactionCommandRepository(db *sqlx.DB) *transactionCommandRepository {
	return &transactionCommandRepository{
		db: db,
	}
}

func (r *transactionCommandRepository) CreateTransaction(ctx context.Context, request *requests.CreateTransactionRequest) (*db.Transaction, error) {
	return r.createTransaction(ctx, r.db, request)
}

// CreateTransactionInTx persists the transaction inside the given database
// transaction so the caller can commit the business write and its outbox event
// atomically (Phase 6 — transactional outbox).
func (r *transactionCommandRepository) CreateTransactionInTx(ctx context.Context, tx *sqlx.Tx, request *requests.CreateTransactionRequest) (*db.Transaction, error) {
	return r.createTransaction(ctx, tx, request)
}

func (r *transactionCommandRepository) createTransaction(ctx context.Context, q sqlx.QueryerContext, request *requests.CreateTransactionRequest) (*db.Transaction, error) {
	var transaction db.Transaction
	err := sqlx.GetContext(ctx, q, &transaction, createTransaction,
		int32(request.MerchantID),
		request.PaymentMethod,
		int32(request.Amount),
		nil,
		*request.PaymentStatus,
		int32(request.OrderID),
	)
	if err != nil {
		return nil, transaction_errors.ErrCreateTransaction
	}

	return &transaction, nil
}

func (r *transactionCommandRepository) UpdateTransaction(ctx context.Context, request *requests.UpdateTransactionRequest) (*db.Transaction, error) {
	var transaction db.Transaction
	err := sqlx.GetContext(ctx, r.db, &transaction, updateTransaction,
		int32(*request.TransactionID),
		int32(request.MerchantID),
		request.PaymentMethod,
		int32(request.Amount),
		nil,
		*request.PaymentStatus,
		int32(request.OrderID),
	)
	if err != nil {
		return nil, transaction_errors.ErrUpdateTransaction
	}

	return &transaction, nil
}

func (r *transactionCommandRepository) TrashTransaction(ctx context.Context, transaction_id int) (*db.Transaction, error) {
	var transaction db.Transaction
	err := sqlx.GetContext(ctx, r.db, &transaction, trashTransaction, int32(transaction_id))
	if err != nil {
		return nil, transaction_errors.ErrTrashTransaction
	}

	return &transaction, nil
}

func (r *transactionCommandRepository) RestoreTransaction(ctx context.Context, transaction_id int) (*db.Transaction, error) {
	var transaction db.Transaction
	err := sqlx.GetContext(ctx, r.db, &transaction, restoreTransaction, int32(transaction_id))
	if err != nil {
		return nil, transaction_errors.ErrRestoreTransaction
	}

	return &transaction, nil
}

func (r *transactionCommandRepository) DeleteTransactionPermanently(ctx context.Context, transaction_id int) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteTransactionPermanently, int32(transaction_id)); err != nil {
		return false, transaction_errors.ErrDeleteTransactionPermanently
	}

	return true, nil
}

func (r *transactionCommandRepository) RestoreAllTransactions(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, restoreAllTransactions); err != nil {
		return false, transaction_errors.ErrRestoreAllTransactions
	}

	return true, nil
}

func (r *transactionCommandRepository) DeleteAllTransactionPermanent(ctx context.Context) (bool, error) {
	if _, err := r.db.ExecContext(ctx, deleteAllPermanentTransactions); err != nil {
		return false, transaction_errors.ErrDeleteAllTransactionPermanent
	}

	return true, nil
}
