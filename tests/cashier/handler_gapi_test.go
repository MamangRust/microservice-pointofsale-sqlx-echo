package cashier_test

import (
	"context"
	cashier_cache "github.com/MamangRust/microservice-point-of-sale-cashier/cache"
	cashier_handler "github.com/MamangRust/microservice-point-of-sale-cashier/handler"
	cashier_repo "github.com/MamangRust/microservice-point-of-sale-cashier/repository"
	cashier_service "github.com/MamangRust/microservice-point-of-sale-cashier/service"
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"testing"
	"time"
)

type CashierGapiTestSuite struct {
	tests.BaseTestSuite
	client  pbcashier.CashierServiceClient
	userID  int
	merchID int
}

func (s *CashierGapiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	s.SetupMerchantService()

	ctx := context.Background()

	// Seed dependencies via gRPC (roles are auto-seeded by SetupTestSuite)
	s.userID = s.SeedUser(ctx)
	s.merchID = s.SeedMerchant(ctx, s.userID)

	// Setup cashier service
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	mencache := cashier_cache.NewMencache(cacheStore)
	guardUser := resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, s.Log)
	guardMerchant := resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log)
	repos := cashier_repo.NewRepositories(
		s.SQLxDB(),
		pbuser.NewUserServiceClient(s.Conns["user"]),
		pbmerchant.NewMerchantServiceClient(s.Conns["merchant"]),
		cashier_repo.GuardOptions{
			User:     []adapter.GuardOption{adapter.WithDependencyGuard(guardUser)},
			Merchant: []adapter.GuardOption{adapter.WithDependencyGuard(guardMerchant)},
		},
	)
	svc := cashier_service.NewService(&cashier_service.Deps{
		Ctx:           context.Background(),
		Repositories:  repos,
		Logger:        s.Log,
		Mencache:      mencache,
		Observability: s.Obs,
	})

	handler := cashier_handler.NewHandler(&cashier_handler.Deps{
		Service: svc,
		Logger:  s.Log,
	})

	server := grpc.NewServer()
	pbcashier.RegisterCashierServiceServer(server, handler.Cashier)

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)
	s.client = pbcashier.NewCashierServiceClient(conn)
}

func (s *CashierGapiTestSuite) TestCashierGapiLifecycle() {
	ctx := context.Background()

	// 1. Create
	createRes, err := s.client.CreateCashier(ctx, &pbcashier.CreateCashierRequest{
		MerchantId: int32(s.merchID),
		UserId:     int32(s.userID),
		Name:       "GAPI Cashier",
	})
	s.NoError(err)
	s.NotNil(createRes)
	cashierID := createRes.Data.Id

	// 2. FindById
	getRes, err := s.client.FindById(ctx, &pbcashier.FindByIdCashierRequest{Id: cashierID})
	s.NoError(err)
	s.Equal("GAPI Cashier", getRes.Data.Name)

	// 3. FindAll
	allRes, err := s.client.FindAll(ctx, &pbcashier.FindAllCashierRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(allRes.Data)

	// 4. FindByActive
	activeRes, err := s.client.FindByActive(ctx, &pbcashier.FindAllCashierRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(activeRes.Data)

	// 5. FindByMerchant
	merchantRes, err := s.client.FindByMerchant(ctx, &pbcashier.FindByMerchantCashierRequest{
		MerchantId: int32(s.merchID),
		Page:       1,
		PageSize:   10,
	})
	s.NoError(err)
	s.NotEmpty(merchantRes.Data)

	// 6. Update
	updateRes, err := s.client.UpdateCashier(ctx, &pbcashier.UpdateCashierRequest{
		CashierId: cashierID,
		Name:      "GAPI Cashier Updated",
	})
	s.NoError(err)
	s.Equal("GAPI Cashier Updated", updateRes.Data.Name)

	// 7. Trash
	_, err = s.client.TrashedCashier(ctx, &pbcashier.FindByIdCashierRequest{Id: cashierID})
	s.NoError(err)

	// 8. FindByTrashed
	trashedRes, err := s.client.FindByTrashed(ctx, &pbcashier.FindAllCashierRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(trashedRes.Data)

	// 9. Restore
	_, err = s.client.RestoreCashier(ctx, &pbcashier.FindByIdCashierRequest{Id: cashierID})
	s.NoError(err)

	// 10. DeletePermanent (trash first)
	_, _ = s.client.TrashedCashier(ctx, &pbcashier.FindByIdCashierRequest{Id: cashierID})
	_, err = s.client.DeleteCashierPermanent(ctx, &pbcashier.FindByIdCashierRequest{Id: cashierID})
	s.NoError(err)

	// 11. RestoreAll
	_, err = s.client.RestoreAllCashier(ctx, &emptypb.Empty{})
	s.NoError(err)

	// 12. DeleteAll
	_, err = s.client.DeleteAllCashierPermanent(ctx, &emptypb.Empty{})
	s.NoError(err)
}

func TestCashierGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CashierGapiTestSuite))
}
