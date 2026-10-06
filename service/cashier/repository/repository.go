package repository

import (
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	merchantadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/merchant"
	useradapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/user"
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	UserQuery      useradapter.QueryRepository
	MerchantQuery  merchantadapter.Repository
	CashierQuery   CashierQueryRepository
	CashierCommand CashierCommandRepository
}

// GuardOptions groups dependency guards by downstream gRPC dependency. Each
// non-nil slice is applied to the matching repository at construction (F6 §8.1
// poin 5 — deadline + circuit breaker + bulkhead per dependency).
type GuardOptions struct {
	User     []adapter.GuardOption
	Merchant []adapter.GuardOption
}

func NewRepositories(
	DB *sqlx.DB,
	userClient pbuser.UserServiceClient,
	merchantClient pbmerchant.MerchantServiceClient,
	guards ...GuardOptions,
) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &Repositories{
		UserQuery:      useradapter.New(userClient, g.User...),
		MerchantQuery:  merchantadapter.New(merchantClient, g.Merchant...),
		CashierQuery:   NewCashierQueryRepository(DB),
		CashierCommand: NewCashierCommandRepository(DB),
	}
}
