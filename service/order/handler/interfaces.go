package handler

import (
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
)

type OrderHandleGrpc interface {
	pborder.OrderServiceServer
}
