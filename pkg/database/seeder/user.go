package seeder

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const seedCreateUserSQL = `
INSERT INTO users (firstname, lastname, email, password, verification_code, is_verified, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, current_timestamp, current_timestamp)
RETURNING user_id, email`

const seedTrashUserSQL = `
UPDATE users SET deleted_at = current_timestamp
WHERE user_id = $1 AND deleted_at IS NULL
RETURNING user_id`

type userSeeder struct {
	db     *sqlx.DB
	hash   hash.HashPassword
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewUserSeeder(db *sqlx.DB, hash hash.HashPassword, ctx context.Context, logger logger.LoggerInterface) *userSeeder {
	return &userSeeder{
		db:     db,
		hash:   hash,
		ctx:    ctx,
		logger: logger,
	}
}

func (r *userSeeder) Seed() error {
	for i := 1; i <= 10; i++ {
		email := fmt.Sprintf("user_%s@example.com", uuid.NewString())
		rawPassword := fmt.Sprintf("password%d", i)

		hashedPassword, err := r.hash.HashPassword(rawPassword)
		if err != nil {
			r.logger.Error("failed to hash password", zap.Int("user", i), zap.Error(err))
			return fmt.Errorf("failed to hash password for user %d: %w", i, err)
		}

		var created seedUserRow
		if err := r.db.GetContext(r.ctx, &created, seedCreateUserSQL,
			fmt.Sprintf("User%d", i), fmt.Sprintf("Last%d", i), email, hashedPassword, "", nil); err != nil {
			r.logger.Error("failed to seed user", zap.Int("user", i), zap.Error(err))
			return fmt.Errorf("failed to seed user %d: %w", i, err)
		}

		if i > 5 {
			var trashed seedUserRow
			if err := r.db.GetContext(r.ctx, &trashed, seedTrashUserSQL, created.UserID); err != nil {
				r.logger.Error("failed to trash user", zap.Int("user", i), zap.Error(err))
				return fmt.Errorf("failed to trash user %d: %w", i, err)
			}
		}
	}

	r.logger.Info("User seeding completed successfully", zap.Int("totalUsers", 10))
	return nil
}
