package handler

import (
	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-transacton/service"
)

type Deps struct {
	Service *service.Service
	Logger  logger.LoggerInterface
}

type Handler struct {
	Transaction pbtransaction.TransactionServiceServer
}

func NewHandler(deps *Deps) *Handler {
	return &Handler{
		Transaction: NewTransactionHandleGrpc(
			deps.Service,
			deps.Logger,
		),
	}
}
