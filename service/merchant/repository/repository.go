package repository

import (
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	useradapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/user"
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	MerchantQuery           MerchantQueryRepository
	MerchantCommand         MerchantCommandRepository
	MerchantDocumentCommand MerchantDocumentCommandRepository
	MerchantDocumentQuery   MerchantDocumentQueryRepository
	UserQuery               useradapter.QueryRepository
}

// GuardOptions groups dependency guards by downstream gRPC dependency. Each
// non-nil slice is applied to the matching repository at construction (F6 §8.1
// poin 5 — deadline + circuit breaker + bulkhead per dependency).
type GuardOptions struct {
	User []adapter.GuardOption
}

func NewRepositories(
	DB *sqlx.DB,
	userClient pbuser.UserServiceClient,
	guards ...GuardOptions,
) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &Repositories{
		MerchantQuery:           NewMerchantQueryRepository(DB),
		MerchantCommand:         NewMerchantCommandRepository(DB),
		MerchantDocumentCommand: NewMerchantDocumentCommandRepository(DB),
		MerchantDocumentQuery:   NewMerchantDocumentQueryRepository(DB),
		UserQuery:               useradapter.New(userClient, g.User...),
	}
}
