package repository

import (
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	cashieradapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/cashier"
	merchantadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/merchant"
	orderadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/order"
	orderitemadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/orderitem"
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	CashierQuery                 cashieradapter.QueryRepository
	MerchantQuery                merchantadapter.Repository
	OrderQuery                   orderadapter.QueryRepository
	OrderItemQuery               orderitemadapter.QueryRepository
	TransactionCommandRepository TransactionCommandRepository
	TransactionQueryRepository   TransactionQueryRepository
}

// GuardOptions groups dependency guards by downstream gRPC dependency. Each
// non-nil slice is applied to the matching repository at construction (F6 §8.1
// poin 5 — deadline + circuit breaker + bulkhead per dependency).
type GuardOptions struct {
	Cashier   []adapter.GuardOption
	Merchant  []adapter.GuardOption
	Order     []adapter.GuardOption
	OrderItem []adapter.GuardOption
}

func NewRepositories(
	DB *sqlx.DB,
	cashierClient pbcashier.CashierServiceClient,
	merchantClient pbmerchant.MerchantServiceClient,
	orderClient pborder.OrderServiceClient,
	orderItemClient pborderitem.OrderItemServiceClient,
	guards ...GuardOptions,
) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}
	return &Repositories{
		CashierQuery:                 cashieradapter.New(cashierClient, g.Cashier...),
		MerchantQuery:                merchantadapter.New(merchantClient, g.Merchant...),
		OrderQuery:                   orderadapter.New(orderClient, g.Order...),
		OrderItemQuery:               orderitemadapter.New(orderItemClient, g.OrderItem...),
		TransactionCommandRepository: NewTransactionCommandRepository(DB),
		TransactionQueryRepository:   NewTransactionQueryRepository(DB),
	}
}
