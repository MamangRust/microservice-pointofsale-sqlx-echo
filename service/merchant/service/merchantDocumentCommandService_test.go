package service

import (
	"context"
	"errors"
	"testing"

	db "github.com/MamangRust/microservice-point-of-sale-merchant/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	merchantdocument_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/merchant_document_errors"
	merchant_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/merchant_errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ---- hand-written stubs / mocks (no external mock framework) ----

type stubLogger struct{}

func (stubLogger) Info(string, ...zap.Field)  {}
func (stubLogger) Fatal(string, ...zap.Field) {}
func (stubLogger) Debug(string, ...zap.Field) {}
func (stubLogger) Error(string, ...zap.Field) {}
func (stubLogger) Warn(string, ...zap.Field)  {}
func (stubLogger) Check(zapcore.Level, string) *zapcore.CheckedEntry {
	return nil
}
func (stubLogger) With(...zap.Field) logger.LoggerInterface { return stubLogger{} }
func (stubLogger) Sync() error                              { return nil }

type mockMerchantDocumentCommandRepository struct {
	updateCalls       []*requests.UpdateMerchantDocumentRequest
	updateStatusCalls []*requests.UpdateMerchantDocumentStatusRequest

	updateResult *db.MerchantDocument
	updateErr    error

	updateStatusResult *db.MerchantDocument
	updateStatusErr    error
}

func (m *mockMerchantDocumentCommandRepository) CreateMerchantDocument(context.Context, *requests.CreateMerchantDocumentRequest) (*db.MerchantDocument, error) {
	return nil, errors.New("not implemented in test")
}

func (m *mockMerchantDocumentCommandRepository) CreateMerchantDocumentInTx(context.Context, *sqlx.Tx, *requests.CreateMerchantDocumentRequest) (*db.MerchantDocument, error) {
	return nil, errors.New("not implemented in test")
}

func (m *mockMerchantDocumentCommandRepository) UpdateMerchantDocument(_ context.Context, request *requests.UpdateMerchantDocumentRequest) (*db.MerchantDocument, error) {
	m.updateCalls = append(m.updateCalls, request)
	return m.updateResult, m.updateErr
}

func (m *mockMerchantDocumentCommandRepository) UpdateMerchantDocumentStatus(_ context.Context, request *requests.UpdateMerchantDocumentStatusRequest) (*db.MerchantDocument, error) {
	m.updateStatusCalls = append(m.updateStatusCalls, request)
	return m.updateStatusResult, m.updateStatusErr
}

func (m *mockMerchantDocumentCommandRepository) UpdateMerchantDocumentStatusInTx(_ context.Context, _ *sqlx.Tx, _ *requests.UpdateMerchantDocumentStatusRequest) (*db.MerchantDocument, error) {
	return nil, errors.New("not implemented in test")
}

func (m *mockMerchantDocumentCommandRepository) TrashedMerchantDocument(context.Context, int) (*db.MerchantDocument, error) {
	return nil, errors.New("not implemented in test")
}

func (m *mockMerchantDocumentCommandRepository) RestoreMerchantDocument(context.Context, int) (*db.MerchantDocument, error) {
	return nil, errors.New("not implemented in test")
}

func (m *mockMerchantDocumentCommandRepository) DeleteMerchantDocumentPermanent(context.Context, int) (bool, error) {
	return false, errors.New("not implemented in test")
}

func (m *mockMerchantDocumentCommandRepository) RestoreAllMerchantDocument(context.Context) (bool, error) {
	return false, errors.New("not implemented in test")
}

func (m *mockMerchantDocumentCommandRepository) DeleteAllMerchantDocumentPermanent(context.Context) (bool, error) {
	return false, errors.New("not implemented in test")
}

type mockMerchantQueryRepository struct {
	merchant *db.Merchant
	err      error
}

func (m *mockMerchantQueryRepository) FindById(context.Context, int) (*db.Merchant, error) {
	return m.merchant, m.err
}

func (m *mockMerchantQueryRepository) FindAllMerchants(context.Context, *requests.FindAllMerchants) ([]*db.GetMerchantsRow, *int, error) {
	return nil, nil, errors.New("not implemented in test")
}

func (m *mockMerchantQueryRepository) FindByActive(context.Context, *requests.FindAllMerchants) ([]*db.GetMerchantsActiveRow, *int, error) {
	return nil, nil, errors.New("not implemented in test")
}

func (m *mockMerchantQueryRepository) FindByTrashed(context.Context, *requests.FindAllMerchants) ([]*db.GetMerchantsTrashedRow, *int, error) {
	return nil, nil, errors.New("not implemented in test")
}

type mockUserQueryRepository struct {
	user *models.User
	err  error
}

func (m *mockUserQueryRepository) FindById(context.Context, int) (*models.User, error) {
	return m.user, m.err
}

func (m *mockUserQueryRepository) FindByEmail(context.Context, string) (*models.User, error) {
	return nil, errors.New("not implemented in test")
}

func (m *mockUserQueryRepository) FindByEmailAndVerify(context.Context, string) (*models.User, error) {
	return nil, errors.New("not implemented in test")
}

func (m *mockUserQueryRepository) FindByVerificationCode(context.Context, string) (*models.User, error) {
	return nil, errors.New("not implemented in test")
}

type mockMerchantDocumentCommandCache struct {
	deletedIDs []int
}

func (m *mockMerchantDocumentCommandCache) DeleteCachedMerchantDocuments(_ context.Context, id int) {
	m.deletedIDs = append(m.deletedIDs, id)
}

func (m *mockMerchantDocumentCommandCache) DeleteCachedMerchantDocumentsAllCache(context.Context) {}

func ptr[T any](v T) *T { return &v }

// newMerchantDocumentCommandService wires the command service with the given
// mocks. Kafka is nil so no broker is touched; observability is the real no-op
// implementation (safe without an OpenTelemetry collector).
func newMerchantDocumentCommandService(
	docCmd *mockMerchantDocumentCommandRepository,
	merchantQuery *mockMerchantQueryRepository,
	userQuery *mockUserQueryRepository,
	cache *mockMerchantDocumentCommandCache,
) MerchantDocumentCommandService {
	return NewMerchantDocumentCommandService(&merchantDocumentCommandDeps{
		Kafka:                   nil,
		Cache:                   cache,
		MerchantQuery:           merchantQuery,
		MerchantDocumentCommand: docCmd,
		UserQuery:               userQuery,
		Logger:                  stubLogger{},
		Observability:           observability.NewTraceLoggerObservability(),
	})
}

// Regression tests for the DocumentID-vs-MerchantID fix. The updated document
// fixture deliberately uses DocumentID=42 / MerchantID=999 so cache
// invalidation keyed on the wrong ID cannot pass by coincidence.

func TestUpdateMerchantDocument_InvalidatesCacheByDocumentID_NotMerchantID(t *testing.T) {
	const (
		requestDocID  = 42
		returnedDocID = 43 // differs from the request ID and the merchant ID
		merchantID    = 999
	)

	docCmd := &mockMerchantDocumentCommandRepository{
		updateResult: &db.MerchantDocument{DocumentID: returnedDocID, MerchantID: merchantID, Status: "verified"},
	}
	cache := &mockMerchantDocumentCommandCache{}
	svc := newMerchantDocumentCommandService(docCmd, nil, nil, cache)

	req := &requests.UpdateMerchantDocumentRequest{
		DocumentID:   ptr(requestDocID),
		MerchantID:   merchantID,
		DocumentType: "license",
		DocumentUrl:  "https://example.com/license.pdf",
		Status:       "verified",
		Note:         "Approved",
	}

	updated, err := svc.UpdateMerchantDocument(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, updated)

	// The request (with its DocumentID) must reach the repository untouched.
	require.Len(t, docCmd.updateCalls, 1)
	assert.Equal(t, req, docCmd.updateCalls[0])
	assert.Equal(t, requestDocID, *docCmd.updateCalls[0].DocumentID)

	// Cache invalidation must be keyed on the *returned* document's ID —
	// neither the request DocumentID nor the merchant ID.
	assert.Equal(t, []int{returnedDocID}, cache.deletedIDs)
	assert.NotContains(t, cache.deletedIDs, requestDocID)
	assert.NotContains(t, cache.deletedIDs, merchantID)
}

func TestUpdateMerchantDocument_RepoError_ReturnsFailedUpdateError(t *testing.T) {
	docCmd := &mockMerchantDocumentCommandRepository{
		updateErr: merchantdocument_errors.ErrUpdateMerchantDocumentFailed,
	}
	cache := &mockMerchantDocumentCommandCache{}
	svc := newMerchantDocumentCommandService(docCmd, nil, nil, cache)

	_, err := svc.UpdateMerchantDocument(context.Background(), &requests.UpdateMerchantDocumentRequest{
		DocumentID:   ptr(42),
		MerchantID:   999,
		DocumentType: "license",
		DocumentUrl:  "https://example.com/license.pdf",
		Status:       "verified",
		Note:         "Approved",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, merchantdocument_errors.ErrUpdateMerchantDocumentFailed)
	assert.Empty(t, cache.deletedIDs, "cache must not be invalidated when the update fails")
}

func TestUpdateMerchantDocumentStatus_InvalidatesCacheByDocumentID_NotMerchantID(t *testing.T) {
	const (
		requestDocID  = 42
		returnedDocID = 43 // differs from the request ID and the merchant ID
		merchantID    = 999
		userID        = 7
	)

	docCmd := &mockMerchantDocumentCommandRepository{
		updateStatusResult: &db.MerchantDocument{DocumentID: returnedDocID, MerchantID: merchantID, Status: "approved"},
	}
	merchantQuery := &mockMerchantQueryRepository{
		merchant: &db.Merchant{MerchantID: merchantID, UserID: userID},
	}
	userQuery := &mockUserQueryRepository{
		user: &models.User{UserID: userID, Email: "owner@example.com"},
	}
	cache := &mockMerchantDocumentCommandCache{}
	svc := newMerchantDocumentCommandService(docCmd, merchantQuery, userQuery, cache)

	req := &requests.UpdateMerchantDocumentStatusRequest{
		DocumentID: ptr(requestDocID),
		MerchantID: merchantID,
		Status:     "approved",
		Note:       "All documents verified",
	}

	updated, err := svc.UpdateMerchantDocumentStatus(context.Background(), req)
	require.NoError(t, err)
	require.NotNil(t, updated)

	require.Len(t, docCmd.updateStatusCalls, 1)
	assert.Equal(t, req, docCmd.updateStatusCalls[0])
	assert.Equal(t, requestDocID, *docCmd.updateStatusCalls[0].DocumentID)

	// Cache invalidation must be keyed on the *returned* document's ID —
	// neither the request DocumentID nor the merchant ID.
	assert.Equal(t, []int{returnedDocID}, cache.deletedIDs)
	assert.NotContains(t, cache.deletedIDs, requestDocID)
	assert.NotContains(t, cache.deletedIDs, merchantID)
}

func TestUpdateMerchantDocumentStatus_UnknownStatus_SkipsEmailButReturnsDocumentAndInvalidatesCache(t *testing.T) {
	const (
		returnedDocID = 43
		merchantID    = 999
	)

	docCmd := &mockMerchantDocumentCommandRepository{
		updateStatusResult: &db.MerchantDocument{DocumentID: returnedDocID, MerchantID: merchantID, Status: "unknown-status"},
	}
	merchantQuery := &mockMerchantQueryRepository{
		merchant: &db.Merchant{MerchantID: merchantID, UserID: 7},
	}
	userQuery := &mockUserQueryRepository{
		user: &models.User{UserID: 7, Email: "owner@example.com"},
	}
	cache := &mockMerchantDocumentCommandCache{}
	svc := newMerchantDocumentCommandService(docCmd, merchantQuery, userQuery, cache)

	req := &requests.UpdateMerchantDocumentStatusRequest{
		DocumentID: ptr(42),
		MerchantID: merchantID,
		Status:     "unknown-status",
		Note:       "note",
	}
	updated, err := svc.UpdateMerchantDocumentStatus(context.Background(), req)

	// Status tanpa template email: update di DB tetap sukses — dokumen
	// dikembalikan dan cache di-invalidasi (bukan return nil yang memicu
	// nil-pointer di response mapper apigateway).
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, int32(returnedDocID), updated.DocumentID)
	require.Len(t, docCmd.updateStatusCalls, 1)
	assert.Equal(t, req, docCmd.updateStatusCalls[0])
	assert.Equal(t, []int{returnedDocID}, cache.deletedIDs)
}

func TestUpdateMerchantDocumentStatus_MerchantNotFound_ReturnsFailedFindMerchant(t *testing.T) {
	merchantQuery := &mockMerchantQueryRepository{
		err: merchant_errors.ErrFailedFindMerchantById,
	}
	docCmd := &mockMerchantDocumentCommandRepository{}
	userQuery := &mockUserQueryRepository{}
	cache := &mockMerchantDocumentCommandCache{}
	svc := newMerchantDocumentCommandService(docCmd, merchantQuery, userQuery, cache)

	_, err := svc.UpdateMerchantDocumentStatus(context.Background(), &requests.UpdateMerchantDocumentStatusRequest{
		DocumentID: ptr(42),
		MerchantID: 999,
		Status:     "approved",
		Note:       "ok",
	})
	require.Error(t, err)
	assert.ErrorIs(t, err, merchant_errors.ErrFailedFindMerchantById)
	assert.Empty(t, docCmd.updateStatusCalls)
	assert.Empty(t, cache.deletedIDs)
}
