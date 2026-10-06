package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	auth_cache_api "github.com/MamangRust/microservice-point-of-sale-apigateway/cache/auth"
	apigateway "github.com/MamangRust/microservice-point-of-sale-apigateway/handler"
	auth_cache "github.com/MamangRust/microservice-point-of-sale-auth/cache"
	"github.com/MamangRust/microservice-point-of-sale-auth/handler"
	"github.com/MamangRust/microservice-point-of-sale-auth/repository"
	"github.com/MamangRust/microservice-point-of-sale-auth/service"
	pbauth "github.com/MamangRust/microservice-point-of-sale-pb/auth"
	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/auth"
	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	role_cache "github.com/MamangRust/microservice-point-of-sale-role/cache"
	role_handler "github.com/MamangRust/microservice-point-of-sale-role/handler"
	role_repo "github.com/MamangRust/microservice-point-of-sale-role/repository"
	role_service "github.com/MamangRust/microservice-point-of-sale-role/service"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	response_api "github.com/MamangRust/microservice-point-of-sale-shared/mapper/response/api"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	user_cache "github.com/MamangRust/microservice-point-of-sale-user/cache"
	user_handler "github.com/MamangRust/microservice-point-of-sale-user/handler"
	user_repo "github.com/MamangRust/microservice-point-of-sale-user/repository"
	user_service "github.com/MamangRust/microservice-point-of-sale-user/service"

	"github.com/jmoiron/sqlx"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthHandlerApiTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	dbPool      *sqlx.DB
	redisClient *redis.Client
	server      *echo.Echo
	email       string
	password    string
	accessToken string
	userID      int
}

func (s *AuthHandlerApiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	pool, err := sqlx.ConnectContext(s.ts.Ctx, "pgx", s.ts.DBURL)
	s.Require().NoError(err)
	s.dbPool = pool

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)
	s.redisClient.FlushAll(context.Background())

	roleQueries := pool
	userQueries := pool
	authQueries := pool

	log, _ := logger.NewLogger("test", nil)
	hasher := hash.NewHashingPassword()
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)
	obs, _ := observability.NewObservability("test", log)

	// 1. Setup Role Service & gRPC Server
	roleMencache := role_cache.NewMencache(cacheStore)
	roleRepos := role_repo.NewRepositories(roleQueries)
	roleSvc := role_service.NewService(&role_service.Deps{
		Repositories:  roleRepos,
		Logger:        log,
		Mencache:      roleMencache,
		Observability: obs,
	})
	roleGapi := role_handler.NewHandler(&role_handler.Deps{
		Service: roleSvc,
		Logger:  log,
	})
	roleServer := grpc.NewServer()
	pbrole.RegisterRoleServiceServer(roleServer, roleGapi.Role)
	pbuserrole.RegisterUserRoleServiceServer(roleServer, roleGapi.UserRole)
	roleLis, _ := net.Listen("tcp", "localhost:0")
	go roleServer.Serve(roleLis)
	roleConn, _ := grpc.NewClient(roleLis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	roleClient := pbrole.NewRoleServiceClient(roleConn)
	userRoleClient := pbuserrole.NewUserRoleServiceClient(roleConn)

	// 2. Setup User Service & gRPC Server
	userMencache := user_cache.NewMencache(cacheStore)
	userRepos := user_repo.NewRepositories(userQueries, roleClient, userRoleClient)
	userSvc := user_service.NewService(&user_service.Deps{
		Repositories:  userRepos,
		Logger:        log,
		Hash:          hasher,
		Mencache:      userMencache,
		Observability: obs,
	})
	userGapi := user_handler.NewHandler(&user_handler.Deps{
		Service: userSvc,
		Logger:  log,
	})
	userServer := grpc.NewServer()
	pbuser.RegisterUserServiceServer(userServer, userGapi.User)
	userLis, _ := net.Listen("tcp", "localhost:0")
	go userServer.Serve(userLis)

	// 3. Setup Auth Service
	repos := repository.NewRepositories(authQueries, roleClient, userRoleClient)

	tokenManager, _ := auth.NewManager("mysecret")
	mencache := auth_cache.NewMencache(cacheStore)
	apiAuthCache := auth_cache_api.NewMencache(cacheStore)

	svc := service.NewService(&service.Deps{
		Repositories:  repos,
		Logger:        log,
		Mencache:      mencache,
		Token:         tokenManager,
		Hash:          hasher,
		Kafka:         nil,
		Observability: obs,
	})

	h := handler.NewAuthHandleGrpc(svc, log)

	grpcServer := grpc.NewServer()
	pbauth.RegisterAuthServiceServer(grpcServer, h)

	lis, err := net.Listen("tcp", "localhost:0")
	s.Require().NoError(err)

	go func() {
		_ = grpcServer.Serve(lis)
	}()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)

	s.server = echo.New()
	apiHandler := errors.NewApiHandler(obs, log)
	mapper := response_api.NewAuthResponseMapper()

	// Auth bypass middleware for /api/auth/me
	s.server.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if s.userID != 0 {
				c.Set("userId", strconv.Itoa(s.userID))
			}
			return next(c)
		}
	})

	fmt.Println("Calling RegisterAuthHandler...")
	apigateway.NewHandlerAuth(s.server, pbauth.NewAuthServiceClient(conn), log, mapper, apiHandler, apiAuthCache)
	fmt.Println("RegisterAuthHandler called.")

	s.server.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	s.email = "auth.handler.api.test@example.com"
	s.password = "password123"

	// ROLE_ADMIN sudah di-seed oleh tests.SetupTestSuite (SeedMinimalRoles).
	// CreateRole di bawah hanya memastikan wiring role service aktif;
	// konflik duplikat (sudah ada) diabaikan.
	if _, err := roleClient.CreateRole(context.Background(), &pbrole.CreateRoleRequest{
		Name: "ROLE_ADMIN",
	}); err != nil {
		fmt.Printf("DEBUG: Seed ROLE_ADMIN gRPC (expected conflict): %v\n", err)
	}
}

