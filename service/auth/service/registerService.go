package service

import (
	"context"
	"strconv"
	"strings"
	"time"

	mencache "github.com/MamangRust/microservice-point-of-sale-auth/cache"
	db "github.com/MamangRust/microservice-point-of-sale-auth/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-auth/repository"
	"github.com/MamangRust/microservice-point-of-sale-pkg/email"
	"github.com/MamangRust/microservice-point-of-sale-pkg/event"
	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-pkg/kafka"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-pkg/outbox"
	"github.com/MamangRust/microservice-point-of-sale-pkg/randomstring"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	sharederrorhandler "github.com/MamangRust/microservice-point-of-sale-shared/errorhandler"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/user_errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/viper"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type RegisterServiceDeps struct {
	Cache mencache.RegisterCache

	User repository.UserRepository

	Role repository.RoleRepository

	UserRole repository.UserRoleRepository

	Hash hash.HashPassword

	Kafka *kafka.Kafka

	Pool *sqlx.DB

	Outbox *outbox.OutboxService

	Logger logger.LoggerInterface

	Observability observability.TraceLoggerObservability
}

type registerService struct {
	mencache mencache.RegisterCache

	user repository.UserRepository

	role repository.RoleRepository

	userRole repository.UserRoleRepository

	hash hash.HashPassword

	kafka *kafka.Kafka

	pool *sqlx.DB

	outbox *outbox.OutboxService

	logger logger.LoggerInterface

	observability observability.TraceLoggerObservability
}

func NewRegisterService(params *RegisterServiceDeps) *registerService {
	return &registerService{
		mencache:      params.Cache,
		user:          params.User,
		role:          params.Role,
		userRole:      params.UserRole,
		hash:          params.Hash,
		kafka:         params.Kafka,
		pool:          params.Pool,
		outbox:        params.Outbox,
		logger:        params.Logger,
		observability: params.Observability,
	}
}

func (s *registerService) Register(ctx context.Context, request *requests.RegisterRequest) (*db.User, error) {
	const method = "Register"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method, attribute.String("email", request.Email))

	defer func() {
		end(status)
	}()

	existingUser, err := s.user.FindByEmail(ctx, request.Email)
	if err == nil && existingUser != nil {
		status = "error"
		return sharederrorhandler.HandleError[*db.User](
			s.logger,
			user_errors.ErrUserEmailAlready,
			method,
			span,
			zap.String("email", request.Email),
		)
	}

	passwordHash, err := s.hash.HashPassword(request.Password)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*db.User](s.logger, err, method, span)
	}
	request.Password = passwordHash

	const defaultRoleName = "ROLE_ADMIN"
	role, err := s.role.FindByName(ctx, defaultRoleName)
	if err != nil || role == nil {
		status = "error"
		return sharederrorhandler.HandleError[*db.User](s.logger, err, method, span, zap.String("role_name", defaultRoleName))
	}

	random, err := randomstring.GenerateRandomString(10)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*db.User](s.logger, err, method, span)
	}
	request.VerifiedCode = random
	// Auto-verify in non-production environments so the full register → verify →
	// login journey is testable locally; production keeps email verification.
	switch strings.ToLower(viper.GetString("APP_ENV")) {
	case "production", "kubernetes":
		request.IsVerified = false
	default:
		request.IsVerified = true
	}

	htmlBody := email.GenerateEmailHTML(map[string]string{
		"Title":   "Welcome to SanEdge",
		"Message": "Your account has been successfully created.",
		"Button":  "Verify Now",
		"Link":    "https://sanedge.example.com/login?verify_code=" + request.VerifiedCode,
	})

	payloadBytes, err := event.MarshalEmail("auth.register", request.Email, "Welcome to SanEdge", htmlBody)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*db.User](s.logger, err, method, span)
	}

	// Phase 6 — transactional outbox: user insert, role assignment, and the
	// outbox event commit in a single database transaction, so a crash between
	// them cannot lose the registration email. The relay then publishes with
	// retry/dead-letter semantics. Without a pool (tests/local) this falls back
	// to the previous direct-Kafka behavior.
	var newUser *db.User
	if s.pool != nil {
		tx, beginErr := s.pool.Beginx()
		if beginErr != nil {
			status = "error"
			return sharederrorhandler.HandleError[*db.User](s.logger, beginErr, method, span)
		}
		defer func() { _ = tx.Rollback() }()

		newUser, err = s.user.CreateUserInTx(ctx, tx, request)
		if err == nil && s.outbox != nil {
			err = s.outbox.EnqueueInTx(ctx, tx, "email-service-topic-auth-register", strconv.Itoa(int(newUser.UserID)), payloadBytes)
		}
		if err != nil {
			status = "error"
			return sharederrorhandler.HandleError[*db.User](s.logger, err, method, span, zap.Int("user.id", int(newUser.UserID)))
		}
		if err := tx.Commit(); err != nil {
			status = "error"
			return sharederrorhandler.HandleError[*db.User](s.logger, err, method, span, zap.Int("user.id", int(newUser.UserID)))
		}

		// Role assignment now goes through the user_role service, so it cannot
		// join the local transaction above; it runs right after the commit.
		if _, err := s.userRole.AssignRoleToUser(ctx, &requests.CreateUserRoleRequest{
			UserId: int(newUser.UserID),
			RoleId: int(role.RoleID),
		}); err != nil {
			status = "error"
			return sharederrorhandler.HandleError[*db.User](s.logger, err, method, span, zap.Int("user.id", int(newUser.UserID)))
		}
	} else {
		// Fallback: direct writes without a transaction (event loss possible
		// between commit and publish). Intended for tests/local only.
		newUser, err = s.user.CreateUser(ctx, request)
		if err != nil {
			status = "error"
			return sharederrorhandler.HandleError[*db.User](s.logger, err, method, span)
		}

		_, err = s.userRole.AssignRoleToUser(ctx, &requests.CreateUserRoleRequest{
			UserId: int(newUser.UserID),
			RoleId: int(role.RoleID),
		})
		if err != nil {
			status = "error"
			return sharederrorhandler.HandleError[*db.User](s.logger, err, method, span, zap.Int("user.id", int(newUser.UserID)))
		}

		if s.kafka != nil {
			if sendErr := s.kafka.SendMessage(ctx, "email-service-topic-auth-register", strconv.Itoa(int(newUser.UserID)), payloadBytes); sendErr != nil {
				s.logger.Error("failed to send registration email via kafka", zap.Error(sendErr), zap.String("email", request.Email))
			}
		}
	}

	s.mencache.SetVerificationCodeCache(ctx, request.Email, random, 15*time.Minute)

	logSuccess("User registered successfully",
		zap.String("email", request.Email),
		zap.String("first_name", request.FirstName),
		zap.String("last_name", request.LastName),
	)

	return newUser, nil
}
