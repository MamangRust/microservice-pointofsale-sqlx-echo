package handler

import (
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
)

type CashierHandleGrpc interface {
	pbcashier.CashierServiceServer
}
