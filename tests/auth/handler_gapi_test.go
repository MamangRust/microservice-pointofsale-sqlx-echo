package auth_test

import (
	"context"
	"net"
	"testing"
	"time"

	mencache "github.com/MamangRust/microservice-point-of-sale-auth/cache"
	"github.com/MamangRust/microservice-point-of-sale-auth/handler"
	"github.com/MamangRust/microservice-point-of-sale-auth/repository"
	"github.com/MamangRust/microservice-point-of-sale-auth/service"
	pbauth "github.com/MamangRust/microservice-point-of-sale-pb/auth"
	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	"github.com/MamangRust/microservice-point-of-sale-pkg/auth"
	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"

	role_cache "github.com/MamangRust/microservice-point-of-sale-role/cache"
	role_handler "github.com/MamangRust/microservice-point-of-sale-role/handler"
	role_repo "github.com/MamangRust/microservice-point-of-sale-role/repository"
	role_service "github.com/MamangRust/microservice-point-of-sale-role/service"
	user_cache "github.com/MamangRust/microservice-point-of-sale-user/cache"
	user_handler "github.com/MamangRust/microservice-point-of-sale-user/handler"
	user_repo "github.com/MamangRust/microservice-point-of-sale-user/repository"
	user_service "github.com/MamangRust/microservice-point-of-sale-user/service"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthHandlerGapiTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	dbPool      *sqlx.DB
	redisClient *redis.Client
	client      pbauth.AuthServiceClient
	conn        *grpc.ClientConn
	grpcServer  *grpc.Server
	email       string
	password    string
	accessToken string
}

func (s *AuthHandlerGapiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.Require().NoError(err)
	s.ts = ts

	pool, err := sqlx.ConnectContext(s.ts.Ctx, "pgx", s.ts.DBURL)
	s.Require().NoError(err)
	s.dbPool = pool

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

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
	roleConn, err := grpc.NewClient(roleLis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)

	// 2. Setup User Service & gRPC Server
	userMencache := user_cache.NewMencache(cacheStore)
	userGuardRole := resilience.NewDependencyGuard("role", 5, 30, 100, 3*time.Second, log)
	userGuardUserRole := resilience.NewDependencyGuard("user_role", 5, 30, 100, 3*time.Second, log)
	userRepos := user_repo.NewRepositories(
		userQueries,
		pbrole.NewRoleServiceClient(roleConn),
		pbuserrole.NewUserRoleServiceClient(roleConn),
		user_repo.GuardOptions{
			Role:     []adapter.GuardOption{adapter.WithDependencyGuard(userGuardRole)},
			UserRole: []adapter.GuardOption{adapter.WithDependencyGuard(userGuardUserRole)},
		},
	)
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
	authGuardRole := resilience.NewDependencyGuard("role", 5, 30, 100, 3*time.Second, log)
	authGuardUserRole := resilience.NewDependencyGuard("user_role", 5, 30, 100, 3*time.Second, log)
	repos := repository.NewRepositories(
		authQueries,
		pbrole.NewRoleServiceClient(roleConn),
		pbuserrole.NewUserRoleServiceClient(roleConn),
		repository.GuardOptions{
			Role:     []adapter.GuardOption{adapter.WithDependencyGuard(authGuardRole)},
			UserRole: []adapter.GuardOption{adapter.WithDependencyGuard(authGuardUserRole)},
		},
	)

	tokenManager, _ := auth.NewManager("mysecret")
	svc := service.NewService(&service.Deps{
		Repositories:  repos,
		Logger:        log,
		Mencache:      mencache.NewMencache(cacheStore),
		Token:         tokenManager,
		Hash:          hasher,
		Kafka:         nil,
		Observability: obs,
	})

	h := handler.NewAuthHandleGrpc(svc, log)

	s.grpcServer = grpc.NewServer()
	pbauth.RegisterAuthServiceServer(s.grpcServer, h)

	lis, err := net.Listen("tcp", "localhost:0")
	s.Require().NoError(err)

	go func() {
		_ = s.grpcServer.Serve(lis)
	}()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn
	s.client = pbauth.NewAuthServiceClient(conn)

	s.email = "auth.handler.gapi.test@example.com"
	s.password = "password123"

	// Seed ROLE_ADMIN
	_, _ = pool.ExecContext(context.Background(), "INSERT INTO roles (role_name) VALUES ('ROLE_ADMIN')")
}

func (s *AuthHandlerGapiTestSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	if s.redisClient != nil {
		s.redisClient.Close()
	}
	if s.dbPool != nil {
		s.dbPool.Close()
	}
	s.ts.Teardown()
}

// verifyUser simulates the email-verification step (login only accepts
// is_verified = true users).
func (s *AuthHandlerGapiTestSuite) verifyUser(email string) {
	_, err := s.dbPool.ExecContext(context.Background(),
		"UPDATE users SET is_verified = true WHERE email = $1", email)
	s.Require().NoError(err)
}

func (s *AuthHandlerGapiTestSuite) Test1_Register() {
	ctx := context.Background()
	req := &pbauth.RegisterRequest{
		Firstname:       "Auth",
		Lastname:        "Handler",
		Email:           s.email,
		Password:        s.password,
		ConfirmPassword: s.password,
	}

	res, err := s.client.RegisterUser(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal("success", res.Status)
	s.Equal(s.email, res.Data.Email)
}

func (s *AuthHandlerGapiTestSuite) Test2_Login() {
	ctx := context.Background()

	s.verifyUser(s.email)

	req := &pbauth.LoginRequest{
		Email:    s.email,
		Password: s.password,
	}

	res, err := s.client.LoginUser(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal("success", res.Status)
	s.NotEmpty(res.Data.AccessToken)
	s.accessToken = res.Data.AccessToken
}

func (s *AuthHandlerGapiTestSuite) Test4_LoginLockout() {
	ctx := context.Background()
	email := "locked.gapi@example.com"
	password := "wrongpassword"

	// Register user first
	regReq := &pbauth.RegisterRequest{
		Firstname:       "Locked",
		Lastname:        "Gapi",
		Email:           email,
		Password:        "correctpassword",
		ConfirmPassword: "correctpassword",
	}
	_, err := s.client.RegisterUser(ctx, regReq)
	s.NoError(err)
	s.verifyUser(email)

	loginReq := &pbauth.LoginRequest{
		Email:    email,
		Password: password,
	}

	// Fail login 5 times
	for i := 0; i < 5; i++ {
		_, err := s.client.LoginUser(ctx, loginReq)
		s.Error(err)
	}

	// 6th attempt should return error
	_, err = s.client.LoginUser(ctx, loginReq)
	s.Error(err)
	s.Contains(err.Error(), "Account is locked")
}

func (s *AuthHandlerGapiTestSuite) Test3_GetMe() {
	s.Require().NotEmpty(s.accessToken)
	ctx := context.Background()

	res, err := s.client.GetMe(ctx, &pbauth.GetMeRequest{AccessToken: s.accessToken})
	s.NoError(err)
	s.NotNil(res)
	s.Equal("success", res.Status)
	s.Equal(s.email, res.Data.Email)
}

func TestAuthHandlerGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(AuthHandlerGapiTestSuite))
}
