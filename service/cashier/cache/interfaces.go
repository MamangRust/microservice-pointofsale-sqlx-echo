package mencache

import (
	"context"

	db "github.com/MamangRust/microservice-point-of-sale-cashier/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type CashierQueryCache interface {
	GetCachedCashiersCache(ctx context.Context, req *requests.FindAllCashiers) ([]*db.GetCashiersRow, *int, bool)
	SetCachedCashiersCache(ctx context.Context, req *requests.FindAllCashiers, res []*db.GetCashiersRow, total *int)

	GetCachedCashier(ctx context.Context, cashierID int) (*db.Cashier, bool)
	SetCachedCashier(ctx context.Context, res *db.Cashier)

	GetCachedCashiersActive(ctx context.Context, req *requests.FindAllCashiers) ([]*db.GetCashiersActiveRow, *int, bool)
	SetCachedCashiersActive(ctx context.Context, req *requests.FindAllCashiers, res []*db.GetCashiersActiveRow, total *int)

	GetCachedCashiersTrashed(ctx context.Context, req *requests.FindAllCashiers) ([]*db.GetCashiersTrashedRow, *int, bool)
	SetCachedCashiersTrashed(ctx context.Context, req *requests.FindAllCashiers, res []*db.GetCashiersTrashedRow, total *int)

	GetCachedCashiersByMerchant(ctx context.Context, req *requests.FindAllCashierMerchant) ([]*db.GetCashiersByMerchantRow, *int, bool)
	SetCachedCashiersByMerchant(ctx context.Context, req *requests.FindAllCashierMerchant, res []*db.GetCashiersByMerchantRow, total *int)
}

type CashierCommandCache interface {
	DeleteCashierCache(ctx context.Context, id int)
	DeleteCashierListCache(ctx context.Context)
}
