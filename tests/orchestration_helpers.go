package tests

import (
	"bytes"
	"context"
	"mime/multipart"
	"time"

	pbauth "github.com/MamangRust/microservice-point-of-sale-pb/auth"
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	"github.com/MamangRust/microservice-point-of-sale-pkg/auth"
	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	"github.com/jmoiron/sqlx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	// Role
	role_cache "github.com/MamangRust/microservice-point-of-sale-role/cache"
	role_handler "github.com/MamangRust/microservice-point-of-sale-role/handler"
	role_repo "github.com/MamangRust/microservice-point-of-sale-role/repository"
	role_service "github.com/MamangRust/microservice-point-of-sale-role/service"

	// User
	user_cache "github.com/MamangRust/microservice-point-of-sale-user/cache"
	user_handler "github.com/MamangRust/microservice-point-of-sale-user/handler"
	user_repo "github.com/MamangRust/microservice-point-of-sale-user/repository"
	user_service "github.com/MamangRust/microservice-point-of-sale-user/service"

	// Auth
	auth_cache "github.com/MamangRust/microservice-point-of-sale-auth/cache"
	auth_handler "github.com/MamangRust/microservice-point-of-sale-auth/handler"
	auth_repo "github.com/MamangRust/microservice-point-of-sale-auth/repository"
	auth_service "github.com/MamangRust/microservice-point-of-sale-auth/service"

	// Category
	category_cache "github.com/MamangRust/microservice-point-of-sale-category/cache"
	category_handler "github.com/MamangRust/microservice-point-of-sale-category/handler"
	category_repo "github.com/MamangRust/microservice-point-of-sale-category/repository"
	category_service "github.com/MamangRust/microservice-point-of-sale-category/service"

	// Product
	product_cache "github.com/MamangRust/microservice-point-of-sale-product/cache"
	product_handler "github.com/MamangRust/microservice-point-of-sale-product/handler"
	product_repo "github.com/MamangRust/microservice-point-of-sale-product/repository"
	product_service "github.com/MamangRust/microservice-point-of-sale-product/service"

	// Merchant
	merchant_cache "github.com/MamangRust/microservice-point-of-sale-merchant/cache"
	merchant_handler "github.com/MamangRust/microservice-point-of-sale-merchant/handler"
	merchant_repo "github.com/MamangRust/microservice-point-of-sale-merchant/repository"
	merchant_service "github.com/MamangRust/microservice-point-of-sale-merchant/service"

	// Order
	order_cache "github.com/MamangRust/microservice-point-of-sale-order/cache"
	order_handler "github.com/MamangRust/microservice-point-of-sale-order/handler"
	order_repo "github.com/MamangRust/microservice-point-of-sale-order/repository"
	order_service "github.com/MamangRust/microservice-point-of-sale-order/service"

	// Transaction
	transaction_cache "github.com/MamangRust/microservice-point-of-sale-transacton/cache"
	transaction_handler "github.com/MamangRust/microservice-point-of-sale-transacton/handler"
	transaction_repo "github.com/MamangRust/microservice-point-of-sale-transacton/repository"
	transaction_service "github.com/MamangRust/microservice-point-of-sale-transacton/service"

	// Order Item
	order_item_cache "github.com/MamangRust/microservice-point-of-sale-order-item/cache"
	order_item_handler "github.com/MamangRust/microservice-point-of-sale-order-item/handler"
	order_item_repo "github.com/MamangRust/microservice-point-of-sale-order-item/repository"
	order_item_service "github.com/MamangRust/microservice-point-of-sale-order-item/service"

	// Cashier
	mencache "github.com/MamangRust/microservice-point-of-sale-cashier/cache"
	cashier_handler "github.com/MamangRust/microservice-point-of-sale-cashier/handler"
	cashier_repo "github.com/MamangRust/microservice-point-of-sale-cashier/repository"
	cashier_service "github.com/MamangRust/microservice-point-of-sale-cashier/service"
)

// NewRoleServiceClients starts an in-process role service (RoleService +
// UserRoleService) backed by pool and returns clients for both, plus the server
// so the caller can stop it on teardown. It exists for suites that build their
// own repositories instead of embedding BaseTestSuite.
func NewRoleServiceClients(
	pool *sqlx.DB,
	cacheStore *cache.CacheStore,
	log logger.LoggerInterface,
	obs observability.TraceLoggerObservability,
) (pbrole.RoleServiceClient, pbuserrole.UserRoleServiceClient, *grpc.Server, error) {
	roleRepos := role_repo.NewRepositories(pool)
	roleSvc := role_service.NewService(&role_service.Deps{
		Repositories:  roleRepos,
		Logger:        log,
		Mencache:      role_cache.NewMencache(cacheStore),
		Observability: obs,
	})
	roleGapi := role_handler.NewHandler(&role_handler.Deps{
		Service: roleSvc,
		Logger:  log,
	})

	server := grpc.NewServer()
	pbrole.RegisterRoleServiceServer(server, roleGapi.Role)
	pbuserrole.RegisterUserRoleServiceServer(server, roleGapi.UserRole)

	addr, err := RunGRPCServer(server)
	if err != nil {
		return nil, nil, nil, err
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, nil, err
	}

	return pbrole.NewRoleServiceClient(conn), pbuserrole.NewUserRoleServiceClient(conn), server, nil
}

