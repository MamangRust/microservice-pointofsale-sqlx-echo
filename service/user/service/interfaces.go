package service

import (
	"context"

	db "github.com/MamangRust/microservice-point-of-sale-user/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type UserQueryService interface {
	FindAll(ctx context.Context, req *requests.FindAllUsers) ([]*db.GetUsersRow, *int, error)
	FindByID(ctx context.Context, id int) (*db.User, error)
	FindByActive(ctx context.Context, req *requests.FindAllUsers) ([]*db.GetUsersActiveRow, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllUsers) ([]*db.GetUserTrashedRow, *int, error)
}

type UserCommandService interface {
	CreateUser(ctx context.Context, request *requests.CreateUserRequest) (*db.User, error)
	UpdateUser(ctx context.Context, request *requests.UpdateUserRequest) (*db.User, error)
	TrashedUser(ctx context.Context, user_id int) (*db.User, error)
	RestoreUser(ctx context.Context, user_id int) (*db.User, error)
	DeleteUserPermanent(ctx context.Context, user_id int) (bool, error)

	RestoreAllUser(ctx context.Context) (bool, error)
	DeleteAllUserPermanent(ctx context.Context) (bool, error)
}
