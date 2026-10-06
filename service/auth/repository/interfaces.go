package repository

import (
	"context"

	db "github.com/MamangRust/microservice-point-of-sale-auth/database/schema"
	roleadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/role"
	userroleadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/user_role"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/jmoiron/sqlx"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*db.User, error)
	FindById(ctx context.Context, id int) (*db.User, error)
	FindByEmailAndVerify(ctx context.Context, email string) (*db.User, error)
	CreateUser(ctx context.Context, request *requests.RegisterRequest) (*db.User, error)
	CreateUserInTx(ctx context.Context, tx *sqlx.Tx, request *requests.RegisterRequest) (*db.User, error)
	UpdateUserIsVerified(ctx context.Context, user_id int, is_verified bool) (*db.User, error)
	UpdateUserPassword(ctx context.Context, user_id int, password string) (*db.User, error)
	FindByVerificationCode(ctx context.Context, verification_code string) (*db.User, error)
}

type ResetTokenRepository interface {
	FindByToken(ctx context.Context, token string) (*db.ResetToken, error)
	CreateResetToken(ctx context.Context, req *requests.CreateResetTokenRequest) (*db.ResetToken, error)
	CreateResetTokenInTx(ctx context.Context, tx *sqlx.Tx, req *requests.CreateResetTokenRequest) (*db.ResetToken, error)
	DeleteResetToken(ctx context.Context, user_id int) error
}

type RefreshTokenRepository interface {
	FindByToken(ctx context.Context, token string) (*db.RefreshToken, error)
	FindByUserId(ctx context.Context, user_id int) (*db.RefreshToken, error)
	CreateRefreshToken(ctx context.Context, req *requests.CreateRefreshToken) (*db.RefreshToken, error)
	UpdateRefreshToken(ctx context.Context, req *requests.UpdateRefreshToken) (*db.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) error
	DeleteRefreshTokenByUserId(ctx context.Context, user_id int) error
}

// UserRoleRepository and RoleRepository are satisfied by the shared gRPC
// adapters: role assignment lives in the user_role service, role lookups in the
// role service. Auth no longer reads the roles/user_roles tables directly.
type UserRoleRepository = userroleadapter.CommandRepository

type RoleRepository = roleadapter.QueryRepository
