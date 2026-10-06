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

const seedCreateOrderSQL = `
INSERT INTO orders (merchant_id, cashier_id, total_price)
VALUES ($1, $2, $3)
RETURNING order_id`

const seedCreateOrderItemSQL = `
INSERT INTO order_items (order_id, product_id, quantity, price)
VALUES ($1, $2, $3, $4)
RETURNING order_item_id`

const seedGetOrdersSQL = `
SELECT order_id
FROM orders
WHERE deleted_at IS NULL
AND ($1::TEXT IS NULL OR order_id::TEXT ILIKE '%' || $1 || '%' OR total_price::TEXT ILIKE '%' || $1 || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

const seedGetProductsForOrderSQL = `
SELECT product_id, price
FROM products p
JOIN categories c ON p.category_id = c.category_id
WHERE p.deleted_at IS NULL
  AND p.merchant_id = $1
  AND (
      p.name ILIKE '%' || COALESCE($2, '') || '%'
      OR p.description ILIKE '%' || COALESCE($2, '') || '%'
      OR $2 IS NULL
  )
  AND (
      c.category_id = NULLIF($3, 0)
      OR NULLIF($3, 0) IS NULL
  )
  AND (
      p.price >= COALESCE(NULLIF($4, 0), 0)
      AND p.price <= COALESCE(NULLIF($5, 0), 999999999)
  )
ORDER BY p.created_at DESC
LIMIT $6 OFFSET $7`

type seedOrderRow struct {
	OrderID int32 `db:"order_id"`
}

type orderSeeder struct {
	db         *sqlx.DB
	cashierDB  *sqlx.DB
	merchantDB *sqlx.DB
	productDB  *sqlx.DB
	ctx        context.Context
	logger     logger.LoggerInterface
}

func NewOrderSeeder(db *sqlx.DB, cashierDB *sqlx.DB, merchantDB *sqlx.DB, productDB *sqlx.DB, ctx context.Context, logger logger.LoggerInterface) *orderSeeder {
	return &orderSeeder{db: db, cashierDB: cashierDB, merchantDB: merchantDB, productDB: productDB, ctx: ctx, logger: logger}
}

func (r *orderSeeder) Seed() error {
	var merchants []seedMerchantRow
	if err := r.merchantDB.SelectContext(r.ctx, &merchants, seedGetMerchantsSQL, "", 20, 0); err != nil {
		r.logger.Error("Failed to get merchants", zap.Error(err))
		return err
	}

	var cashiers []seedCashierRow
	if err := r.cashierDB.SelectContext(r.ctx, &cashiers, seedGetCashiersSQL, "", 20, 0); err != nil {
		r.logger.Error("Failed to get cashiers", zap.Error(err))
		return err
	}

	if len(merchants) == 0 || len(cashiers) == 0 {
		r.logger.Error("No merchants or cashiers found. Order seeding aborted.")
		return fmt.Errorf("no merchants or cashiers found for order seeding")
	}

	rand.Seed(uint64(time.Now().UnixNano()))

	for i := 0; i < 10; i++ {
		merchant := merchants[rand.Intn(len(merchants))]
		cashier := cashiers[rand.Intn(len(cashiers))]
		totalPrice := int32(rand.Intn(500000) + 50000)

		var order seedOrderRow
		if err := r.db.GetContext(r.ctx, &order, seedCreateOrderSQL,
			merchant.MerchantID, cashier.CashierID, int64(totalPrice)); err != nil {
			r.logger.Error("Failed to create order", zap.Error(err))
			return err
		}

		orderID := order.OrderID

		var products []seedProductRow
		if err := r.productDB.SelectContext(r.ctx, &products, seedGetProductsForOrderSQL,
			merchant.MerchantID, nil, 0, 0, 0, 10, 0); err != nil {
			r.logger.Error("Failed to get products", zap.Error(err))
			return err
		}

		if len(products) == 0 {
			r.logger.Debug("No products found for merchant", zap.Int32("merchant_id", merchant.MerchantID))
			continue
		}

		for j := 0; j < rand.Intn(5)+1; j++ {
			product := products[rand.Intn(len(products))]
			quantity := int32(rand.Intn(5) + 1)
			price := product.Price * quantity

			if err := r.execSeed(r.ctx, seedCreateOrderItemSQL, orderID, product.ProductID, quantity, price); err != nil {
				r.logger.Error("Failed to create order item", zap.Error(err))
				return err
			}
		}
	}

	r.logger.Info("Order seeding completed successfully.", zap.Int("count", 10))
	return nil
}

func (r *orderSeeder) execSeed(ctx context.Context, query string, args ...any) error {
	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		return err
	}
	return nil
}
