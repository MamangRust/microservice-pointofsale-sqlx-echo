package handler

import (
	"github.com/MamangRust/microservice-point-of-sale-order-item/service"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
)

type Deps struct {
	Service *service.Service
	Logger  logger.LoggerInterface
}

type Handler struct {
	OrderItem pborderitem.OrderItemServiceServer
}

func NewHandler(deps *Deps) *Handler {
	return &Handler{
		OrderItem: NewOrderItemHandleGrpc(
			deps.Service.OrderItemQuery,
			deps.Service.OrderItemCommand,
			deps.Logger,
		),
	}
}
