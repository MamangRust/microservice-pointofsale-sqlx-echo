package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

type GetRolesRow struct {
	RoleID     int32            `db:"role_id" json:"role_id"`
	RoleName   string           `db:"role_name" json:"role_name"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetActiveRolesRow struct {
	RoleID     int32            `db:"role_id" json:"role_id"`
	RoleName   string           `db:"role_name" json:"role_name"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetTrashedRolesRow struct {
	RoleID     int32            `db:"role_id" json:"role_id"`
	RoleName   string           `db:"role_name" json:"role_name"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount int64            `db:"total_count" json:"total_count"`
}

type GetTrashedUserRolesRow struct {
	UserRoleID int32            `db:"user_role_id" json:"user_role_id"`
	UserID     int32            `db:"user_id" json:"user_id"`
	RoleID     int32            `db:"role_id" json:"role_id"`
	RoleName   string           `db:"role_name" json:"role_name"`
	CreatedAt  pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt  pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt  pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
}
