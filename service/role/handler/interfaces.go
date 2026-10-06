package handler

import (
	"context"

	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	"google.golang.org/protobuf/types/known/emptypb"
)

type RoleHandleGrpc interface {
	pbrole.RoleServiceServer

	FindAllRole(ctx context.Context, req *pbrole.FindAllRoleRequest) (*pbrole.ApiResponsePaginationRole, error)
	FindByIdRole(ctx context.Context, req *pbrole.FindByIdRoleRequest) (*pbrole.ApiResponseRole, error)
	FindByNameRole(ctx context.Context, req *pbrole.FindByNameRoleRequest) (*pbrole.ApiResponseRole, error)
	FindByActive(ctx context.Context, req *pbrole.FindAllRoleRequest) (*pbrole.ApiResponsePaginationRoleDeleteAt, error)
	FindByTrashed(ctx context.Context, req *pbrole.FindAllRoleRequest) (*pbrole.ApiResponsePaginationRoleDeleteAt, error)
	CreateRole(ctx context.Context, req *pbrole.CreateRoleRequest) (*pbrole.ApiResponseRole, error)
	UpdateRole(ctx context.Context, req *pbrole.UpdateRoleRequest) (*pbrole.ApiResponseRole, error)
	TrashedRole(ctx context.Context, req *pbrole.FindByIdRoleRequest) (*pbrole.ApiResponseRole, error)
	RestoreRole(ctx context.Context, req *pbrole.FindByIdRoleRequest) (*pbrole.ApiResponseRole, error)
	DeleteRolePermanent(ctx context.Context, req *pbrole.FindByIdRoleRequest) (*pbrole.ApiResponseRoleDelete, error)
	RestoreAllRole(ctx context.Context, req *emptypb.Empty) (*pbrole.ApiResponseRoleAll, error)
	DeleteAllRolePermanent(ctx context.Context, req *emptypb.Empty) (*pbrole.ApiResponseRoleAll, error)
}
