package handler

import (
	"context"

	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserHandleGrpc interface {
	pbuser.UserServiceServer

	FindAll(ctx context.Context, request *pbuser.FindAllUserRequest) (*pbuser.ApiResponsePaginationUser, error)
	FindById(ctx context.Context, request *pbuser.FindByIdUserRequest) (*pbuser.ApiResponseUser, error)
	FindByActive(ctx context.Context, request *pbuser.FindAllUserRequest) (*pbuser.ApiResponsePaginationUserDeleteAt, error)
	FindByTrashed(ctx context.Context, request *pbuser.FindAllUserRequest) (*pbuser.ApiResponsePaginationUserDeleteAt, error)
	Create(ctx context.Context, request *pbuser.CreateUserRequest) (*pbuser.ApiResponseUser, error)
	Update(ctx context.Context, request *pbuser.UpdateUserRequest) (*pbuser.ApiResponseUser, error)
	TrashedUser(ctx context.Context, request *pbuser.FindByIdUserRequest) (*pbuser.ApiResponseUserDeleteAt, error)
	RestoreUser(ctx context.Context, request *pbuser.FindByIdUserRequest) (*pbuser.ApiResponseUserDeleteAt, error)
	DeleteUserPermanent(ctx context.Context, request *pbuser.FindByIdUserRequest) (*pbuser.ApiResponseUserDelete, error)

	RestoreAllUser(context.Context, *emptypb.Empty) (*pbuser.ApiResponseUserAll, error)
	DeleteAllUserPermanent(context.Context, *emptypb.Empty) (*pbuser.ApiResponseUserAll, error)
}
