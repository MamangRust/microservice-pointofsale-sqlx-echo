package apps

import (
	"os"
	"time"

	mencache "github.com/MamangRust/microservice-point-of-sale-auth/cache"
	"github.com/MamangRust/microservice-point-of-sale-auth/handler"
	"github.com/MamangRust/microservice-point-of-sale-auth/repository"
	"github.com/MamangRust/microservice-point-of-sale-auth/service"
	pbauth "github.com/MamangRust/microservice-point-of-sale-pb/auth"
	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	"github.com/MamangRust/microservice-point-of-sale-pkg/auth"
	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-pkg/kafka"
	"github.com/MamangRust/microservice-point-of-sale-pkg/outbox"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-pkg/server"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

func NewServer(cfg *server.Config) (*server.GRPCServer, error) {
	srv, err := server.New(cfg)
	if err != nil {
		return nil, err
	}

	tokenManager, err := auth.NewManager(viper.GetString("SECRET_KEY"))
	if err != nil {
		return nil, err
	}

	roleAddr := viper.GetString("GRPC_ROLE_ADDR")
	if roleAddr == "" {
		roleAddr = "localhost:50052"
	}
	roleConn, err := server.NewGRPCClient(roleAddr)
	if err != nil {
		return nil, err
	}
	go func() {
		<-srv.Ctx.Done()
		roleConn.Close()
	}()

	roleClient := pbrole.NewRoleServiceClient(roleConn)
	userRoleClient := pbuserrole.NewUserRoleServiceClient(roleConn)
	guardRole := resilience.NewDependencyGuard("role", 5, 30, 100, 3*time.Second, srv.Logger)
	guardUserRole := resilience.NewDependencyGuard("user_role", 5, 30, 100, 3*time.Second, srv.Logger)

	repos := repository.NewRepositories(
		srv.DB, roleClient, userRoleClient,
		repository.GuardOptions{
			Role:     []adapter.GuardOption{adapter.WithDependencyGuard(guardRole)},
			UserRole: []adapter.GuardOption{adapter.WithDependencyGuard(guardUserRole)},
		},
	)
	hash := hash.NewHashingPassword()

	var myKafka *kafka.Kafka
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		myKafka = kafka.NewKafka(srv.Logger, []string{brokers})
	}

	mencacheObj := mencache.NewMencache(srv.CacheStore)

	outboxService := outbox.NewOutboxService(srv.DB, myKafka, srv.Logger)

	services := service.NewService(&service.Deps{
		Mencache:      mencacheObj,
		Repositories:  repos,
		Hash:          hash,
		Token:         tokenManager,
		Logger:        srv.Logger,
		Kafka:         myKafka,
		Pool:          srv.DB,
		Outbox:        outboxService,
		Observability: observability.NewTraceLoggerObservability(srv.Logger),
	})

	handlers := handler.NewHandler(&handler.Deps{
		Service: services,
		Logger:  srv.Logger,
	})

	srv.RegisterServices = func(gs *grpc.Server) {
		pbauth.RegisterAuthServiceServer(gs, handlers.Auth)
	}

	// Start the outbox relay so events committed with the business writes are
	// published to Kafka with durable retry and dead-letter semantics.
	go outboxService.Start(srv.Ctx, outbox.OutboxRelayInterval, outbox.OutboxRelayBatchSize)

	return srv, nil
}
