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
	UserCommand UserCommandRepository
	UserQuery   UserQueryRepository
	Role        roleadapter.QueryRepository
	UserRole    userroleadapter.CommandRepository
}

// GuardOptions groups dependency guards by downstream gRPC dependency.
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

	roleRepo := roleadapter.New(roleClient, g.Role...)
	userRoleRepo := userroleadapter.New(userRoleClient, g.UserRole...)

	return &Repositories{
		UserCommand: NewUserCommandRepository(DB),
		UserQuery:   NewUserQueryRepository(DB),
		Role:        roleRepo,
		UserRole:    userRoleRepo,
	}
}
