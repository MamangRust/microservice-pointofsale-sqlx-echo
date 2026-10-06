package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/user_errors"
	db "github.com/MamangRust/microservice-point-of-sale-user/database/schema"
	"github.com/jmoiron/sqlx"
)

const getUsers = `SELECT
    *,
    COUNT(*) OVER() AS total_count
FROM users
WHERE deleted_at IS NULL
  AND ($1::TEXT IS NULL OR firstname ILIKE '%' || $1 || '%' OR lastname ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

const getUsersActive = `SELECT
    *,
    COUNT(*) OVER() AS total_count
FROM users
WHERE deleted_at IS NULL
  AND ($1::TEXT IS NULL OR firstname ILIKE '%' || $1 || '%' OR lastname ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

const getUserTrashed = `SELECT
    *,
    COUNT(*) OVER() AS total_count
FROM users
WHERE deleted_at IS NOT NULL
  AND ($1::TEXT IS NULL OR firstname ILIKE '%' || $1 || '%' OR lastname ILIKE '%' || $1 || '%' OR email ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

const getUserByID = `SELECT * FROM users WHERE user_id = $1 AND deleted_at IS NULL`

const getUserByEmail = `SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL`

type userQueryRepository struct {
	db *sqlx.DB
}

func NewUserQueryRepository(db *sqlx.DB) *userQueryRepository {
	return &userQueryRepository{
		db: db,
	}
}

func (r *userQueryRepository) FindAllUsers(ctx context.Context, req *requests.FindAllUsers) ([]*db.GetUsersRow, *int, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetUsersRow
	err := r.db.SelectContext(ctx, &res, getUsers,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, nil, user_errors.ErrFindAllUsers
	}

	var totalCount int
	if len(res) > 0 {
		totalCount = int(res[0].TotalCount)
	} else {
		totalCount = 0
	}

	return res, &totalCount, nil
}

func (r *userQueryRepository) FindById(ctx context.Context, user_id int) (*db.User, error) {
	var user db.User
	err := r.db.GetContext(ctx, &user, getUserByID, int32(user_id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, user_errors.ErrUserNotFound
		}
		return nil, user_errors.ErrInternalServerError.WithInternal(err)
	}

	return &user, nil
}

func (r *userQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllUsers) ([]*db.GetUsersActiveRow, *int, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetUsersActiveRow
	err := r.db.SelectContext(ctx, &res, getUsersActive,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, nil, user_errors.ErrFindActiveUsers
	}

	var totalCount int
	if len(res) > 0 {
		totalCount = int(res[0].TotalCount)
	} else {
		totalCount = 0
	}

	return res, &totalCount, nil
}

func (r *userQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllUsers) ([]*db.GetUserTrashedRow, *int, error) {
	offset := (req.Page - 1) * req.PageSize

	var res []*db.GetUserTrashedRow
	err := r.db.SelectContext(ctx, &res, getUserTrashed,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, nil, user_errors.ErrFindTrashedUsers
	}

	var totalCount int
	if len(res) > 0 {
		totalCount = int(res[0].TotalCount)
	} else {
		totalCount = 0
	}

	return res, &totalCount, nil
}

func (r *userQueryRepository) FindByEmail(ctx context.Context, email string) (*db.User, error) {
	var user db.User
	err := r.db.GetContext(ctx, &user, getUserByEmail, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, user_errors.ErrInternalServerError.WithInternal(err)
	}

	return &user, nil
}
