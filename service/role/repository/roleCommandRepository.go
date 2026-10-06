package repository

import (
	"context"

	db "github.com/MamangRust/microservice-point-of-sale-role/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/jmoiron/sqlx"
)

const createRole = `INSERT INTO roles (
    role_name,
    created_at,
    updated_at
) VALUES (
    $1,
    current_timestamp,
    current_timestamp
) RETURNING
    role_id,
    role_name,
    created_at,
    updated_at,
    deleted_at`

const updateRole = `UPDATE roles
SET
    role_name = $2,
    updated_at = current_timestamp
WHERE
    role_id = $1
RETURNING
    role_id,
    role_name,
    created_at,
    updated_at,
    deleted_at`

const trashRole = `UPDATE roles
SET
    deleted_at = current_timestamp
WHERE
    role_id = $1
RETURNING *`

const restoreRole = `UPDATE roles
SET
    deleted_at = NULL
WHERE
    role_id = $1
RETURNING *`

const deletePermanentRole = `DELETE FROM roles
WHERE
    role_id = $1 AND deleted_at IS NOT NULL`

const restoreAllRoles = `UPDATE roles
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL`

const deleteAllPermanentRoles = `DELETE FROM roles
WHERE
    deleted_at IS NOT NULL`

type roleCommandRepository struct {
	db *sqlx.DB
}

func NewRoleCommandRepository(db *sqlx.DB) *roleCommandRepository {
	return &roleCommandRepository{
		db: db,
	}
}

func (r *roleCommandRepository) CreateRole(ctx context.Context, req *requests.CreateRoleRequest) (*db.Role, error) {
	var role db.Role
	err := r.db.GetContext(ctx, &role, createRole, req.Name)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *roleCommandRepository) UpdateRole(ctx context.Context, req *requests.UpdateRoleRequest) (*db.Role, error) {
	var role db.Role
	err := r.db.GetContext(ctx, &role, updateRole,
		int32(*req.ID),
		req.Name,
	)
	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *roleCommandRepository) TrashedRole(ctx context.Context, id int) (*db.Role, error) {
	var role db.Role
	err := r.db.GetContext(ctx, &role, trashRole, int32(id))
	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *roleCommandRepository) RestoreRole(ctx context.Context, id int) (*db.Role, error) {
	var role db.Role
	err := r.db.GetContext(ctx, &role, restoreRole, int32(id))
	if err != nil {
		return nil, err
	}

	return &role, nil
}

func (r *roleCommandRepository) DeleteRolePermanent(ctx context.Context, roleID int) (bool, error) {
	_, err := r.db.ExecContext(ctx, deletePermanentRole, int32(roleID))
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *roleCommandRepository) RestoreAllRole(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, restoreAllRoles)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (r *roleCommandRepository) DeleteAllRolePermanent(ctx context.Context) (bool, error) {
	_, err := r.db.ExecContext(ctx, deleteAllPermanentRoles)
	if err != nil {
		return false, err
	}

	return true, nil
}
