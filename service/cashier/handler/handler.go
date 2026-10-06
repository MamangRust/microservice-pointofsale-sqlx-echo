package handler

import (
	"github.com/MamangRust/microservice-point-of-sale-cashier/service"
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
)

type Deps struct {
	Service *service.Service
	Logger  logger.LoggerInterface
}

type Handler struct {
	Cashier pbcashier.CashierServiceServer
}

func NewHandler(deps *Deps) *Handler {
	return &Handler{
		Cashier: NewCashierHandleGrpc(
			deps.Service,
			deps.Logger,
		),
	}
}
