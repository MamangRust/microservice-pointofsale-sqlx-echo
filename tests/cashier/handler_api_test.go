package cashier_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MamangRust/microservice-point-of-sale-apigateway/cache/gateway_cache"
	api_handler "github.com/MamangRust/microservice-point-of-sale-apigateway/handler"
	cashier_cache "github.com/MamangRust/microservice-point-of-sale-cashier/cache"
	cashier_handler "github.com/MamangRust/microservice-point-of-sale-cashier/handler"
	cashier_repo "github.com/MamangRust/microservice-point-of-sale-cashier/repository"
	cashier_service "github.com/MamangRust/microservice-point-of-sale-cashier/service"
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

type CashierApiTestSuite struct {
	tests.BaseTestSuite
	echo      *echo.Echo
	cashierID int
	userID    int
	merchID   int
}

func (s *CashierApiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	s.SetupMerchantService()

	ctx := context.Background()

	// Seed dependencies via gRPC
	s.userID = s.SeedUser(ctx)
	s.merchID = s.SeedMerchant(ctx, s.userID)

	// Setup cashier gRPC service
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	mencache := cashier_cache.NewMencache(cacheStore)
	repos := cashier_repo.NewRepositories(s.SQLxDB(), pbuser.NewUserServiceClient(s.Conns["user"]), pbmerchant.NewMerchantServiceClient(s.Conns["merchant"]))
	svc := cashier_service.NewService(&cashier_service.Deps{
		Ctx:           context.Background(),
		Repositories:  repos,
		Logger:        s.Log,
		Mencache:      mencache,
		Observability: s.Obs,
	})
	cashierGrpcHandler := cashier_handler.NewHandler(&cashier_handler.Deps{
		Service: svc,
		Logger:  s.Log,
	})

	// Start gRPC server for cashier
	grpcServer := grpc.NewServer()
	pbcashier.RegisterCashierServiceServer(grpcServer, cashierGrpcHandler.Cashier)
	addr := s.RegisterServer(grpcServer)
	conn := s.GetConnection(addr)

	// Setup apigateway echo with cashier routes
	s.echo = echo.New()
	apiObs, _ := observability.NewObservability("test", s.Log)
	apiErrHandler := errors.NewApiHandler(apiObs, s.Log)
	gwCache := gateway_cache.NewGatewayCache(cacheStore)

	api_handler.NewHandlerCashier(
		s.echo,
		pbcashier.NewCashierServiceClient(conn),
		s.Log,
		response_api.NewCashierResponseMapper(),
		apiErrHandler,
		gwCache,
	)
}

func (s *CashierApiTestSuite) TestCashierApiLifecycle() {
	// 1. Create
	reqBody := requests.CreateCashierRequest{
		MerchantID: s.merchID,
		UserID:     s.userID,
		Name:       "Api Cashier",
	}
	body, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/cashier/create", bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Require().True(rec.Code == http.StatusOK || rec.Code == http.StatusCreated, rec.Body.String())

	var res map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &res)
	data := res["data"].(map[string]interface{})
	s.cashierID = int(data["id"].(float64))

	// 2. FindById
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/cashier/%d", s.cashierID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 3. FindAll
	req = httptest.NewRequest(http.MethodGet, "/api/cashier", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 4. FindByActive
	req = httptest.NewRequest(http.MethodGet, "/api/cashier/active", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 5. Update
	updateBody := requests.UpdateCashierRequest{
		Name: "Api Cashier Updated",
	}
	body, _ = json.Marshal(updateBody)
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/cashier/update/%d", s.cashierID), bytes.NewBuffer(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 6. Trash
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/cashier/trashed/%d", s.cashierID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 7. FindByTrashed
	req = httptest.NewRequest(http.MethodGet, "/api/cashier/trashed", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 8. Restore
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/cashier/restore/%d", s.cashierID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 9. DeletePermanent (trash first)
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/cashier/trashed/%d", s.cashierID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/cashier/permanent/%d", s.cashierID), nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 10. RestoreAll
	req = httptest.NewRequest(http.MethodPost, "/api/cashier/restore/all", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)

	// 11. DeleteAll
	req = httptest.NewRequest(http.MethodPost, "/api/cashier/permanent/all", nil)
	rec = httptest.NewRecorder()
	s.echo.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
}

func TestCashierApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(CashierApiTestSuite))
}
