package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	db "github.com/MamangRust/microservice-point-of-sale-transacton/database/schema"
	"github.com/jmoiron/sqlx"
)

type TransactionQueryRepository interface {
	FindAllTransactions(ctx context.Context, req *requests.FindAllTransaction) ([]*db.GetTransactionsRow, *int, error)
	FindByActive(ctx context.Context, req *requests.FindAllTransaction) ([]*db.GetTransactionsActiveRow, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllTransaction) ([]*db.GetTransactionsTrashedRow, *int, error)
	FindByMerchant(ctx context.Context, req *requests.FindAllTransactionByMerchant) ([]*db.GetTransactionByMerchantRow, *int, error)
	FindById(ctx context.Context, transaction_id int) (*db.Transaction, error)
	FindByOrderId(ctx context.Context, order_id int) (*db.Transaction, error)
}

type TransactionCommandRepository interface {
	CreateTransaction(ctx context.Context, request *requests.CreateTransactionRequest) (*db.Transaction, error)
	CreateTransactionInTx(ctx context.Context, tx *sqlx.Tx, request *requests.CreateTransactionRequest) (*db.Transaction, error)
	UpdateTransaction(ctx context.Context, request *requests.UpdateTransactionRequest) (*db.Transaction, error)
	TrashTransaction(ctx context.Context, transaction_id int) (*db.Transaction, error)
	RestoreTransaction(ctx context.Context, transaction_id int) (*db.Transaction, error)
	DeleteTransactionPermanently(ctx context.Context, transaction_id int) (bool, error)
	RestoreAllTransactions(ctx context.Context) (bool, error)
	DeleteAllTransactionPermanent(ctx context.Context) (bool, error)
}