func (s *AuthHandlerApiTestSuite) TearDownSuite() {
	if s.redisClient != nil {
		s.redisClient.Close()
	}
	if s.dbPool != nil {
		s.dbPool.Close()
	}
	s.ts.Teardown()
}

func (s *AuthHandlerApiTestSuite) Test0_Ping() {
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	rec := httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
	s.Equal("pong", rec.Body.String())
}

func (s *AuthHandlerApiTestSuite) Test0_Hello() {
	req := httptest.NewRequest(http.MethodGet, "/api/auth/hello", nil)
	rec := httptest.NewRecorder()
	s.server.ServeHTTP(rec, req)
	s.Equal(http.StatusOK, rec.Code)
	s.Equal("Hello", rec.Body.String())
}

// verifyUser simulates the email-verification step (login only accepts
// is_verified = true users).
func (s *AuthHandlerApiTestSuite) verifyUser(email string) {
	_, err := s.dbPool.ExecContext(context.Background(),
		"UPDATE users SET is_verified = true WHERE email = $1", email)
	s.Require().NoError(err)
}

func (s *AuthHandlerApiTestSuite) Test1_Register() {
	body := map[string]string{
		"firstname":        "Auth",
		"lastname":         "API",
		"email":            s.email,
		"password":         s.password,
		"confirm_password": s.password,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	s.server.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code, "Expected StatusOK, got %d. Body: %s", rec.Code, rec.Body.String())

	var res map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	s.NoError(err)

	data, ok := res["data"].(map[string]interface{})
	s.True(ok, "Expected 'data' to be a map, got %T. Body: %s", res["data"], rec.Body.String())
	s.userID = int(data["id"].(float64))
}

func (s *AuthHandlerApiTestSuite) Test2_Login() {
	s.verifyUser(s.email)

	body := map[string]string{
		"email":    s.email,
		"password": s.password,
	}
	jsonBody, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(jsonBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	s.server.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code, "Expected StatusOK, got %d. Body: %s", rec.Code, rec.Body.String())

	var res map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	s.NoError(err)

	data, ok := res["data"].(map[string]interface{})
	s.True(ok, "Expected 'data' to be a map, got %T. Body: %s", res["data"], rec.Body.String())
	s.accessToken = data["access_token"].(string)
}

func (s *AuthHandlerApiTestSuite) Test4_LoginLockout() {
	email := "locked.api@example.com"
	password := "wrongpassword"

	// Register user first
	regBody := map[string]string{
		"firstname":        "Locked",
		"lastname":         "API",
		"email":            email,
		"password":         "correctpassword",
		"confirm_password": "correctpassword",
	}
	jsonRegBody, _ := json.Marshal(regBody)
	regReq := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(jsonRegBody))
	regReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	regRec := httptest.NewRecorder()
	s.server.ServeHTTP(regRec, regReq)
	s.Equal(http.StatusOK, regRec.Code)
	s.verifyUser(email)

	loginBody := map[string]string{
		"email":    email,
		"password": password,
	}
	jsonLoginBody, _ := json.Marshal(loginBody)

	// Fail login 5 times (total 5) — password mismatch is 400 Bad Request
	for i := 0; i < 5; i++ {
		loginReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(jsonLoginBody))
		loginReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		loginRec := httptest.NewRecorder()
		s.server.ServeHTTP(loginRec, loginReq)
		s.Equal(http.StatusBadRequest, loginRec.Code)
	}

	// 6th attempt should return 400 Bad Request (ErrAccountLocked)
	lockedReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(jsonLoginBody))
	lockedReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	lockedRec := httptest.NewRecorder()
	s.server.ServeHTTP(lockedRec, lockedReq)
	s.Equal(http.StatusBadRequest, lockedRec.Code)
}

func (s *AuthHandlerApiTestSuite) Test3_GetMe() {
	s.Require().NotZero(s.userID)
	s.Require().NotEmpty(s.accessToken)

	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+s.accessToken)
	rec := httptest.NewRecorder()

	s.server.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code, "Expected StatusOK, got %d. Body: %s", rec.Code, rec.Body.String())

	var res map[string]interface{}
	err := json.Unmarshal(rec.Body.Bytes(), &res)
	s.NoError(err)

	data, ok := res["data"].(map[string]interface{})
	s.True(ok, "Expected 'data' to be a map, got %T. Body: %s", res["data"], rec.Body.String())
	s.Equal(s.email, data["email"])
}

func TestAuthHandlerApiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(AuthHandlerApiTestSuite))
}
