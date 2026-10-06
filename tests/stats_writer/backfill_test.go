// Package stats_writer_test exercises the stats-writer backfill end to end: it
// seeds the owning PostgreSQL tables (catalog + sales + merchant contexts),
// starts the real order/transaction/product/order_item gRPC services, runs the
// backfiller — which reads every source through those services — against a real
// ClickHouse, and asserts the materialized rows.
package stats_writer_test

import (
	"fmt"
	"testing"
	"time"

	chDriver "github.com/ClickHouse/clickhouse-go/v2"
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	cashieradapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/cashier"
	orderadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/order"
	orderitemadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/orderitem"
	productadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/product"
	transactionadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/transaction"
	pkgclickhouse "github.com/MamangRust/microservice-point-of-sale-pkg/clickhouse"
	"github.com/MamangRust/microservice-point-of-sale-stats-writer/backfill"
	statsrepo "github.com/MamangRust/microservice-point-of-sale-stats-writer/repository"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	"github.com/stretchr/testify/suite"
)

const (
	// validItemCount is deliberately larger than the backfiller's 1000-row page
	// size so the walk has to cross the page boundary and stop on a short page.
	validItemCount = 1005
	orphanProduct  = 999999

	productOnePrice = 1000
	productTwoPrice = 2000
	orderTotalPrice = 1507000

	// seedCashierID is the cashier on the seeded order. Transactions carry no
	// cashier_id themselves; the backfill must recover it from the order.
	seedCashierID = 7
)

type BackfillSuite struct {
	tests.BaseTestSuite

	chConn chDriver.Conn
	repo   statsrepo.Repository

	merchantID int32
	orderID    int32
	categoryA  int32
	categoryB  int32
	productOne int32
	productTwo int32
	cashierOne int32
	cashierTwo int32

	// expectations derived from the seeded rows (the source of truth)
	expMinItemID   int32
	expMaxItemID   int32
	expSumSubtotal int64
	expCategoryA   uint64
	expCategoryB   uint64
}

func (s *BackfillSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupStatsEnv()

	// Brings up order (and its cashier/merchant/product/order_item deps),
	// transaction and product services over real gRPC.
	s.SetupTransactionService()
	s.SetupProductService()

	conn, err := pkgclickhouse.NewClient(s.Log)
	s.Require().NoError(err)
	s.Require().NoError(pkgclickhouse.ApplySchema(s.Ctx, conn, s.Log))

	s.chConn = conn
	s.repo = statsrepo.NewClickhouseRepository(conn, s.Log)

	s.seedFixtures()
}

func (s *BackfillSuite) TearDownSuite() {
	if s.repo != nil {
		_ = s.repo.Close()
	}
	if s.chConn != nil {
		_ = s.chConn.Close()
	}
	s.BaseTestSuite.TearDownSuite()
}

