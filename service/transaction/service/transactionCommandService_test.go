package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	db "github.com/MamangRust/microservice-point-of-sale-transacton/database/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// fakePublisher records SendMessage calls; err can be set to simulate a
// Kafka publish failure.
type fakePublisher struct {
	calls int
	topic string
	key   string
	value []byte
	err   error
}

func (f *fakePublisher) SendMessage(_ context.Context, topic, key string, value []byte) error {
	f.calls++
	f.topic = topic
	f.key = key
	f.value = value
	return f.err
}

// fakeLogger records Warn calls for graceful-degradation assertions.
type fakeLogger struct {
	warns []string
}

func (l *fakeLogger) Info(message string, fields ...zap.Field)  {}
func (l *fakeLogger) Fatal(message string, fields ...zap.Field) {}
func (l *fakeLogger) Debug(message string, fields ...zap.Field) {}
func (l *fakeLogger) Error(message string, fields ...zap.Field) {}
func (l *fakeLogger) Warn(message string, fields ...zap.Field)  { l.warns = append(l.warns, message) }
func (l *fakeLogger) Check(zapcore.Level, string) *zapcore.CheckedEntry {
	return nil
}
func (l *fakeLogger) With(...zap.Field) logger.LoggerInterface { return l }
func (l *fakeLogger) Sync() error                              { return nil }

func newTestCommandService(publisher EmailEventPublisher, log *fakeLogger) *transactionCommandService {
	// Only kafka, pool, outbox, and logger are exercised by the event paths;
	// all other dependencies can be nil.
	return NewTransactionCommandService(publisher, nil, nil, nil, nil, nil, nil, nil, nil, nil, log, nil)
}

func strPtr(s string) *string { return &s }

func TestSendTransactionCreateEvent_KafkaNotConfigured(t *testing.T) {
	log := &fakeLogger{}
	s := newTestCommandService(nil, log)

	s.sendTransactionCreateEvent(
		context.Background(),
		&models.Merchant{MerchantID: 7, ContactEmail: strPtr("merchant@example.com")},
		&db.Transaction{TransactionID: 42, MerchantID: 7},
	)

	require.Len(t, log.warns, 1)
	assert.Contains(t, log.warns[0], "Kafka not configured")
}

func TestSendTransactionCreateEvent_EmptyContactEmail(t *testing.T) {
	for _, tc := range []struct {
		name     string
		email    *string
		expected string
	}{
		{name: "nil contact email", email: nil, expected: "Merchant contact email not available"},
		{name: "empty contact email", email: strPtr(""), expected: "Merchant contact email not available"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			publisher := &fakePublisher{}
			log := &fakeLogger{}
			s := newTestCommandService(publisher, log)

			s.sendTransactionCreateEvent(
				context.Background(),
				&models.Merchant{MerchantID: 7, ContactEmail: tc.email},
				&db.Transaction{TransactionID: 42, MerchantID: 7},
			)

			assert.Zero(t, publisher.calls, "no message should be published")
			require.Len(t, log.warns, 1)
			assert.Contains(t, log.warns[0], tc.expected)
		})
	}
}

func TestSendTransactionCreateEvent_Success(t *testing.T) {
	publisher := &fakePublisher{}
	log := &fakeLogger{}
	s := newTestCommandService(publisher, log)

	s.sendTransactionCreateEvent(
		context.Background(),
		&models.Merchant{MerchantID: 7, ContactEmail: strPtr("merchant@example.com")},
		&db.Transaction{TransactionID: 42, MerchantID: 7, Amount: 11000},
	)

	assert.Equal(t, 1, publisher.calls)
	assert.Equal(t, "email-service-topic-transaction-create", publisher.topic)
	assert.Equal(t, "42", publisher.key)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(publisher.value, &payload))
	assert.Equal(t, "merchant@example.com", payload["email"])
	assert.NotEmpty(t, payload["subject"])
	assert.NotEmpty(t, payload["body"])
	// Phase 2 (event contract): producer must publish the standard envelope.
	assert.NotEmpty(t, payload["event_id"], "event_id is required")
	assert.Equal(t, float64(1), payload["schema_version"])
	assert.Equal(t, "transaction.created", payload["event_type"])
	assert.NotEmpty(t, payload["occurred_at"])
	assert.Len(t, log.warns, 0, "no warning expected on success")
}

func TestSendTransactionCreateEvent_PublishError(t *testing.T) {
	publisher := &fakePublisher{err: errors.New("kafka broker unreachable")}
	log := &fakeLogger{}
	s := newTestCommandService(publisher, log)

	// Must not panic and must not propagate the error (graceful degradation).
	s.sendTransactionCreateEvent(
		context.Background(),
		&models.Merchant{MerchantID: 7, ContactEmail: strPtr("merchant@example.com")},
		&db.Transaction{TransactionID: 42, MerchantID: 7},
	)

	assert.Equal(t, 1, publisher.calls)
	require.Len(t, log.warns, 1)
	assert.Contains(t, log.warns[0], "failed to send transaction email via kafka")
}
