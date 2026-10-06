package handler

import (
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-user/service"
)

type Deps struct {
	Service *service.Service
	Logger  logger.LoggerInterface
}

type Handler struct {
	User pbuser.UserServiceServer
}

func NewHandler(deps *Deps) *Handler {
	return &Handler{
		User: NewUserHandleGrpc(
			deps.Service,
			deps.Logger,
		),
	}
}