// seedFixtures writes the fixture rows into the table that owns each entity:
// categories/products into the catalog context, orders/order_items/transactions
// into the sales context, and merchants/cashiers into the merchant context. The
// test harness collapses all five bounded contexts onto a single PostgreSQL
// container (see BaseTestSuite.SetupStatsEnv), so one pool is enough here.
//
// The backfill itself never reads PostgreSQL — it goes through the owning
// services — so this is purely a data setup step.
func (s *BackfillSuite) seedFixtures() {
	ctx := s.Ctx
	db := s.SQLxDB()
	now := time.Now().UTC()

	// Merchant context: one merchant, then two cashiers under it. The cashiers
	// FK points at merchants, so the merchant has to exist first.
	s.Require().NoError(db.QueryRowContext(ctx, `
		INSERT INTO merchants (user_id, name, status, created_at, updated_at)
		VALUES ($1, 'Seed Merchant', 'active', $2, $2)
		RETURNING merchant_id
	`, int32(1), now).Scan(&s.merchantID))

	s.Require().NoError(db.QueryRowContext(ctx, `
		INSERT INTO cashiers (merchant_id, user_id, name, created_at, updated_at)
		VALUES ($1, $2, 'C1', $3, $3)
		RETURNING cashier_id
	`, s.merchantID, int32(1), now).Scan(&s.cashierOne))
	s.Require().NoError(db.QueryRowContext(ctx, `
		INSERT INTO cashiers (merchant_id, user_id, name, created_at, updated_at)
		VALUES ($1, $2, 'C2', $3, $3)
		RETURNING cashier_id
	`, s.merchantID, int32(1), now).Scan(&s.cashierTwo))

	// Catalog context: categories → products. slug_product and barcode stay
	// NULL so the unique constraints never collide.
	s.Require().NoError(db.QueryRowContext(ctx, `
		INSERT INTO categories (name, description, created_at, updated_at)
		VALUES ('Cat A', 'a', $1, $1)
		RETURNING category_id
	`, now).Scan(&s.categoryA))
	s.Require().NoError(db.QueryRowContext(ctx, `
		INSERT INTO categories (name, description, created_at, updated_at)
		VALUES ('Cat B', 'b', $1, $1)
		RETURNING category_id
	`, now).Scan(&s.categoryB))

	s.Require().NoError(db.QueryRowContext(ctx, `
		INSERT INTO products (merchant_id, category_id, name, price, count_in_stock, created_at, updated_at)
		VALUES ($1, $2, 'P1', $3, 100, $4, $4)
		RETURNING product_id
	`, s.merchantID, s.categoryA, productOnePrice, now).Scan(&s.productOne))
	s.Require().NoError(db.QueryRowContext(ctx, `
		INSERT INTO products (merchant_id, category_id, name, price, count_in_stock, created_at, updated_at)
		VALUES ($1, $2, 'P2', $3, 100, $4, $4)
		RETURNING product_id
	`, s.merchantID, s.categoryB, productTwoPrice, now).Scan(&s.productTwo))

	// Sales context: one order.
	s.Require().NoError(db.QueryRowContext(ctx, `
		INSERT INTO orders (merchant_id, cashier_id, total_price, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $4)
		RETURNING order_id
	`, s.merchantID, int32(seedCashierID), int64(orderTotalPrice), now).Scan(&s.orderID))

	// Sales context: order items. Rows with an odd series index use product one
	// and even rows product two, so both catalog categories are exercised.
	_, err := db.ExecContext(ctx, `
		INSERT INTO order_items (order_id, product_id, quantity, price, created_at, updated_at)
		SELECT $1,
		       CASE WHEN g % 2 = 1 THEN $2::INT ELSE $3::INT END,
		       1,
		       CASE WHEN g % 2 = 1 THEN $4::INT ELSE $5::INT END,
		       $6,
		       $6
		FROM generate_series(1, $7) AS g
	`, s.orderID, s.productOne, s.productTwo, productOnePrice, productTwoPrice, now, validItemCount)
	s.Require().NoError(err)

	// The orphan is appended last so it lands on the second page and points at a
	// product that does not exist — the backfill must skip it.
	_, err = db.ExecContext(ctx, `
		INSERT INTO order_items (order_id, product_id, quantity, price, created_at, updated_at)
		VALUES ($1, $2, 1, 500, $3, $3)
	`, s.orderID, int32(orphanProduct), now)
	s.Require().NoError(err)

	// Derive the expectations from the same parity rule the insert used.
	for i := 0; i < validItemCount; i++ {
		if i%2 == 0 {
			s.expCategoryA++
			s.expSumSubtotal += productOnePrice
		} else {
			s.expCategoryB++
			s.expSumSubtotal += productTwoPrice
		}
	}

	// The walkable id range excludes the orphan, which the backfill drops.
	s.Require().NoError(db.QueryRowContext(ctx, `
		SELECT min(order_item_id), max(order_item_id)
		FROM order_items
		WHERE product_id <> $1
	`, int32(orphanProduct)).Scan(&s.expMinItemID, &s.expMaxItemID))

	// Sales context: two transactions for the order.
	paid := "success"
	_, err = db.ExecContext(ctx, `
		INSERT INTO transactions (order_id, merchant_id, payment_method, amount, payment_status, created_at, updated_at)
		VALUES ($1, $2, 'cash', 1000000, $3, $4, $4),
		       ($1, $2, 'transfer', 500000, $3, $4, $4)
	`, s.orderID, s.merchantID, paid, now)
	s.Require().NoError(err)
}

