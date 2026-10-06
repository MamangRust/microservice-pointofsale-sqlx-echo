package order_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	mencache "github.com/MamangRust/microservice-point-of-sale-order/cache"
	"github.com/MamangRust/microservice-point-of-sale-order/handler"
	"github.com/MamangRust/microservice-point-of-sale-order/repository"
	"github.com/MamangRust/microservice-point-of-sale-order/service"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	merchantadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/merchant"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	tests "github.com/MamangRust/microservice-point-of-sale-test"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// faultMerchantServer implements the merchant gRPC service with FindById that
// ALWAYS fails with Unavailable (injected transport failure) and counts calls,
// so tests can prove the dependency guard fails fast without hitting the
// dependency once the circuit breaker opens (§8.1 poin 5).
type faultMerchantServer struct {
	pbmerchant.UnimplementedMerchantServiceServer
	calls int32
}

func (f *faultMerchantServer) FindById(ctx context.Context, req *pbmerchant.FindByIdMerchantRequest) (*pbmerchant.ApiResponseMerchant, error) {
	atomic.AddInt32(&f.calls, 1)
	return nil, status.Error(codes.Unavailable, "merchant service unavailable (injected)")
}

func (f *faultMerchantServer) callCount() int {
	return int(atomic.LoadInt32(&f.calls))
}

// OrderFailureInjectionTestSuite exercises the F6 dependency guard end-to-end
// through the REAL merchant repository (guarded) + REAL gRPC handler: an
// injected Unavailable from the merchant dependency must (a) not be preceded by
// remote calls when the request is invalid, (b) trip the circuit breaker after
// the failure threshold, and (c) fail fast afterwards without touching the
// dependency.
type OrderFailureInjectionTestSuite struct {
	tests.BaseTestSuite
	merchantServer *faultMerchantServer
	svc            *service.Service
	client         pborder.OrderServiceClient
	orderID        int // merchant id
	cashierID      int
}

func (s *OrderFailureInjectionTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	// Seed the OLTP entities CreateOrder validates against.
	var userID int
	err := s.SQLxDB().QueryRowContext(s.Ctx,
		`INSERT INTO users (firstname, lastname, email, password, verification_code, is_verified) VALUES ($1, $2, $3, $4, 'test-verify', true) RETURNING user_id`,
		"Fail", "Inject", "fail.inject@example.com", "password123",
	).Scan(&userID)
	s.Require().NoError(err)

	err = s.SQLxDB().QueryRowContext(s.Ctx,
		`INSERT INTO merchants (user_id, name, description, address, contact_email, contact_phone, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING merchant_id`,
		userID, "Failure Inject Merchant", "Desc", "Addr", "fi@example.com", "123", "active",
	).Scan(&s.orderID)
	s.Require().NoError(err)

	err = s.SQLxDB().QueryRowContext(s.Ctx,
		`INSERT INTO cashiers (merchant_id, user_id, name) VALUES ($1, $2, 'Failure Cashier') RETURNING cashier_id`,
		s.orderID, userID,
	).Scan(&s.cashierID)
	s.Require().NoError(err)

	// Fault-injecting merchant dependency: real gRPC server + real client +
	// real guarded repository, so the transport-failure path is authentic.
	s.merchantServer = &faultMerchantServer{}
	merchantGRPC := grpc.NewServer()
	pbmerchant.RegisterMerchantServiceServer(merchantGRPC, s.merchantServer)
	addr, err := tests.RunGRPCServer(merchantGRPC)
	s.Require().NoError(err)
	merchantConn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.Servers = append(s.Servers, merchantGRPC)
	s.Conns["merchant"] = merchantConn

	merchantGuard := resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log)

	// Build the repository set directly so the merchant repo is the REAL guarded
	// gRPC repository while cashier/product/order_item are never reached (the
	// merchant dependency fails first in every scenario below).
	repos := &repository.Repositories{
		CashierQuery: &stubCashierRepo{cashierID: s.cashierID},
		MerchantQuery: merchantadapter.New(
			pbmerchant.NewMerchantServiceClient(merchantConn),
			adapter.WithDependencyGuard(merchantGuard),
		),
		ProductQuery:     &stubProductRepo{},
		ProductCommand:   &stubProductRepo{},
		OrderQuery:       repository.NewOrderQueryRepository(s.SQLxDB()),
		OrderCommand:     repository.NewOrderCommandRepository(s.SQLxDB()),
		OrderItemQuery:   &stubOrderItemRepo{},
		OrderItemCommand: &stubOrderItemRepo{},
	}

	mencacheObj := mencache.NewMencache(s.GetCacheStore())
	s.svc = service.NewService(&service.Deps{
		Repositories:  repos,
		Logger:        s.Log,
		Mencache:      mencacheObj,
		Observability: s.Obs,
	})

	// Real gRPC handler so the validation-before-remote-call test goes through
	// the authentic handler path (req.Validate()).
	orderGapi := handler.NewHandler(&handler.Deps{
		Service: s.svc,
		Logger:  s.Log,
	})
	orderGRPC := grpc.NewServer()
	pborder.RegisterOrderServiceServer(orderGRPC, orderGapi.Order)
	orderAddr, err := tests.RunGRPCServer(orderGRPC)
	s.Require().NoError(err)
	s.Servers = append(s.Servers, orderGRPC)
	orderConn, err := grpc.NewClient(orderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.Conns["order"] = orderConn
	s.client = pborder.NewOrderServiceClient(orderConn)
}

