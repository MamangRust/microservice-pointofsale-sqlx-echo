package handler

import (
	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-role/service"
)

type Deps struct {
	Service *service.Service
	Logger  logger.LoggerInterface
}

type Handler struct {
	Role     pbrole.RoleServiceServer
	UserRole pbuserrole.UserRoleServiceServer
}

func NewHandler(deps *Deps) *Handler {
	return &Handler{
		Role: NewRoleHandleGrpc(
			deps.Service,
			deps.Logger,
		),
		UserRole: NewUserRoleHandleGrpc(
			deps.Service,
			deps.Logger,
		),
	}
}
