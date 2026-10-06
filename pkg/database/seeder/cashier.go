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

const seedGetCashiersSQL = `
SELECT cashier_id
FROM cashiers
WHERE deleted_at IS NULL
  AND ($1::TEXT IS NULL OR name ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

const seedCreateCashierSQL = `
INSERT INTO cashiers (merchant_id, user_id, name)
VALUES ($1, $2, $3) RETURNING cashier_id`

type seedCashierRow struct {
	CashierID int32 `db:"cashier_id"`
}

type cashierSeeder struct {
	db         *sqlx.DB
	merchantDB *sqlx.DB
	userDB     *sqlx.DB
	ctx        context.Context
	logger     logger.LoggerInterface
}

func NewCashierSeeder(db *sqlx.DB, merchantDB *sqlx.DB, userDB *sqlx.DB, ctx context.Context, logger logger.LoggerInterface) *cashierSeeder {
	return &cashierSeeder{db: db, merchantDB: merchantDB, userDB: userDB, ctx: ctx, logger: logger}
}

func (r *cashierSeeder) Seed() error {
	var merchants []seedMerchantRow
	if err := r.merchantDB.SelectContext(r.ctx, &merchants, seedGetMerchantsSQL, "", int32(20), 0); err != nil {
		r.logger.Error("Failed to fetch merchants:", zap.Any("err", err))
		return err
	}

	var users []seedUserRow
	if err := r.userDB.SelectContext(r.ctx, &users, seedGetUsersForRolesSQL, "", int32(20), 0); err != nil {
		r.logger.Error("Failed to fetch users:", zap.Any("error", err))
		return err
	}

	if len(merchants) == 0 || len(users) == 0 {
		r.logger.Error("Merchants or Users not found. Cashier seeding aborted.")
		return fmt.Errorf("no merchants or users found for cashier seeding")
	}

	rand.Seed(uint64(time.Now().UnixNano()))

	for i := 1; i <= 10; i++ {
		merchant := merchants[rand.Intn(len(merchants))]
		user := users[rand.Intn(len(users))]

		var created seedCashierRow
		if err := r.db.GetContext(r.ctx, &created, seedCreateCashierSQL,
			merchant.MerchantID, user.UserID, fmt.Sprintf("Cashier %d", i)); err != nil {
			r.logger.Error("Failed to create cashier:", zap.Any("error", err))
			return err
		}
	}

	r.logger.Info("Cashier seeding completed successfully.", zap.Int("count", 10))
	return nil
}