func (s *OrderFailureInjectionTestSuite) TearDownSuite() {
	s.BaseTestSuite.TearDownSuite()
}

// TestF1_InvalidRequest_RejectedWith400_BeforeRemoteCalls verifies that
// validation happens BEFORE any remote dependency call: with the merchant
// dependency down, an invalid request must fail with InvalidArgument and the
// merchant server must not receive a single call.
func (s *OrderFailureInjectionTestSuite) TestF1_InvalidRequest_RejectedWith400_BeforeRemoteCalls() {
	ctx := context.Background()

	_, err := s.client.Create(ctx, &pborder.CreateOrderRequest{
		MerchantId: int32(s.orderID),
		CashierId:  int32(s.cashierID),
		// Items empty -> Validate() fails.
	})

	s.Require().Error(err, "empty items must be rejected")
	s.Equal(codes.InvalidArgument, status.Code(err), "validation failure must map to InvalidArgument (400)")
	s.Zero(s.merchantServer.callCount(), "no remote call may happen before validation")
}

// TestF2_DependencyDown_OpensCircuit_ThenFailsFast verifies the dependency
// guard end-to-end: repeated Unavailable responses from the merchant dependency
// (a) are surfaced as errors (not a crash), (b) open the circuit breaker after
// the 5th failure, and (c) make the next call fail fast WITHOUT touching the
// dependency.
func (s *OrderFailureInjectionTestSuite) TestF2_DependencyDown_OpensCircuit_ThenFailsFast() {
	ctx := context.Background()
	req := &requests.CreateOrderRequest{
		MerchantID: s.orderID,
		CashierID:  s.cashierID,
		Items: []requests.CreateOrderItemRequest{
			{ProductID: 1, Quantity: 1},
		},
	}

	baseline := s.merchantServer.callCount()

	// Threshold = 5: each of these 5 calls reaches the dependency and records a
	// transport failure; the 5th opens the circuit.
	for i := 0; i < 5; i++ {
		_, err := s.svc.OrderCommand.CreateOrder(ctx, req)
		s.Require().Error(err, "dependency down must fail the create (%d)", i+1)
	}
	s.Equal(baseline+5, s.merchantServer.callCount(), "all 5 failures must reach the dependency")

	// The 6th call must fail fast through the open circuit without contacting
	// the merchant service.
	afterOpen := s.merchantServer.callCount()
	_, err := s.svc.OrderCommand.CreateOrder(ctx, req)
	s.Require().Error(err, "open circuit must fail fast with an error")
	s.Equal(afterOpen, s.merchantServer.callCount(), "fail-fast must not reach the dependency")
}

func TestOrderFailureInjectionSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(OrderFailureInjectionTestSuite))
}
