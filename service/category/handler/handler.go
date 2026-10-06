package handler

import (
	"github.com/MamangRust/microservice-point-of-sale-category/service"
	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
)

type Deps struct {
	Service *service.Service
	Logger  logger.LoggerInterface
}

type Handler struct {
	Category pbcategory.CategoryServiceServer
}

func NewHandler(deps *Deps) *Handler {
	return &Handler{
		Category: NewCategoryHandleGrpc(
			deps.Service,
			deps.Logger,
		),
	}
}
