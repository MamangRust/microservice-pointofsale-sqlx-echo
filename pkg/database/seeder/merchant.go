package seeder

import (
	"context"
	"fmt"

	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const seedGetMerchantsSQL = `
SELECT merchant_id
FROM merchants
WHERE deleted_at IS NULL
AND ($1::TEXT IS NULL OR name ILIKE '%' || $1 || '%' OR contact_email ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

const seedCreateMerchantSQL = `
INSERT INTO merchants (user_id, name, description, address, contact_email, contact_phone, status)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING merchant_id`

type seedMerchantRow struct {
	MerchantID int32 `db:"merchant_id"`
}

type merchantSeeder struct {
	db     *sqlx.DB
	userDB *sqlx.DB
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewMerchantSeeder(db *sqlx.DB, userDB *sqlx.DB, ctx context.Context, logger logger.LoggerInterface) *merchantSeeder {
	return &merchantSeeder{db: db, userDB: userDB, ctx: ctx, logger: logger}
}

func (r *merchantSeeder) Seed() error {
	var users []seedUserRow
	if err := r.userDB.SelectContext(r.ctx, &users, seedGetUsersForRolesSQL, "", int32(20), 0); err != nil {
		r.logger.Error("Failed to fetch users:", zap.Any("error", err))
		return err
	}

	if len(users) == 0 {
		r.logger.Error("No users found. Merchant seeding aborted.")
		return fmt.Errorf("no users found for merchant seeding")
	}

	for i := 1; i <= 10; i++ {
		userID := users[i%len(users)].UserID

		var created seedMerchantRow
		if err := r.db.GetContext(r.ctx, &created, seedCreateMerchantSQL,
			userID,
			fmt.Sprintf("Toko %d", i),
			ptrString(fmt.Sprintf("Deskripsi untuk Toko %d", i)),
			ptrString(fmt.Sprintf("Jl. Toko %d", i)),
			ptrString(fmt.Sprintf("toko%d@example.com", i)),
			ptrString(fmt.Sprintf("0812345678%d", i)),
			"active",
		); err != nil {
			r.logger.Error("Failed to create merchant:", zap.Error(err))
			return err
		}
	}

	r.logger.Info("Merchant seeding completed successfully.", zap.Int("count", 10))
	return nil
}