func (s *BackfillSuite) TestBackfillMaterializesEverySource() {
	orderRepo := orderadapter.New(pborder.NewOrderServiceClient(s.Conns["order"]))
	orderItemRepo := orderitemadapter.New(pborderitem.NewOrderItemServiceClient(s.Conns["order-item"]))
	productRepo := productadapter.New(pbproduct.NewProductServiceClient(s.Conns["product"]))
	transactionRepo := transactionadapter.New(pbtransaction.NewTransactionServiceClient(s.Conns["transaction"]))
	cashierRepo := cashieradapter.New(pbcashier.NewCashierServiceClient(s.Conns["cashier"]))

	bf := backfill.New(s.Log, s.repo, orderRepo, orderItemRepo, productRepo, transactionRepo, cashierRepo)
	s.Require().NoError(bf.Run(s.Ctx))

	s.Equal(uint64(validItemCount), s.count(`SELECT count() FROM order_item_daily FINAL`),
		"the orphan item must be skipped and both pages must be walked")

	var minID, maxID uint64
	s.Require().NoError(s.chConn.QueryRow(s.Ctx,
		`SELECT min(order_item_id), max(order_item_id) FROM order_item_daily FINAL`,
	).Scan(&minID, &maxID))
	s.Equal(uint64(s.expMinItemID), minID, "first page must be reached")
	s.Equal(uint64(s.expMaxItemID), maxID, "second page must be reached")

	s.Equal(s.expCategoryA,
		s.count(fmt.Sprintf(`SELECT count() FROM order_item_daily FINAL WHERE category_id = %d`, s.categoryA)),
		"category must be denormalized from the catalog")
	s.Equal(s.expCategoryB,
		s.count(fmt.Sprintf(`SELECT count() FROM order_item_daily FINAL WHERE category_id = %d`, s.categoryB)),
		"category must be denormalized from the catalog")

	s.Equal(uint64(0), s.count(
		`SELECT count() FROM order_item_daily FINAL WHERE subtotal != toInt64(quantity) * unit_price`),
		"subtotal must equal quantity * unit_price")

	var sumSubtotal int64
	s.Require().NoError(s.chConn.QueryRow(s.Ctx,
		`SELECT sum(subtotal) FROM order_item_daily FINAL`).Scan(&sumSubtotal))
	s.Equal(s.expSumSubtotal, sumSubtotal)

	s.Equal(uint64(0),
		s.count(fmt.Sprintf(`SELECT count() FROM order_item_daily FINAL WHERE product_id = %d`, orphanProduct)),
		"orphan order item must not be materialized")

	s.Equal(uint64(1), s.count(`SELECT count() FROM order_daily FINAL`))

	var totalPrice int64
	s.Require().NoError(s.chConn.QueryRow(s.Ctx,
		`SELECT total_price FROM order_daily FINAL LIMIT 1`).Scan(&totalPrice))
	s.Equal(int64(orderTotalPrice), totalPrice)

	s.Equal(uint64(2), s.count(`SELECT count() FROM transaction_daily FINAL`))

	// The transaction rows must carry the cashier recovered from their order.
	s.Equal(uint64(2),
		s.count(fmt.Sprintf(`SELECT count() FROM transaction_daily FINAL WHERE cashier_id = %d`, seedCashierID)),
		"cashier_id must be recovered from the owning order")

	// Cashiers are materialized from the cashier service into cashier_daily.
	s.Equal(uint64(2), s.count(`SELECT count() FROM cashier_daily FINAL`),
		"both cashiers must be materialized")
	s.Equal(uint64(2), s.count(fmt.Sprintf(
		`SELECT count() FROM cashier_daily FINAL WHERE merchant_id = %d AND status = 'active'`, s.merchantID)),
		"cashier rows must carry merchant and active status")
}

func (s *BackfillSuite) count(query string) uint64 {
	var n uint64
	s.Require().NoError(s.chConn.QueryRow(s.Ctx, query).Scan(&n))
	return n
}

func TestBackfillSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test requires Docker")
	}
	suite.Run(t, new(BackfillSuite))
}
