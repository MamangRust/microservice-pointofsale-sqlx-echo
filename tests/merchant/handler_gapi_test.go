package merchant_test

import (
	"context"
	"testing"
	"time"

	merchant_cache "github.com/MamangRust/microservice-point-of-sale-merchant/cache"
	"github.com/MamangRust/microservice-point-of-sale-merchant/handler"
	"github.com/MamangRust/microservice-point-of-sale-merchant/repository"
	"github.com/MamangRust/microservice-point-of-sale-merchant/service"
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
)

type MerchantGapiTestSuite struct {
	tests.BaseTestSuite
	client     pbmerchant.MerchantServiceClient
	userID     int
	merchantID int
}

func (s *MerchantGapiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	guardUser := resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, s.Log)
	repos := repository.NewRepositories(
		s.SQLxDB(), pbuser.NewUserServiceClient(s.Conns["user"]),
		repository.GuardOptions{
			User: []adapter.GuardOption{adapter.WithDependencyGuard(guardUser)},
		},
	)

	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	mencache := merchant_cache.NewMencache(cacheStore)

	svc := service.NewService(&service.Deps{
		Kafka:         nil,
		Repositories:  repos,
		Logger:        s.Log,
		Mencache:      mencache,
		Observability: s.Obs,
	})

	merchantHandler := handler.NewHandler(&handler.Deps{
		Service: svc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pbmerchant.RegisterMerchantServiceServer(server, merchantHandler.Merchant)

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)

	s.client = pbmerchant.NewMerchantServiceClient(conn)

	// 1. Seed dependencies
	s.userID = s.SeedUser(context.Background())
}

func (s *MerchantGapiTestSuite) TestMerchantGapiLifecycle() {
	ctx := context.Background()

	// 1. Create
	createReq := &pbmerchant.CreateMerchantRequest{
		UserId:       int32(s.userID),
		Name:         "Gapi Merchant",
		Description:  "Detailed description of the merchant.",
		Address:      "Merchant Street No. 1",
		ContactEmail: "gapi.merchant@example.com",
		ContactPhone: "08123456789",
		Status:       "active",
	}
	res, err := s.client.Create(ctx, createReq)
	s.NoError(err)
	s.Equal(createReq.Name, res.Data.Name)
	merchantID := res.Data.Id

	// 2. FindById
	found, err := s.client.FindById(ctx, &pbmerchant.FindByIdMerchantRequest{Id: merchantID})
	s.NoError(err)
	s.Equal(merchantID, found.Data.Id)

	// 3. FindAll
	allRes, err := s.client.FindAll(ctx, &pbmerchant.FindAllMerchantRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(allRes.Data)

	// 4. FindByActive
	activeRes, err := s.client.FindByActive(ctx, &pbmerchant.FindAllMerchantRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(activeRes.Data)

	// 5. Update
	updateReq := &pbmerchant.UpdateMerchantRequest{
		MerchantId:   merchantID,
		UserId:       int32(s.userID),
		Name:         "Gapi Merchant Updated",
		Description:  "Updated description.",
		Address:      "New Street 2",
		ContactEmail: "updated@example.com",
		ContactPhone: "08987654321",
		Status:       "waiting",
	}
	updateRes, err := s.client.Update(ctx, updateReq)
	s.NoError(err)
	s.Equal(updateReq.Name, updateRes.Data.Name)

	// 6. Update Status
	statusRes, err := s.client.UpdateMerchantStatus(ctx, &pbmerchant.UpdateMerchantStatusRequest{
		MerchantId: merchantID,
		Status:     "active",
	})
	s.NoError(err)
	s.Equal("active", statusRes.Data.Status)

	// 7. Trash
	_, err = s.client.TrashedMerchant(ctx, &pbmerchant.FindByIdMerchantRequest{Id: merchantID})
	s.NoError(err)

	// 8. FindByTrashed
	trashedRes, err := s.client.FindByTrashed(ctx, &pbmerchant.FindAllMerchantRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(trashedRes.Data)

	// 9. Restore
	_, err = s.client.RestoreMerchant(ctx, &pbmerchant.FindByIdMerchantRequest{Id: merchantID})
	s.NoError(err)

	// 10. DeletePermanent
	_, _ = s.client.TrashedMerchant(ctx, &pbmerchant.FindByIdMerchantRequest{Id: merchantID})
	_, err = s.client.DeleteMerchantPermanent(ctx, &pbmerchant.FindByIdMerchantRequest{Id: merchantID})
	s.NoError(err)

	// 11. RestoreAll
	_, err = s.client.RestoreAllMerchant(ctx, &emptypb.Empty{})
	s.NoError(err)

	// 12. DeleteAll
	_, err = s.client.DeleteAllMerchantPermanent(ctx, &emptypb.Empty{})
	s.NoError(err)
}

func TestMerchantGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantGapiTestSuite))
}
