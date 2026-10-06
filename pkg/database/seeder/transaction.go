package seeder

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const seedGetOrdersForTransactionSQL = `
SELECT order_id
FROM orders
WHERE deleted_at IS NULL
AND ($1::TEXT IS NULL OR order_id::TEXT ILIKE '%' || $1 || '%' OR total_price::TEXT ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

const seedCreateTransactionSQL = `
INSERT INTO transactions (merchant_id, payment_method, amount, change_amount, payment_status, order_id, created_at, updated_at, deleted_at)
VALUES ($1, $2, $3, $4, $5, $6, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP, NULL)
RETURNING transaction_id`

type transactionSeeder struct {
	db         *sqlx.DB
	merchantDB *sqlx.DB
	orderDB    *sqlx.DB
	ctx        context.Context
	logger     logger.LoggerInterface
}

func NewTransactionSeeder(db *sqlx.DB, merchantDB *sqlx.DB, orderDB *sqlx.DB, ctx context.Context, logger logger.LoggerInterface) *transactionSeeder {
	return &transactionSeeder{db: db, merchantDB: merchantDB, orderDB: orderDB, ctx: ctx, logger: logger}
}

func (r *transactionSeeder) Seed() error {
	var orders []seedOrderRow
	if err := r.orderDB.SelectContext(r.ctx, &orders, seedGetOrdersForTransactionSQL, "", 20, 0); err != nil {
		r.logger.Error("Failed to get transactions:", zap.Any("error", err))
		return err
	}

	var merchants []seedMerchantRow
	if err := r.merchantDB.SelectContext(r.ctx, &merchants, seedGetMerchantsSQL, "", 20, 0); err != nil {
		r.logger.Error("Failed to get transactions:", zap.Any("error", err))
		return err
	}

	if len(orders) == 0 || len(merchants) == 0 {
		r.logger.Error("No orders or merchants found. Transaction seeding aborted.")
		return fmt.Errorf("no orders or merchants found for transaction seeding")
	}

	for i := 0; i < 10; i++ {
		selectedMerchant := merchants[rand.Intn(len(merchants))]
		selectedOrder := orders[rand.Intn(len(orders))]

		var paymentMethod string
		var amount, changeAmount float64
		var paymentStatus string

		paymentMethod = "Credit Card"
		amount = float64(100 + i)
		changeAmount = float64(5 + i)
		paymentStatus = "Completed"

		if _, err := r.db.ExecContext(r.ctx, seedCreateTransactionSQL,
			selectedMerchant.MerchantID, paymentMethod, int32(amount), ptrInt32(int32(changeAmount)),
			paymentStatus, selectedOrder.OrderID); err != nil {
			r.logger.Error("Failed to create transaction:", zap.Any("error", err))
			return err
		}
	}

	r.logger.Info("Successfully seeded 10 transactions.", zap.Int("count", 10))
	return nil
}
