package order_test

import (
	"context"
	"testing"
	"time"

	order_cache "github.com/MamangRust/microservice-point-of-sale-order/cache"
	order_handler "github.com/MamangRust/microservice-point-of-sale-order/handler"
	order_repo "github.com/MamangRust/microservice-point-of-sale-order/repository"
	order_service "github.com/MamangRust/microservice-point-of-sale-order/service"
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OrderGapiTestSuite struct {
	tests.BaseTestSuite
	client pborder.OrderServiceClient
}

func (s *OrderGapiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	// Setup dependencies
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupOrderItemService()
	s.SetupCashierService()
	s.SetupTransactionService()

	// Infrastructure
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	// Order dependencies
	mencache := order_cache.NewMencache(cacheStore)
	guardCashier := resilience.NewDependencyGuard("cashier", 5, 30, 100, 3*time.Second, s.Log)
	guardMerchant := resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log)
	guardProduct := resilience.NewDependencyGuard("product", 5, 30, 100, 3*time.Second, s.Log)
	guardOrderItem := resilience.NewDependencyGuard("order_item", 5, 30, 100, 3*time.Second, s.Log)
	repos := order_repo.NewRepositories(
		s.SQLxDB(),
		pbcashier.NewCashierServiceClient(s.Conns["cashier"]),
		pbmerchant.NewMerchantServiceClient(s.Conns["merchant"]),
		pbproduct.NewProductServiceClient(s.Conns["product"]),
		pborderitem.NewOrderItemServiceClient(s.Conns["order-item"]),
		order_repo.GuardOptions{
			Cashier:   []adapter.GuardOption{adapter.WithDependencyGuard(guardCashier)},
			Merchant:  []adapter.GuardOption{adapter.WithDependencyGuard(guardMerchant)},
			Product:   []adapter.GuardOption{adapter.WithDependencyGuard(guardProduct)},
			OrderItem: []adapter.GuardOption{adapter.WithDependencyGuard(guardOrderItem)},
		},
	)
	svc := order_service.NewService(&order_service.Deps{
		Mencache:      mencache,
		Repositories:  repos,
		Logger:        s.Log,
		Observability: s.Obs,
	})

	// Handler
	handler := order_handler.NewHandler(&order_handler.Deps{
		Service: svc,
		Logger:  s.Log,
	})

	// Server
	server := grpc.NewServer()
	pborder.RegisterOrderServiceServer(server, handler.Order)

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)

	s.client = pborder.NewOrderServiceClient(conn)
}

func (s *OrderGapiTestSuite) TestOrderGapiLifecycle() {
	ctx := context.Background()

	// 1. Seed dependencies
	userID := s.SeedUser(ctx)
	catID := s.SeedCategory(ctx)
	merchID := s.SeedMerchant(ctx, userID)
	prodID := s.SeedProduct(ctx, merchID, catID)

	// Seed a cashier (orders.cashier_id references cashiers.cashier_id)
	var cashierID int
	err := s.SQLxDB().QueryRowContext(ctx,
		`INSERT INTO cashiers (merchant_id, user_id, name) VALUES ($1, $2, 'Order Gapi Cashier') RETURNING cashier_id`,
		merchID, userID,
	).Scan(&cashierID)
	s.Require().NoError(err)

	// 2. Create
	createRes, err := s.client.Create(ctx, &pborder.CreateOrderRequest{
		MerchantId: int32(merchID),
		CashierId:  int32(cashierID),
		Items: []*pborder.CreateOrderItemRequest{
			{
				ProductId: int32(prodID),
				Quantity:  1,
			},
		},
	})
	s.Require().NoError(err)
	s.Require().NotNil(createRes)
	orderID := createRes.Data.Id

	// 3. FindById
	getRes, err := s.client.FindById(ctx, &pborder.FindByIdOrderRequest{Id: orderID})
	s.Require().NoError(err)
	s.Equal(int32(userID), getRes.Data.CashierId)

	// 4. FindAll
	allRes, err := s.client.FindAll(ctx, &pborder.FindAllOrderRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(allRes.Data)

	// 5. FindByActive
	activeRes, err := s.client.FindByActive(ctx, &pborder.FindAllOrderRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(activeRes.Data)

	// 6. Update
	// Fetch order items first
	itemClient := pborderitem.NewOrderItemServiceClient(s.Conns["order-item"])
	itemsRes, err := itemClient.FindOrderItemByOrder(ctx, &pborderitem.FindByIdOrderItemRequest{Id: orderID})
	s.Require().NoError(err)
	s.NotEmpty(itemsRes.Data)
	orderItemID := itemsRes.Data[0].Id

	_, err = s.client.Update(ctx, &pborder.UpdateOrderRequest{
		OrderId: orderID,
		Items: []*pborder.UpdateOrderItemRequest{
			{
				OrderItemId: orderItemID,
				ProductId:   int32(prodID),
				Quantity:    1,
			},
		},
	})
	s.Require().NoError(err)

	// 7. Trash
	_, err = s.client.TrashedOrder(ctx, &pborder.FindByIdOrderRequest{Id: orderID})
	s.Require().NoError(err)

	// 8. FindByTrashed
	trashedRes, err := s.client.FindByTrashed(ctx, &pborder.FindAllOrderRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(trashedRes.Data)

	// 9. Restore
	_, err = s.client.RestoreOrder(ctx, &pborder.FindByIdOrderRequest{Id: orderID})
	s.Require().NoError(err)

	// 10. DeletePermanent
	_, _ = s.client.TrashedOrder(ctx, &pborder.FindByIdOrderRequest{Id: orderID})
	_, err = s.client.DeleteOrderPermanent(ctx, &pborder.FindByIdOrderRequest{Id: orderID})
	s.Require().NoError(err)

	// 11. RestoreAll
	_, err = s.client.RestoreAllOrder(ctx, &emptypb.Empty{})
	s.Require().NoError(err)

	// 12. DeleteAll
	_, err = s.client.DeleteAllOrderPermanent(ctx, &emptypb.Empty{})
	s.Require().NoError(err)
}

func TestOrderGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(OrderGapiTestSuite))
}
