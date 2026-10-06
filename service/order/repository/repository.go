package repository

import (
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	cashieradapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/cashier"
	merchantadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/merchant"
	orderitemadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/orderitem"
	productadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/product"
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	CashierQuery     cashieradapter.QueryRepository
	MerchantQuery    merchantadapter.Repository
	ProductQuery     productadapter.QueryRepository
	ProductCommand   productadapter.CommandRepository
	OrderQuery       OrderQueryRepository
	OrderCommand     OrderCommandRepository
	OrderItemQuery   orderitemadapter.QueryRepository
	OrderItemCommand orderitemadapter.CommandRepository
}

// GuardOptions groups dependency guards by downstream gRPC dependency. Each
// non-nil slice is applied to the matching repository at construction (F6 §8.1
// poin 5 — deadline + circuit breaker + bulkhead per dependency).
type GuardOptions struct {
	Cashier   []adapter.GuardOption
	Merchant  []adapter.GuardOption
	Product   []adapter.GuardOption
	OrderItem []adapter.GuardOption
}

func NewRepositories(
	DB *sqlx.DB,
	cashierClient pbcashier.CashierServiceClient,
	merchantClient pbmerchant.MerchantServiceClient,
	productClient pbproduct.ProductServiceClient,
	orderItemClient pborderitem.OrderItemServiceClient,
	guards ...GuardOptions,
) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	// Product and order item each expose one concrete adapter instance that
	// satisfies both the query and the command interface, so the same instance
	// is shared across the two fields.
	productRepo := productadapter.New(productClient, g.Product...)
	orderItemRepo := orderitemadapter.New(orderItemClient, g.OrderItem...)

	return &Repositories{
		CashierQuery:     cashieradapter.New(cashierClient, g.Cashier...),
		MerchantQuery:    merchantadapter.New(merchantClient, g.Merchant...),
		ProductQuery:     productRepo,
		ProductCommand:   productRepo,
		OrderQuery:       NewOrderQueryRepository(DB),
		OrderCommand:     NewOrderCommandRepository(DB),
		OrderItemQuery:   orderItemRepo,
		OrderItemCommand: orderItemRepo,
	}
}
