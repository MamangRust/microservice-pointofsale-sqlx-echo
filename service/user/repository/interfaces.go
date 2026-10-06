package repository

import (
	"context"

	db "github.com/MamangRust/microservice-point-of-sale-user/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type UserQueryRepository interface {
	FindAllUsers(ctx context.Context, req *requests.FindAllUsers) ([]*db.GetUsersRow, *int, error)
	FindById(ctx context.Context, user_id int) (*db.User, error)
	FindByEmail(ctx context.Context, email string) (*db.User, error)
	FindByActive(ctx context.Context, req *requests.FindAllUsers) ([]*db.GetUsersActiveRow, *int, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllUsers) ([]*db.GetUserTrashedRow, *int, error)
}

type UserCommandRepository interface {
	CreateUser(ctx context.Context, request *requests.CreateUserRequest) (*db.User, error)
	UpdateUser(ctx context.Context, request *requests.UpdateUserRequest) (*db.User, error)
	TrashedUser(ctx context.Context, user_id int) (*db.User, error)
	RestoreUser(ctx context.Context, user_id int) (*db.User, error)
	DeleteUserPermanent(ctx context.Context, user_id int) (bool, error)
	RestoreAllUser(ctx context.Context) (bool, error)
	DeleteAllUserPermanent(ctx context.Context) (bool, error)
}
