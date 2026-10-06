// Package db — parameter structs kept from the former sqlc generation.
// Only the types still referenced by repositories/services live here; sqlx
// uses positional $n placeholders, so these are plain data holders.
package db

import (
	"time"
)

type CreateUserParams struct {
	Firstname        string `json:"firstname" db:"firstname"`
	Lastname         string `json:"lastname" db:"lastname"`
	Email            string `json:"email" db:"email"`
	Password         string `json:"password" db:"password"`
	VerificationCode string `json:"verification_code" db:"verification_code"`
	IsVerified       *bool  `json:"is_verified" db:"is_verified"`
}

type UpdateUserIsVerifiedParams struct {
	UserID     int32 `json:"user_id" db:"user_id"`
	IsVerified *bool `json:"is_verified" db:"is_verified"`
}

type UpdateUserPasswordParams struct {
	UserID   int32  `json:"user_id" db:"user_id"`
	Password string `json:"password" db:"password"`
}

type CreateRefreshTokenParams struct {
	UserID     int32     `json:"user_id" db:"user_id"`
	Token      string    `json:"token" db:"token"`
	Expiration time.Time `json:"expiration" db:"expiration"`
}

type UpdateRefreshTokenByUserIdParams struct {
	UserID     int32     `json:"user_id" db:"user_id"`
	Token      string    `json:"token" db:"token"`
	Expiration time.Time `json:"expiration" db:"expiration"`
}

type CreateResetTokenParams struct {
	UserID     int64     `json:"user_id" db:"user_id"`
	Token      string    `json:"token" db:"token"`
	ExpiryDate time.Time `json:"expiry_date" db:"expiry_date"`
}

type AssignRoleToUserParams struct {
	UserID int32 `json:"user_id" db:"user_id"`
	RoleID int32 `json:"role_id" db:"role_id"`
}

type RemoveRoleFromUserParams struct {
	UserID int32 `json:"user_id" db:"user_id"`
	RoleID int32 `json:"role_id" db:"role_id"`
}
