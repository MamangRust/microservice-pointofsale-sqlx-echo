package seeder

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/jmoiron/sqlx"
)

type Deps struct {
	User        *sqlx.DB
	Role        *sqlx.DB
	Merchant    *sqlx.DB
	Cashier     *sqlx.DB
	Category    *sqlx.DB
	Product     *sqlx.DB
	Order       *sqlx.DB
	Transaction *sqlx.DB
	Ctx         context.Context
	Logger      logger.LoggerInterface
	Hash        hash.HashPassword
}

type Seeder struct {
	User        *userSeeder
	Role        *roleSeeder
	UserRole    *userRoleSeeder
	Cashier     *cashierSeeder
	Category    *categorySeeder
	Product     *productSeeder
	Merchant    *merchantSeeder
	Order       *orderSeeder
	Transaction *transactionSeeder
}

func NewSeeder(deps Deps) *Seeder {
	return &Seeder{
		User:        NewUserSeeder(deps.User, deps.Hash, deps.Ctx, deps.Logger),
		Role:        NewRoleSeeder(deps.Role, deps.Ctx, deps.Logger),
		UserRole:    NewUserRoleSeeder(deps.Role, deps.User, deps.Ctx, deps.Logger),
		Merchant:    NewMerchantSeeder(deps.Merchant, deps.User, deps.Ctx, deps.Logger),
		Cashier:     NewCashierSeeder(deps.Cashier, deps.Merchant, deps.User, deps.Ctx, deps.Logger),
		Category:    NewCategorySeeder(deps.Category, deps.Ctx, deps.Logger),
		Product:     NewProductSeeder(deps.Product, deps.Category, deps.Merchant, deps.Ctx, deps.Logger),
		Order:       NewOrderSeeder(deps.Order, deps.Cashier, deps.Merchant, deps.Product, deps.Ctx, deps.Logger),
		Transaction: NewTransactionSeeder(deps.Transaction, deps.Merchant, deps.Order, deps.Ctx, deps.Logger),
	}
}

func (s *Seeder) Run() error {
	if err := s.seedWithDelay("users", s.User.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("roles", s.Role.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("user_roles", s.UserRole.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("merchant", s.Merchant.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("cashier", s.Cashier.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("category", s.Category.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("product", s.Product.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("order", s.Order.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("transaction", s.Transaction.Seed); err != nil {
		return err
	}

	return nil
}

func (s *Seeder) seedWithDelay(entityName string, seedFunc func() error) error {
	if err := seedFunc(); err != nil {
		return fmt.Errorf("failed to seed %s: %w", entityName, err)
	}
	time.Sleep(seedDelay())
	return nil
}

// seedDelay returns the pause after each seeded entity. Default 30s (historical
// behavior); SEED_DELAY_SECONDS overrides it so CI/e2e runs can finish quickly
// (run_e2e.sh sets it to 1).
func seedDelay() time.Duration {
	if raw := os.Getenv("SEED_DELAY_SECONDS"); raw != "" {
		if secs, err := strconv.Atoi(raw); err == nil && secs >= 0 {
			return time.Duration(secs) * time.Second
		}
	}
	return 30 * time.Second
}

func ptrString(s string) *string {
	return &s
}

func ptrInt32(i int32) *int32 {
	return &i
}
