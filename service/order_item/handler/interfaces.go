package handler

import (
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
)

type OrderItemHandlerGrpc interface {
	pborderitem.OrderItemServiceServer
}
