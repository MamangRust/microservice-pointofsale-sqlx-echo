package handler

import (
	"github.com/MamangRust/microservice-point-of-sale-order/service"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
)

type Deps struct {
	Service *service.Service
	Logger  logger.LoggerInterface
}

type Handler struct {
	Order pborder.OrderServiceServer
}

func NewHandler(deps *Deps) *Handler {
	return &Handler{
		Order: NewOrderHandleGrpc(
			deps.Service,
			deps.Logger,
		),
	}
}