func (s *BaseTestSuite) SetupRoleService() {
	cacheStore := s.GetCacheStore()
	roleMencache := role_cache.NewMencache(cacheStore)
	roleRepos := role_repo.NewRepositories(s.ts.SQLxDB())
	roleSvc := role_service.NewService(&role_service.Deps{
		Repositories:  roleRepos,
		Logger:        s.Log,
		Mencache:      roleMencache,
		Observability: s.Obs,
	})
	roleGapi := role_handler.NewHandler(&role_handler.Deps{
		Service: roleSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pbrole.RegisterRoleServiceServer(server, roleGapi.Role)
	pbuserrole.RegisterUserRoleServiceServer(server, roleGapi.UserRole)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.Conns["role"] = conn
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupUserService() {
	if _, ok := s.Conns["role"]; !ok {
		s.SetupRoleService()
	}

	cacheStore := s.GetCacheStore()
	hasher := hash.NewHashingPassword()

	userMencache := user_cache.NewMencache(cacheStore)
	guardRole := resilience.NewDependencyGuard("role", 5, 30, 100, 3*time.Second, s.Log)
	guardUserRole := resilience.NewDependencyGuard("user_role", 5, 30, 100, 3*time.Second, s.Log)
	userRepos := user_repo.NewRepositories(
		s.ts.SQLxDB(),
		pbrole.NewRoleServiceClient(s.Conns["role"]),
		pbuserrole.NewUserRoleServiceClient(s.Conns["role"]),
		user_repo.GuardOptions{
			Role:     []adapter.GuardOption{adapter.WithDependencyGuard(guardRole)},
			UserRole: []adapter.GuardOption{adapter.WithDependencyGuard(guardUserRole)},
		},
	)
	userSvc := user_service.NewService(&user_service.Deps{
		Repositories:  userRepos,
		Logger:        s.Log,
		Hash:          hasher,
		Mencache:      userMencache,
		Observability: s.Obs,
	})
	userGapi := user_handler.NewHandler(&user_handler.Deps{
		Service: userSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pbuser.RegisterUserServiceServer(server, userGapi.User)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.Conns["user"] = conn
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupAuthService() {
	if _, ok := s.Conns["role"]; !ok {
		s.SetupRoleService()
	}

	cacheStore := s.GetCacheStore()
	hasher := hash.NewHashingPassword()
	tokenManager, _ := auth.NewManager("mysecret")

	guardRole := resilience.NewDependencyGuard("role", 5, 30, 100, 3*time.Second, s.Log)
	guardUserRole := resilience.NewDependencyGuard("user_role", 5, 30, 100, 3*time.Second, s.Log)
	authRepos := auth_repo.NewRepositories(
		s.ts.SQLxDB(),
		pbrole.NewRoleServiceClient(s.Conns["role"]),
		pbuserrole.NewUserRoleServiceClient(s.Conns["role"]),
		auth_repo.GuardOptions{
			Role:     []adapter.GuardOption{adapter.WithDependencyGuard(guardRole)},
			UserRole: []adapter.GuardOption{adapter.WithDependencyGuard(guardUserRole)},
		},
	)
	authMencache := auth_cache.NewMencache(cacheStore)
	authSvc := auth_service.NewService(&auth_service.Deps{
		Repositories:  authRepos,
		Logger:        s.Log,
		Mencache:      authMencache,
		Token:         tokenManager,
		Hash:          hasher,
		Kafka:         nil,
		Observability: s.Obs,
	})
	authGapi := auth_handler.NewAuthHandleGrpc(authSvc, s.Log)
	server := grpc.NewServer()
	pbauth.RegisterAuthServiceServer(server, authGapi)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.Conns["auth"] = conn
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupCategoryService() {
	cacheStore := s.GetCacheStore()
	catMencache := category_cache.NewMencache(cacheStore)
	catRepos := category_repo.NewRepositories(s.ts.SQLxDB())
	catSvc := category_service.NewService(&category_service.Deps{
		Mencache:      catMencache,
		Repositories:  catRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	catGapi := category_handler.NewHandler(&category_handler.Deps{
		Service: catSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pbcategory.RegisterCategoryServiceServer(server, catGapi.Category)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["category"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupProductService() {
	if _, ok := s.Conns["category"]; !ok {
		s.SetupCategoryService()
	}
	if _, ok := s.Conns["merchant"]; !ok {
		s.SetupMerchantService()
	}

	cacheStore := s.GetCacheStore()
	prodMencache := product_cache.NewMencache(cacheStore)
	categoryClient := pbcategory.NewCategoryServiceClient(s.Conns["category"])
	merchantClient := pbmerchant.NewMerchantServiceClient(s.Conns["merchant"])
	guardCategory := resilience.NewDependencyGuard("category", 5, 30, 100, 3*time.Second, s.Log)
	guardMerchant := resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log)
	prodRepos := product_repo.NewRepositories(
		s.ts.SQLxDB(), categoryClient, merchantClient,
		product_repo.GuardOptions{
			Category: []adapter.GuardOption{adapter.WithDependencyGuard(guardCategory)},
			Merchant: []adapter.GuardOption{adapter.WithDependencyGuard(guardMerchant)},
		},
	)
	prodSvc := product_service.NewService(&product_service.Deps{
		Mencache:      prodMencache,
		Repositories:  prodRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	prodGapi := product_handler.NewHandler(&product_handler.Deps{
		Service: prodSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pbproduct.RegisterProductServiceServer(server, prodGapi.Product)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["product"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupMerchantService() {
	if _, ok := s.Conns["user"]; !ok {
		s.SetupUserService()
	}

	cacheStore := s.GetCacheStore()
	merchantMencache := merchant_cache.NewMencache(cacheStore)
	userQueryClient := pbuser.NewUserServiceClient(s.Conns["user"])
	guardUser := resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, s.Log)
	merchantRepos := merchant_repo.NewRepositories(
		s.ts.SQLxDB(), userQueryClient,
		merchant_repo.GuardOptions{
			User: []adapter.GuardOption{adapter.WithDependencyGuard(guardUser)},
		},
	)
	merchantSvc := merchant_service.NewService(&merchant_service.Deps{
		Mencache:      merchantMencache,
		Repositories:  merchantRepos,
		Logger:        s.Log,
		Observability: s.Obs,
		Kafka:         nil,
	})
	merchantGapi := merchant_handler.NewHandler(&merchant_handler.Deps{
		Service: merchantSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pbmerchant.RegisterMerchantServiceServer(server, merchantGapi.Merchant)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["merchant"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupOrderService() {
	if _, ok := s.Conns["cashier"]; !ok {
		s.SetupCashierService()
	}
	if _, ok := s.Conns["merchant"]; !ok {
		s.SetupMerchantService()
	}
	if _, ok := s.Conns["product"]; !ok {
		s.SetupProductService()
	}
	if _, ok := s.Conns["order-item"]; !ok {
		s.SetupOrderItemService()
	}

	cacheStore := s.GetCacheStore()
	cashierClient := pbcashier.NewCashierServiceClient(s.Conns["cashier"])
	merchantClient := pbmerchant.NewMerchantServiceClient(s.Conns["merchant"])
	productClient := pbproduct.NewProductServiceClient(s.Conns["product"])
	orderItemClient := pborderitem.NewOrderItemServiceClient(s.Conns["order-item"])

	orderMencache := order_cache.NewMencache(cacheStore)
	guardCashier := resilience.NewDependencyGuard("cashier", 5, 30, 100, 3*time.Second, s.Log)
	guardMerchant := resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log)
	guardProduct := resilience.NewDependencyGuard("product", 5, 30, 100, 3*time.Second, s.Log)
	guardOrderItem := resilience.NewDependencyGuard("order_item", 5, 30, 100, 3*time.Second, s.Log)
	orderRepos := order_repo.NewRepositories(
		s.ts.SQLxDB(), cashierClient, merchantClient, productClient, orderItemClient,
		order_repo.GuardOptions{
			Cashier:   []adapter.GuardOption{adapter.WithDependencyGuard(guardCashier)},
			Merchant:  []adapter.GuardOption{adapter.WithDependencyGuard(guardMerchant)},
			Product:   []adapter.GuardOption{adapter.WithDependencyGuard(guardProduct)},
			OrderItem: []adapter.GuardOption{adapter.WithDependencyGuard(guardOrderItem)},
		},
	)
	orderSvc := order_service.NewService(&order_service.Deps{
		Mencache:      orderMencache,
		Repositories:  orderRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	orderGapi := order_handler.NewHandler(&order_handler.Deps{
		Service: orderSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pborder.RegisterOrderServiceServer(server, orderGapi.Order)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["order"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupTransactionService() {
	if _, ok := s.Conns["cashier"]; !ok {
		s.SetupCashierService()
	}
	if _, ok := s.Conns["merchant"]; !ok {
		s.SetupMerchantService()
	}
	if _, ok := s.Conns["order"]; !ok {
		s.SetupOrderService()
	}
	if _, ok := s.Conns["order-item"]; !ok {
		s.SetupOrderItemService()
	}

	cacheStore := s.GetCacheStore()
	cashierClient := pbcashier.NewCashierServiceClient(s.Conns["cashier"])
	merchantClient := pbmerchant.NewMerchantServiceClient(s.Conns["merchant"])
	orderClient := pborder.NewOrderServiceClient(s.Conns["order"])
	orderItemClient := pborderitem.NewOrderItemServiceClient(s.Conns["order-item"])

	transactionMencache := transaction_cache.NewMencache(cacheStore)
	guardCashier := resilience.NewDependencyGuard("cashier", 5, 30, 100, 3*time.Second, s.Log)
	guardMerchant := resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log)
	guardOrder := resilience.NewDependencyGuard("order", 5, 30, 100, 3*time.Second, s.Log)
	guardOrderItem := resilience.NewDependencyGuard("order_item", 5, 30, 100, 3*time.Second, s.Log)
	transactionRepos := transaction_repo.NewRepositories(
		s.ts.SQLxDB(), cashierClient, merchantClient, orderClient, orderItemClient,
		transaction_repo.GuardOptions{
			Cashier:   []adapter.GuardOption{adapter.WithDependencyGuard(guardCashier)},
			Merchant:  []adapter.GuardOption{adapter.WithDependencyGuard(guardMerchant)},
			Order:     []adapter.GuardOption{adapter.WithDependencyGuard(guardOrder)},
			OrderItem: []adapter.GuardOption{adapter.WithDependencyGuard(guardOrderItem)},
		},
	)
	transactionSvc := transaction_service.NewService(&transaction_service.Deps{
		Mencache:      transactionMencache,
		Repositories:  transactionRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	transactionGapi := transaction_handler.NewHandler(&transaction_handler.Deps{
		Service: transactionSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pbtransaction.RegisterTransactionServiceServer(server, transactionGapi.Transaction)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["transaction"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupOrderItemService() {
	cacheStore := s.GetCacheStore()
	itemMencache := order_item_cache.NewMencache(cacheStore)
	itemRepos := order_item_repo.NewRepositories(s.ts.SQLxDB())
	itemSvc := order_item_service.NewService(&order_item_service.Deps{
		Mencache:      itemMencache,
		Repositories:  itemRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	itemGapi := order_item_handler.NewHandler(&order_item_handler.Deps{
		Service: itemSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pborderitem.RegisterOrderItemServiceServer(server, itemGapi.OrderItem)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["order-item"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) SetupCashierService() {
	if _, ok := s.Conns["user"]; !ok {
		s.SetupUserService()
	}
	if _, ok := s.Conns["merchant"]; !ok {
		s.SetupMerchantService()
	}

	cacheStore := s.GetCacheStore()
	cashierMencache := mencache.NewMencache(cacheStore)
	userClient := pbuser.NewUserServiceClient(s.Conns["user"])
	merchantClient := pbmerchant.NewMerchantServiceClient(s.Conns["merchant"])
	guardUser := resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, s.Log)
	guardMerchant := resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log)
	cashierRepos := cashier_repo.NewRepositories(
		s.ts.SQLxDB(), userClient, merchantClient,
		cashier_repo.GuardOptions{
			User:     []adapter.GuardOption{adapter.WithDependencyGuard(guardUser)},
			Merchant: []adapter.GuardOption{adapter.WithDependencyGuard(guardMerchant)},
		},
	)
	cashierSvc := cashier_service.NewService(&cashier_service.Deps{
		Ctx:           context.Background(),
		Mencache:      cashierMencache,
		Repositories:  cashierRepos,
		Logger:        s.Log,
		Observability: s.Obs,
	})
	cashierGapi := cashier_handler.NewHandler(&cashier_handler.Deps{
		Service: cashierSvc,
		Logger:  s.Log,
	})
	server := grpc.NewServer()
	pbcashier.RegisterCashierServiceServer(server, cashierGapi.Cashier)
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Conns["cashier"] = s.dial(addr)
	s.Servers = append(s.Servers, server)
}

func (s *BaseTestSuite) dial(addr string) *grpc.ClientConn {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	return conn
}

func (s *BaseTestSuite) GetCacheStore() *cache.CacheStore {
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	return cache.NewCacheStore(s.ts.RedisClient(), s.Log, cacheMetrics)
}

func (s *BaseTestSuite) BuildMultipartRequestBody(fields map[string]string, fieldName, fileName string) ([]byte, string) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	for key, r := range fields {
		fw, _ := w.CreateFormField(key)
		fw.Write([]byte(r))
	}
	fw, _ := w.CreateFormFile(fieldName, fileName)
	fw.Write([]byte("dummy image content"))
	w.Close()
	return b.Bytes(), w.FormDataContentType()
}
