package repository

import (
	"context"
	"time"

	db "github.com/MamangRust/microservice-point-of-sale-auth/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"github.com/jmoiron/sqlx"
)

// SQL source of truth (formerly database/query/reset_token.sql).
const (
	createResetToken = `INSERT INTO reset_tokens (user_id, token, expiry_date)
VALUES ($1, $2, $3)
RETURNING *;`

	deleteResetToken = `DELETE FROM reset_tokens
WHERE user_id = $1;`

	getResetToken = `SELECT * FROM reset_tokens
WHERE token = $1;`
)

type resetTokenRepository struct {
	db *sqlx.DB
}

func NewResetTokenRepository(db *sqlx.DB) *resetTokenRepository {
	return &resetTokenRepository{
		db: db,
	}
}

func (r *resetTokenRepository) FindByToken(ctx context.Context, code string) (*db.ResetToken, error) {
	var resetToken db.ResetToken
	err := r.db.GetContext(ctx, &resetToken, getResetToken, code)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &resetToken, nil
}

func (r *resetTokenRepository) CreateResetToken(ctx context.Context, req *requests.CreateResetTokenRequest) (*db.ResetToken, error) {
	expiryDate, err := time.Parse("2006-01-02 15:04:05", req.ExpiredAt)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	var resetToken db.ResetToken
	err = r.db.GetContext(ctx, &resetToken, createResetToken,
		int64(req.UserID),
		req.ResetToken,
		expiryDate,
	)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &resetToken, nil
}

// CreateResetTokenInTx persists the reset token inside the given database
// transaction so the caller can commit the token write and its outbox event
// atomically (Phase 6).
func (r *resetTokenRepository) CreateResetTokenInTx(ctx context.Context, tx *sqlx.Tx, req *requests.CreateResetTokenRequest) (*db.ResetToken, error) {
	expiryDate, err := time.Parse("2006-01-02 15:04:05", req.ExpiredAt)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	var resetToken db.ResetToken
	err = tx.GetContext(ctx, &resetToken, createResetToken,
		int64(req.UserID),
		req.ResetToken,
		expiryDate,
	)
	if err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &resetToken, nil
}

func (r *resetTokenRepository) DeleteResetToken(ctx context.Context, user_id int) error {
	_, err := r.db.ExecContext(ctx, deleteResetToken, int64(user_id))
	if err != nil {
		return sharedErrors.ErrInternal.WithInternal(err)
	}
	return nil
}
