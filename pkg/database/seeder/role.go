package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
	"golang.org/x/exp/rand"
)

const seedCreateRoleSQL = `
INSERT INTO roles (role_name, created_at, updated_at)
VALUES ($1, current_timestamp, current_timestamp)
RETURNING role_id, role_name`

const seedGetRolesSQL = `
SELECT role_id, role_name
FROM roles
WHERE $1::TEXT IS NULL OR role_name ILIKE '%' || $1 || '%'
ORDER BY created_at ASC
LIMIT $2 OFFSET $3`

const seedAssignRoleSQL = `
INSERT INTO user_roles (user_id, role_id, created_at, updated_at)
VALUES ($1, $2, current_timestamp, current_timestamp)`

const seedGetUsersForRolesSQL = `
SELECT user_id, email
FROM users
WHERE deleted_at IS NULL
  AND ($1::TEXT IS NULL OR firstname ILIKE '%' || $1 || '%' OR lastname ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

type seedRoleRow struct {
	RoleID   int32  `db:"role_id"`
	RoleName string `db:"role_name"`
}

type seedUserRow struct {
	UserID int32  `db:"user_id"`
	Email  string `db:"email"`
}

type roleSeeder struct {
	db     *sqlx.DB
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewRoleSeeder(db *sqlx.DB, ctx context.Context, logger logger.LoggerInterface) *roleSeeder {
	return &roleSeeder{db: db, ctx: ctx, logger: logger}
}

func (r *roleSeeder) Seed() error {
	randomRoles := []string{
		"ROLE_ADMIN",
		"Admin Access 1",
		"Super Admin",
		"Admin",
		"Store Manager",
		"Cashier",
		"Inventory Staff",
		"Support",
		"Auditor",
		"Viewer",
	}

	totalRoles := len(randomRoles)

	for i, roleName := range randomRoles {
		var created seedRoleRow
		if err := r.db.GetContext(r.ctx, &created, seedCreateRoleSQL, roleName); err != nil {
			r.logger.Error("failed to seed role", zap.Int("role", i+1), zap.String("roleName", roleName), zap.Error(err))
			return fmt.Errorf("failed to seed role %d (%s): %w", i+1, roleName, err)
		}
	}

	r.logger.Info("role seeded successfully", zap.Int("totalRoles", totalRoles))
	return nil
}

type userRoleSeeder struct {
	roleDB *sqlx.DB
	userDB *sqlx.DB
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewUserRoleSeeder(roleDB *sqlx.DB, userDB *sqlx.DB, ctx context.Context, logger logger.LoggerInterface) *userRoleSeeder {
	return &userRoleSeeder{roleDB: roleDB, userDB: userDB, ctx: ctx, logger: logger}
}

func (r *userRoleSeeder) Seed() error {
	var users []seedUserRow
	if err := r.userDB.SelectContext(r.ctx, &users, seedGetUsersForRolesSQL, "", int32(20), 0); err != nil {
		r.logger.Error("failed to fetch users", zap.Error(err))
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	var roles []seedRoleRow
	if err := r.roleDB.SelectContext(r.ctx, &roles, seedGetRolesSQL, "", 4, 0); err != nil {
		r.logger.Error("failed to fetch roles", zap.Error(err))
		return fmt.Errorf("failed to fetch roles: %w", err)
	}

	if len(users) == 0 || len(roles) == 0 {
		r.logger.Error("No users or roles found. User-role seeding aborted.")
		return fmt.Errorf("no users or roles found for user-role seeding")
	}

	rand.Seed(uint64(time.Now().UnixNano()))

	for _, user := range users {
		role := roles[rand.Intn(len(roles))]

		if _, err := r.roleDB.ExecContext(r.ctx, seedAssignRoleSQL, user.UserID, role.RoleID); err != nil {
			r.logger.Error("failed to assign role to user", zap.String("user", user.Email), zap.String("role", role.RoleName), zap.Error(err))
			return fmt.Errorf("failed to assign role %s to user %s: %w", role.RoleName, user.Email, err)
		}
	}

	r.logger.Info("user roles assigned successfully", zap.Int("users", len(users)), zap.Int("roles", len(roles)))
	return nil
}
