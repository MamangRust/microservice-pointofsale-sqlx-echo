// Package transaction adapts the Transaction service gRPC API into the shared
// domain model. Query and bulk are exposed as separate interfaces so consumers
// can depend on the narrower one.
package transaction

import (
	"context"

	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter/convert"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	transaction_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/transaction_errors"
)

// QueryRepository reads transactions from the Transaction service over gRPC.
type QueryRepository interface {
	FindById(ctx context.Context, id int) (*models.Transaction, error)
}

// BulkRepository enumerates transactions page by page. It is deliberately a
// separate narrow interface: offline consumers (the stats backfill) need the
// whole collection, while ordinary consumers only ever read one transaction at
// a time and should not depend on the extra method.
type BulkRepository interface {
	FindAll(ctx context.Context, page, pageSize int) ([]*models.Transaction, int, error)
}

// Repository implements both QueryRepository and BulkRepository over a single
// gRPC client, so consumers can share one instance across both roles.
type Repository struct {
	client pbtransaction.TransactionServiceClient
	guard  *resilience.DependencyGuard
}

// New builds a transaction adapter. Passing zero options leaves the guard nil,
// which makes DependencyGuard.Call a plain passthrough.
func New(client pbtransaction.TransactionServiceClient, opts ...adapter.GuardOption) *Repository {
	r := &Repository{client: client}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (r *Repository) SetGuard(g *resilience.DependencyGuard) {
	r.guard = g
}

func (r *Repository) FindById(ctx context.Context, id int) (*models.Transaction, error) {
	var resp *pbtransaction.ApiResponseTransaction
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindById(ctx, &pbtransaction.FindByIdTransactionRequest{Id: int32(id)})
		return callErr
	})
	if err != nil || resp == nil || resp.Data == nil {
		return nil, transaction_errors.ErrFindById
	}

	return mapTransaction(resp.Data), nil
}

// FindAll returns one page of active transactions plus the total number of
// active rows, so callers can walk the collection page by page.
func (r *Repository) FindAll(ctx context.Context, page, pageSize int) ([]*models.Transaction, int, error) {
	var resp *pbtransaction.ApiResponsePaginationTransaction
	err := r.guard.Call(ctx, func(ctx context.Context) error {
		var callErr error
		resp, callErr = r.client.FindAll(ctx, &pbtransaction.FindAllTransactionRequest{
			Page:     int32(page),
			PageSize: int32(pageSize),
		})
		return callErr
	})
	if err != nil || resp == nil {
		return nil, 0, transaction_errors.ErrFindAllTransactions
	}

	transactions := make([]*models.Transaction, 0, len(resp.Data))
	for _, t := range resp.Data {
		if t == nil {
			continue
		}
		transactions = append(transactions, mapTransaction(t))
	}

	total := 0
	if resp.Pagination != nil {
		total = int(resp.Pagination.TotalRecords)
	}
	return transactions, total, nil
}

// mapTransaction converts a proto transaction into the shared domain model.
// The transaction row does not carry cashier_id — the transaction service
// resolves it through the order it belongs to — so the field is left zero for
// the caller to fill in.
func mapTransaction(t *pbtransaction.TransactionResponse) *models.Transaction {
	if t == nil {
		return nil
	}
	return &models.Transaction{
		TransactionID: t.Id,
		OrderID:       t.OrderId,
		MerchantID:    t.MerchantId,
		PaymentMethod: t.PaymentMethod,
		Amount:        t.Amount,
		PaymentStatus: convert.NullableString(t.PaymentStatus),
		CreatedAt:     convert.TimePtr(t.CreatedAt),
		UpdatedAt:     convert.TimePtr(t.UpdatedAt),
	}
}
