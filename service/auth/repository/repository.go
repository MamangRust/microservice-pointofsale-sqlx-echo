package repository

import (
	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	roleadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/role"
	userroleadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/user_role"
	"github.com/jmoiron/sqlx"
)

type Repositories struct {
	User         UserRepository
	RefreshToken RefreshTokenRepository
	UserRole     UserRoleRepository
	Role         RoleRepository
	ResetToken   ResetTokenRepository
}

type GuardOptions struct {
	Role     []adapter.GuardOption
	UserRole []adapter.GuardOption
}

func NewRepositories(
	DB *sqlx.DB,
	roleClient pbrole.RoleServiceClient,
	userRoleClient pbuserrole.UserRoleServiceClient,
	guards ...GuardOptions,
) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &Repositories{
		User:         NewUserRepository(DB),
		RefreshToken: NewRefreshTokenRepository(DB),
		UserRole:     userroleadapter.New(userRoleClient, g.UserRole...),
		Role:         roleadapter.New(roleClient, g.Role...),
		ResetToken:   NewResetTokenRepository(DB),
	}
}
