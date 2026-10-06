// Package db holds the database row models for the auth service.
// Hand-written (migrated from sqlc generation); used with jmoiron/sqlx.
package db

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

type RefreshToken struct {
	RefreshTokenID int32            `json:"refresh_token_id" db:"refresh_token_id"`
	UserID         int32            `json:"user_id" db:"user_id"`
	Token          string           `json:"token" db:"token"`
	Expiration     time.Time        `json:"expiration" db:"expiration"`
	CreatedAt      pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt      pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt      pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}

type ResetToken struct {
	ID         int32     `json:"id" db:"id"`
	UserID     int64     `json:"user_id" db:"user_id"`
	Token      string    `json:"token" db:"token"`
	ExpiryDate time.Time `json:"expiry_date" db:"expiry_date"`
}

type Role struct {
	RoleID    int32            `json:"role_id" db:"role_id"`
	RoleName  string           `json:"role_name" db:"role_name"`
	CreatedAt pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}

type User struct {
	UserID           int32            `json:"user_id" db:"user_id"`
	Firstname        string           `json:"firstname" db:"firstname"`
	Lastname         string           `json:"lastname" db:"lastname"`
	Email            string           `json:"email" db:"email"`
	Password         string           `json:"password" db:"password"`
	VerificationCode string           `json:"verification_code" db:"verification_code"`
	IsVerified       *bool            `json:"is_verified" db:"is_verified"`
	CreatedAt        pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt        pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt        pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}

type UserRole struct {
	UserRoleID int32            `json:"user_role_id" db:"user_role_id"`
	UserID     int32            `json:"user_id" db:"user_id"`
	RoleID     int32            `json:"role_id" db:"role_id"`
	CreatedAt  pgtype.Timestamp `json:"created_at" db:"created_at"`
	UpdatedAt  pgtype.Timestamp `json:"updated_at" db:"updated_at"`
	DeletedAt  pgtype.Timestamp `json:"deleted_at" db:"deleted_at"`
}
