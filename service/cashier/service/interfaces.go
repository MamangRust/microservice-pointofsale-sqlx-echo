package service

import (
	"context"

	db "github.com/MamangRust/microservice-point-of-sale-cashier/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type CashierQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllCashiers) ([]*db.GetCashiersRow, *int, error)
	FindById(ctx context.Context, cashierID int) (*db.Cashier, error)
	FindByActive(ctx context.Context, req *requests.FindAllCashiers) ([]*db.GetCashiersActiveRow, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllCashiers) ([]*db.GetCashiersTrashedRow, *int, error)
	FindByMerchant(ctx context.Context, req *requests.FindAllCashierMerchant) ([]*db.GetCashiersByMerchantRow, *int, error)
}

type CashierCommandService interface {
	CreateCashier(ctx context.Context, req *requests.CreateCashierRequest) (*db.Cashier, error)
	UpdateCashier(ctx context.Context, req *requests.UpdateCashierRequest) (*db.Cashier, error)
	TrashedCashier(ctx context.Context, cashierID int) (*db.Cashier, error)
	RestoreCashier(ctx context.Context, cashierID int) (*db.Cashier, error)
	DeleteCashierPermanent(ctx context.Context, cashierID int) (bool, error)
	RestoreAllCashier(ctx context.Context) (bool, error)
	DeleteAllCashierPermanent(ctx context.Context) (bool, error)
}
