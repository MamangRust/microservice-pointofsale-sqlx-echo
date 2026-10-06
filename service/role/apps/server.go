package apps

import (
	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/server"
	mencache "github.com/MamangRust/microservice-point-of-sale-role/cache"
	"github.com/MamangRust/microservice-point-of-sale-role/handler"
	"github.com/MamangRust/microservice-point-of-sale-role/repository"
	"github.com/MamangRust/microservice-point-of-sale-role/service"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	"google.golang.org/grpc"
)

func NewServer(cfg *server.Config) (*server.GRPCServer, error) {
	srv, err := server.New(cfg)
	if err != nil {
		return nil, err
	}

	repos := repository.NewRepositories(srv.DB)
	traceLoggerObservability := observability.NewTraceLoggerObservability(srv.Logger)

	cacheMetrics, err := observability.NewCacheMetrics("role")
	if err != nil {
		return nil, err
	}

	cacheStore := cache.NewCacheStore(srv.Redis, srv.Logger, cacheMetrics)
	mencacheObj := mencache.NewMencache(cacheStore)

	services := service.NewService(&service.Deps{
		Mencache:      mencacheObj,
		Repositories:  repos,
		Logger:        srv.Logger,
		Observability: traceLoggerObservability,
	})

	handlers := handler.NewHandler(&handler.Deps{
		Service: services,
		Logger:  srv.Logger,
	})

	srv.RegisterServices = func(gs *grpc.Server) {
		pbrole.RegisterRoleServiceServer(gs, handlers.Role)
		pbuserrole.RegisterUserRoleServiceServer(gs, handlers.UserRole)
	}

	return srv, nil
}
