package cashier_test

import (
	"context"
	"github.com/MamangRust/microservice-point-of-sale-cashier/repository"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
)

type CashierRepositoryTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	repo       *repository.Repositories
	cashierID  int
	userID     int
	merchantID int
}

func (s *CashierRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	pool, err := sqlx.ConnectContext(s.ts.Ctx, "pgx", s.ts.DBURL)
	s.Require().NoError(err)

	// Seed user and merchant dependencies
	var userID int
	err = pool.QueryRowContext(s.ts.Ctx, `INSERT INTO users (firstname, lastname, email, password, verification_code)
		VALUES ('Cashier', 'Repo', 'cashier.repo@example.com', 'password', 'CODE') RETURNING user_id`).Scan(&userID)
	s.Require().NoError(err)
	s.userID = userID

	var merchantID int
	err = pool.QueryRowContext(s.ts.Ctx, `INSERT INTO merchants (user_id, name, status) VALUES ($1, 'Cashier Merchant', 'active') RETURNING merchant_id`, userID).Scan(&merchantID)
	s.Require().NoError(err)
	s.merchantID = merchantID

	queries := pool
	s.repo = repository.NewRepositories(queries, nil, nil)
}

func (s *CashierRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *CashierRepositoryTestSuite) Test1_CreateCashier() {
	ctx := context.Background()

	cashier, err := s.repo.CashierCommand.CreateCashier(ctx, &requests.CreateCashierRequest{
		MerchantID: s.merchantID,
		UserID:     s.userID,
		Name:       "Repo Cashier",
	})
	s.NoError(err)
	s.NotNil(cashier)
	s.Equal("Repo Cashier", cashier.Name)
	s.Equal(int32(s.merchantID), cashier.MerchantID)
	s.Equal(int32(s.userID), cashier.UserID)
	s.cashierID = int(cashier.CashierID)
}

func (s *CashierRepositoryTestSuite) Test2_FindById() {
	s.Require().NotZero(s.cashierID)
	ctx := context.Background()

	found, err := s.repo.CashierQuery.FindById(ctx, s.cashierID)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(s.cashierID, int(found.CashierID))
	s.Equal("Repo Cashier", found.Name)
}

func (s *CashierRepositoryTestSuite) Test3_UpdateCashier() {
	s.Require().NotZero(s.cashierID)
	ctx := context.Background()

	updated, err := s.repo.CashierCommand.UpdateCashier(ctx, &requests.UpdateCashierRequest{
		CashierID: &s.cashierID,
		Name:      "Updated Repo Cashier",
	})
	s.NoError(err)
	s.NotNil(updated)
	s.Equal("Updated Repo Cashier", updated.Name)
}

func (s *CashierRepositoryTestSuite) Test4_FindAll() {
	ctx := context.Background()

	results, total, err := s.repo.CashierQuery.FindAllCashiers(ctx, &requests.FindAllCashiers{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	s.NoError(err)
	s.NotNil(total)
	s.GreaterOrEqual(*total, 1)
	s.GreaterOrEqual(len(results), 1)
}

func (s *CashierRepositoryTestSuite) Test5_FindByActive() {
	ctx := context.Background()

	results, total, err := s.repo.CashierQuery.FindByActive(ctx, &requests.FindAllCashiers{
		Search:   "",
		Page:     1,
		PageSize: 10,
	})
	s.NoError(err)
	s.NotNil(total)
	s.GreaterOrEqual(*total, 1)
	s.NotEmpty(results)
}

func (s *CashierRepositoryTestSuite) Test6_TrashAndRestore() {
	s.Require().NotZero(s.cashierID)
	ctx := context.Background()

	// Trash
	trashed, err := s.repo.CashierCommand.TrashedCashier(ctx, s.cashierID)
	s.NoError(err)
	s.NotNil(trashed)
	s.True(trashed.DeletedAt.Valid)

	// Verify trashed
	trashedResults, _, err := s.repo.CashierQuery.FindByTrashed(ctx, &requests.FindAllCashiers{Search: "", Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(trashedResults)

	// Restore
	restored, err := s.repo.CashierCommand.RestoreCashier(ctx, s.cashierID)
	s.NoError(err)
	s.NotNil(restored)
	s.False(restored.DeletedAt.Valid)
}

func (s *CashierRepositoryTestSuite) Test7_FindByMerchant() {
	ctx := context.Background()

	results, total, err := s.repo.CashierQuery.FindByMerchant(ctx, &requests.FindAllCashierMerchant{
		MerchantID: s.merchantID,
		Search:     "",
		Page:       1,
		PageSize:   10,
	})
	s.NoError(err)
	s.NotNil(total)
	s.GreaterOrEqual(*total, 1)
	s.GreaterOrEqual(len(results), 1)
}

func (s *CashierRepositoryTestSuite) Test8_DeletePermanent() {
	s.Require().NotZero(s.cashierID)
	ctx := context.Background()

	// Must be trashed first
	_, err := s.repo.CashierCommand.TrashedCashier(ctx, s.cashierID)
	s.NoError(err)

	deleted, err := s.repo.CashierCommand.DeleteCashierPermanent(ctx, s.cashierID)
	s.NoError(err)
	s.True(deleted)

	// Verify deleted
	_, err = s.repo.CashierQuery.FindById(ctx, s.cashierID)
	s.Error(err)
}

func TestCashierRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CashierRepositoryTestSuite))
}
