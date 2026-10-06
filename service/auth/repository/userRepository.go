package repository

import (
	"context"
	"database/sql"
	"errors"

	db "github.com/MamangRust/microservice-point-of-sale-auth/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/user_errors"
	"github.com/jmoiron/sqlx"
)

// SQL source of truth (formerly database/query/users.sql).
const (
	getUsers = `SELECT
    *,
    COUNT(*) OVER() AS total_count
FROM users
WHERE deleted_at IS NULL
  AND ($1::TEXT IS NULL OR firstname ILIKE '%' || $1 || '%' OR lastname ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;`

	getUsersActive = `SELECT
    *,
    COUNT(*) OVER() AS total_count
FROM users
WHERE deleted_at IS NULL
  AND ($1::TEXT IS NULL OR firstname ILIKE '%' || $1 || '%' OR lastname ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;`

	getUserTrashed = `SELECT
    *,
    COUNT(*) OVER() AS total_count
FROM users
WHERE deleted_at IS NOT NULL
  AND ($1::TEXT IS NULL OR firstname ILIKE '%' || $1 || '%' OR lastname ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;`

	getUserByID = `SELECT * FROM users WHERE user_id = $1 AND deleted_at IS NULL;`

	getUserByVerificationCode = `SELECT * FROM users WHERE verification_code = $1;`

	getUserByEmail = `SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL;`

	getUserByEmailAndVerified = `SELECT *
FROM users
WHERE email = $1
  AND is_verified = true
  AND deleted_at IS NULL;`

	createUser = `INSERT INTO
    users (
        firstname,
        lastname,
        email,
        password,
        verification_code,
        is_verified,
        created_at,
        updated_at
    )
VALUES (
        $1,
        $2,
        $3,
        $4,
        $5,
        $6,
        current_timestamp,
        current_timestamp
    ) RETURNING *;`

	updateUser = `UPDATE users
SET
    firstname = $2,
    lastname = $3,
    email = $4,
    password = $5,
    updated_at = current_timestamp
WHERE
    user_id = $1
    AND deleted_at IS NULL
    RETURNING *;`

	updateUserPassword = `UPDATE users
SET
    password = $2,
    updated_at = current_timestamp
WHERE
    user_id = $1
    AND deleted_at IS NULL
    RETURNING *;`

	updateUserIsVerified = `UPDATE users
SET
    is_verified = $2,
    updated_at = current_timestamp
WHERE
    user_id = $1
    AND deleted_at IS NULL
    RETURNING *;`

	trashUser = `UPDATE users
SET
    deleted_at = current_timestamp
WHERE
    user_id = $1
    AND deleted_at IS NULL
    RETURNING *;`

	restoreUser = `UPDATE users
SET
    deleted_at = NULL
WHERE
    user_id = $1
    AND deleted_at IS NOT NULL
    RETURNING *;`

	deleteUserPermanently = `DELETE FROM users WHERE user_id = $1 AND deleted_at IS NOT NULL;`

	restoreAllUsers = `UPDATE users
SET
    deleted_at = NULL
WHERE
    deleted_at IS NOT NULL;`

	deleteAllPermanentUsers = `DELETE FROM users
WHERE
    deleted_at IS NOT NULL;`
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *userRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) FindById(ctx context.Context, user_id int) (*db.User, error) {
	var user db.User
	err := r.db.GetContext(ctx, &user, getUserByID, int32(user_id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user_errors.ErrUserNotFound
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*db.User, error) {
	var user db.User
	err := r.db.GetContext(ctx, &user, getUserByEmail, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &user, nil
}

func (r *userRepository) FindByEmailAndVerify(ctx context.Context, email string) (*db.User, error) {
	var user db.User
	err := r.db.GetContext(ctx, &user, getUserByEmailAndVerified, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user_errors.ErrUserNotFound
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &user, nil
}

func (r *userRepository) FindByVerificationCode(ctx context.Context, verification_code string) (*db.User, error) {
	var user db.User
	err := r.db.GetContext(ctx, &user, getUserByVerificationCode, verification_code)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user_errors.ErrUserNotFound
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &user, nil
}

func (r *userRepository) CreateUser(ctx context.Context, request *requests.RegisterRequest) (*db.User, error) {
	var user db.User
	err := r.db.GetContext(ctx, &user, createUser,
		request.FirstName,
		request.LastName,
		request.Email,
		request.Password,
		request.VerifiedCode,
		&request.IsVerified,
	)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &user, nil
}

// CreateUserInTx persists the user inside the given database transaction so the
// caller can commit the user write and its outbox event atomically (Phase 6).
func (r *userRepository) CreateUserInTx(ctx context.Context, tx *sqlx.Tx, request *requests.RegisterRequest) (*db.User, error) {
	var user db.User
	err := tx.GetContext(ctx, &user, createUser,
		request.FirstName,
		request.LastName,
		request.Email,
		request.Password,
		request.VerifiedCode,
		&request.IsVerified,
	)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &user, nil
}

func (r *userRepository) UpdateUserIsVerified(ctx context.Context, user_id int, is_verified bool) (*db.User, error) {
	var user db.User
	err := r.db.GetContext(ctx, &user, updateUserIsVerified, int32(user_id), &is_verified)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &user, nil
}

func (r *userRepository) UpdateUserPassword(ctx context.Context, user_id int, password string) (*db.User, error) {
	var user db.User
	err := r.db.GetContext(ctx, &user, updateUserPassword, int32(user_id), password)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &user, nil
}
