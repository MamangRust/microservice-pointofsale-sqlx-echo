package handler

import (
	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
)

type TransactionHandleGrpc interface {
	pbtransaction.TransactionServiceServer
}
