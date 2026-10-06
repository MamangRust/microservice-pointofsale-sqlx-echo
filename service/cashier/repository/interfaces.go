package repository

import (
	"context"

	db "github.com/MamangRust/microservice-point-of-sale-cashier/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type CashierQueryRepository interface {
	FindAllCashiers(ctx context.Context, req *requests.FindAllCashiers) ([]*db.GetCashiersRow, *int, error)
	FindById(ctx context.Context, cashier_id int) (*db.Cashier, error)
	FindByActive(ctx context.Context, req *requests.FindAllCashiers) ([]*db.GetCashiersActiveRow, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllCashiers) ([]*db.GetCashiersTrashedRow, *int, error)
	FindByMerchant(ctx context.Context, req *requests.FindAllCashierMerchant) ([]*db.GetCashiersByMerchantRow, *int, error)
}

type CashierCommandRepository interface {
	CreateCashier(ctx context.Context, request *requests.CreateCashierRequest) (*db.Cashier, error)
	UpdateCashier(ctx context.Context, request *requests.UpdateCashierRequest) (*db.Cashier, error)
	TrashedCashier(ctx context.Context, cashier_id int) (*db.Cashier, error)
	RestoreCashier(ctx context.Context, cashier_id int) (*db.Cashier, error)
	DeleteCashierPermanent(ctx context.Context, cashier_id int) (bool, error)
	RestoreAllCashier(ctx context.Context) (bool, error)
	DeleteAllCashierPermanent(ctx context.Context) (bool, error)
}
