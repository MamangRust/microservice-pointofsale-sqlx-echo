package cashier_test

import (
	"context"
	"testing"
	"time"

	cashier_cache "github.com/MamangRust/microservice-point-of-sale-cashier/cache"
	"github.com/MamangRust/microservice-point-of-sale-cashier/repository"
	"github.com/MamangRust/microservice-point-of-sale-cashier/service"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	"github.com/stretchr/testify/suite"
)

type CashierServiceTestSuite struct {
	tests.BaseTestSuite
	cashierService *service.Service
	cashierID      int
	userID         int
	merchantID     int
}

func (s *CashierServiceTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	s.SetupMerchantService()

	ctx := context.Background()

	// Seed dependencies
	s.userID = s.SeedUser(ctx)
	s.merchantID = s.SeedMerchant(ctx, s.userID)

	guardUser := resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, s.Log)
	guardMerchant := resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log)
	repos := repository.NewRepositories(
		s.SQLxDB(),
		pbuser.NewUserServiceClient(s.Conns["user"]),
		pbmerchant.NewMerchantServiceClient(s.Conns["merchant"]),
		repository.GuardOptions{
			User:     []adapter.GuardOption{adapter.WithDependencyGuard(guardUser)},
			Merchant: []adapter.GuardOption{adapter.WithDependencyGuard(guardMerchant)},
		},
	)

	mencache := cashier_cache.NewMencache(s.GetCacheStore())
	s.cashierService = service.NewService(&service.Deps{
		Ctx:           context.Background(),
		Repositories:  repos,
		Logger:        s.Log,
		Mencache:      mencache,
		Observability: s.Obs,
	})
}

func (s *CashierServiceTestSuite) TestCashierLifecycle() {
	ctx := context.Background()

	// 1. Create
	created, err := s.cashierService.CashierCommand.CreateCashier(ctx, &requests.CreateCashierRequest{
		MerchantID: s.merchantID,
		UserID:     s.userID,
		Name:       "Service Cashier",
	})
	s.Require().NoError(err)
	s.Require().NotNil(created)
	cashierID := int(created.CashierID)

	// 2. FindByID
	found, err := s.cashierService.CashierQuery.FindById(ctx, cashierID)
	s.Require().NoError(err)
	s.Equal("Service Cashier", found.Name)

	// 3. FindAll
	_, total, err := s.cashierService.CashierQuery.FindAll(ctx, &requests.FindAllCashiers{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*total, 1)

	// 4. FindByActive
	active, _, err := s.cashierService.CashierQuery.FindByActive(ctx, &requests.FindAllCashiers{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(active)

	// 5. Update
	updated, err := s.cashierService.CashierCommand.UpdateCashier(ctx, &requests.UpdateCashierRequest{
		CashierID: &cashierID,
		Name:      "Updated Service Cashier",
	})
	s.Require().NoError(err)
	s.Equal("Updated Service Cashier", updated.Name)

	// 6. Trash
	trashed, err := s.cashierService.CashierCommand.TrashedCashier(ctx, cashierID)
	s.Require().NoError(err)
	s.True(trashed.DeletedAt.Valid)

	// 7. FindByTrashed
	trashedList, totalTrashed, err := s.cashierService.CashierQuery.FindByTrashed(ctx, &requests.FindAllCashiers{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*totalTrashed, 1)
	s.NotEmpty(trashedList)

	// 8. Restore
	restored, err := s.cashierService.CashierCommand.RestoreCashier(ctx, cashierID)
	s.Require().NoError(err)
	s.False(restored.DeletedAt.Valid)

	// 9. FindByMerchant
	byMerchant, totalMerchant, err := s.cashierService.CashierQuery.FindByMerchant(ctx, &requests.FindAllCashierMerchant{
		MerchantID: s.merchantID,
		Search:     "",
		Page:       1,
		PageSize:   10,
	})
	s.Require().NoError(err)
	s.GreaterOrEqual(*totalMerchant, 1)
	s.NotEmpty(byMerchant)

	// 10. DeletePermanent (must trash first)
	_, _ = s.cashierService.CashierCommand.TrashedCashier(ctx, cashierID)
	deleted, err := s.cashierService.CashierCommand.DeleteCashierPermanent(ctx, cashierID)
	s.Require().NoError(err)
	s.True(deleted)
}

func (s *CashierServiceTestSuite) TestRestoreAllAndDeleteAll() {
	ctx := context.Background()

	// Create two cashiers
	c1, _ := s.cashierService.CashierCommand.CreateCashier(ctx, &requests.CreateCashierRequest{
		MerchantID: s.merchantID, UserID: s.userID, Name: "Bulk C1",
	})
	c2, _ := s.cashierService.CashierCommand.CreateCashier(ctx, &requests.CreateCashierRequest{
		MerchantID: s.merchantID, UserID: s.userID, Name: "Bulk C2",
	})

	// Trash both
	s.cashierService.CashierCommand.TrashedCashier(ctx, int(c1.CashierID))
	s.cashierService.CashierCommand.TrashedCashier(ctx, int(c2.CashierID))

	// RestoreAll
	res, err := s.cashierService.CashierCommand.RestoreAllCashier(ctx)
	s.Require().NoError(err)
	s.True(res)

	// Trash again
	s.cashierService.CashierCommand.TrashedCashier(ctx, int(c1.CashierID))
	s.cashierService.CashierCommand.TrashedCashier(ctx, int(c2.CashierID))

	// DeleteAll
	res, err = s.cashierService.CashierCommand.DeleteAllCashierPermanent(ctx)
	s.Require().NoError(err)
	s.True(res)
}

func TestCashierServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CashierServiceTestSuite))
}
