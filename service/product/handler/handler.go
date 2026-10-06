package handler

import (
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-product/service"
)

type Deps struct {
	Service *service.Service
	Logger  logger.LoggerInterface
}

type Handler struct {
	Product pbproduct.ProductServiceServer
}

func NewHandler(deps *Deps) *Handler {
	return &Handler{
		Product: NewProductHandleGrpc(
			deps.Service,
			deps.Logger,
		),
	}
}
