package service

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	db "github.com/MamangRust/microservice-point-of-sale-transacton/database/schema"
)

// EmailEventPublisher publishes email notification events to Kafka topics.
// Implemented by *kafka.Kafka; defined as an interface so it can be faked in
// unit tests (graceful degradation of the email notification path).
// ctx carries the OpenTelemetry trace context, which is injected into the
// Kafka message headers so the consumer continues the same trace.
type EmailEventPublisher interface {
	SendMessage(ctx context.Context, topic string, key string, value []byte) error
}

type TransactionQueryService interface {
	FindAllTransactions(ctx context.Context, req *requests.FindAllTransaction) ([]*db.GetTransactionsRow, *int, error)
	FindByMerchant(ctx context.Context, req *requests.FindAllTransactionByMerchant) ([]*db.GetTransactionByMerchantRow, *int, error)
	FindByActive(ctx context.Context, req *requests.FindAllTransaction) ([]*db.GetTransactionsActiveRow, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllTransaction) ([]*db.GetTransactionsTrashedRow, *int, error)
	FindById(ctx context.Context, transactionID int) (*db.Transaction, error)
	FindByOrderId(ctx context.Context, orderID int) (*db.Transaction, error)
}

type TransactionCommandService interface {
	CreateTransaction(ctx context.Context, req *requests.CreateTransactionRequest) (*db.Transaction, error)
	UpdateTransaction(ctx context.Context, req *requests.UpdateTransactionRequest) (*db.Transaction, error)
	TrashedTransaction(ctx context.Context, transaction_id int) (*db.Transaction, error)
	RestoreTransaction(ctx context.Context, transaction_id int) (*db.Transaction, error)
	DeleteTransactionPermanently(ctx context.Context, transactionID int) (bool, error)
	RestoreAllTransactions(ctx context.Context) (bool, error)
	DeleteAllTransactionPermanent(ctx context.Context) (bool, error)
}
