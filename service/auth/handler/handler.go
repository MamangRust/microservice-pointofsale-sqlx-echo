package handler

import (
	"github.com/MamangRust/microservice-point-of-sale-auth/service"
	pbauth "github.com/MamangRust/microservice-point-of-sale-pb/auth"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
)

type Deps struct {
	Service *service.Service
	Logger  logger.LoggerInterface
}

type Handler struct {
	Auth pbauth.AuthServiceServer
}

func NewHandler(deps *Deps) *Handler {
	return &Handler{
		Auth: NewAuthHandleGrpc(
			deps.Service,
			deps.Logger,
		),
	}
}
