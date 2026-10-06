package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-email/config"
	"github.com/MamangRust/microservice-point-of-sale-email/handler"
	"github.com/MamangRust/microservice-point-of-sale-email/mailer"
	"github.com/MamangRust/microservice-point-of-sale-email/metrics"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/dotenv"
	"github.com/MamangRust/microservice-point-of-sale-pkg/emailretry"
	"github.com/MamangRust/microservice-point-of-sale-pkg/kafka"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	otel_pkg "github.com/MamangRust/microservice-point-of-sale-pkg/otel"
	"github.com/MamangRust/microservice-point-of-sale-pkg/outbox"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	if err := dotenv.Viper(); err != nil {
		log.Fatalf("Failed to load .env file: %v", err)
	}

	endpoint := viper.GetString("OTEL_ENDPOINT")
	if endpoint == "" {
		endpoint = "otel-collector:4317"
	}

	// Telemetry must be initialized before the logger: the logger attaches an
	// OTLP log bridge to the LoggerProvider (otelzap).
	telemetry := otel_pkg.NewTelemetry(otel_pkg.Config{
		ServiceName:            "email-service",
		ServiceVersion:         "1.0.0",
		Environment:            viper.GetString("ENV"),
		Endpoint:               endpoint,
		Insecure:               true,
		EnableRuntimeMetrics:   true,
		RuntimeMetricsInterval: 15 * time.Second,
		Disabled:               os.Getenv("OTEL_ENABLED") == "false",
	})

	if err := telemetry.Init(ctx); err != nil {
		log.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer func() {
		if err := telemetry.Shutdown(ctx); err != nil {
			log.Printf("Error shutting down telemetry: %v", err)
		}
	}()

	logger, err := logger.NewLogger("email-service", telemetry.GetLogger())
	if err != nil {
		log.Fatalf("Error creating logger: %v", err)
	}

	cfg := config.Config{
		KafkaBrokers: []string{viper.GetString("KAFKA_BROKERS")},
		SMTPServer:   viper.GetString("SMTP_SERVER"),
		SMTPPort:     viper.GetInt("SMTP_PORT"),
		SMTPUser:     viper.GetString("SMTP_USER"),
		SMTPPass:     viper.GetString("SMTP_PASS"),
		MaxRetries:   viper.GetInt("EMAIL_MAX_RETRIES"),
		RetryBackoff: viper.GetDuration("EMAIL_RETRY_BACKOFF"),
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = emailretry.DefaultMaxAttempts
	}
	if cfg.RetryBackoff <= 0 {
		cfg.RetryBackoff = emailretry.DefaultBackoff
	}

	// Register OTel metric instruments after the SDK is initialized so they
	// are bound to the real meter provider (exported via OTLP), not the noop
	// default from package init.
	metrics.Register()

	m := &mailer.Mailer{
		Server:   cfg.SMTPServer,
		Port:     cfg.SMTPPort,
		User:     cfg.SMTPUser,
		Password: cfg.SMTPPass,
	}

	// Phase 3 (durable idempotency): the consumer inbox lives in PostgreSQL, so
	// the email service now requires a database connection at startup. If the
	// database is unreachable the service refuses to start rather than silently
	// losing the idempotency guarantee.
	dbPool, err := database.NewClientWithPrefix(logger, database.EmailCluster)
	if err != nil {
		logger.Fatal("Failed to connect to database for consumer inbox", zap.Error(err))
	}
	defer dbPool.Close()

	inbox, err := outbox.NewPostgresInbox(dbPool)
	if err != nil {
		logger.Fatal("Failed to initialize consumer inbox", zap.Error(err))
	}
	myKafka := kafka.NewKafka(logger, cfg.KafkaBrokers)

	h := handler.NewEmailHandlerWithInbox(m, inbox, "email-service-group", myKafka, cfg.RetryBackoff)

	err = myKafka.StartConsumersWithContext(ctx, []string{
		"email-service-topic-auth-register",
		"email-service-topic-auth-forgot-password",
		"email-service-topic-auth-verify-code-success",
		"email-service-topic-merchant-create",
		"email-service-topic-merchant-update-status",
		"email-service-topic-merchant-document-create",
		"email-service-topic-merchant-document-update-status",
		"email-service-topic-transaction-create",
	}, "email-service-group", h)

	if err != nil {
		log.Fatalf("Error starting consumer: %v", err)
	}

	// Phase 4: retry processor — drains the shared retry topic with ordered
	// backoff, re-attempts SMTP, and escalates to the DLQ after max attempts.
	retryH := handler.NewRetryHandler(m, inbox, "email-service-group", myKafka, cfg.MaxRetries, cfg.RetryBackoff)
	if err := myKafka.StartConsumersWithContext(ctx, []string{emailretry.RetryTopic}, emailretry.RetryGroup, retryH); err != nil {
		log.Fatalf("Error starting retry consumer: %v", err)
	}

	logger.Info("Email service started", zap.String("retry_topic", emailretry.RetryTopic), zap.String("dlq_topic", emailretry.DLQTopic))

	// Phase 5: graceful shutdown — consumers stop on ctx cancellation, then the
	// producer and the DB pool are closed before the process exits.
	<-ctx.Done()
	logger.Info("Shutting down email service")
	if err := myKafka.Close(); err != nil {
		logger.Error("Failed to close Kafka resources", zap.Error(err))
	}
}
