package repository

import (
	"context"

	db "github.com/MamangRust/microservice-point-of-sale-merchant/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"github.com/jmoiron/sqlx"
)

type merchantDocumentQueryRepository struct {
	db *sqlx.DB
}

func NewMerchantDocumentQueryRepository(db *sqlx.DB) MerchantDocumentQueryRepository {
	return &merchantDocumentQueryRepository{
		db: db,
	}
}

const getMerchantDocuments = `SELECT *, COUNT(*) OVER() AS total_count
FROM merchant_documents
WHERE deleted_at IS NULL
  AND (
    $1::TEXT IS NULL OR
    document_type ILIKE '%' || $1 || '%' OR
    status ILIKE '%' || $1 || '%' OR
    note ILIKE '%' || $1 || '%'
  )
ORDER BY document_id
LIMIT $2 OFFSET $3`

const getActiveMerchantDocuments = `SELECT *, COUNT(*) OVER() AS total_count
FROM merchant_documents
WHERE deleted_at IS NULL AND status != 'deleted'
  AND (
    $1::TEXT IS NULL OR
    document_type ILIKE '%' || $1 || '%' OR
    status ILIKE '%' || $1 || '%' OR
    note ILIKE '%' || $1 || '%'
  )
ORDER BY document_id
LIMIT $2 OFFSET $3`

const getTrashedMerchantDocuments = `SELECT *, COUNT(*) OVER() AS total_count
FROM merchant_documents
WHERE deleted_at IS NOT NULL
  AND (
    $1::TEXT IS NULL OR
    document_type ILIKE '%' || $1 || '%' OR
    status ILIKE '%' || $1 || '%' OR
    note ILIKE '%' || $1 || '%'
  )
ORDER BY document_id
LIMIT $2 OFFSET $3`

const getMerchantDocument = `SELECT *
FROM merchant_documents
WHERE document_id = $1 AND deleted_at IS NULL`

func (r *merchantDocumentQueryRepository) FindAllDocuments(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*db.GetMerchantDocumentsRow, *int, error) {
	offset := (req.Page - 1) * req.PageSize

	var documents []*db.GetMerchantDocumentsRow
	err := r.db.SelectContext(ctx, &documents, getMerchantDocuments,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	totalCount := 0
	if len(documents) > 0 {
		totalCount = int(documents[0].TotalCount)
	}

	return documents, &totalCount, nil
}

func (r *merchantDocumentQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*db.GetActiveMerchantDocumentsRow, *int, error) {
	offset := (req.Page - 1) * req.PageSize

	var documents []*db.GetActiveMerchantDocumentsRow
	err := r.db.SelectContext(ctx, &documents, getActiveMerchantDocuments,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	totalCount := 0
	if len(documents) > 0 {
		totalCount = int(documents[0].TotalCount)
	}

	return documents, &totalCount, nil
}

func (r *merchantDocumentQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllMerchantDocuments) ([]*db.GetTrashedMerchantDocumentsRow, *int, error) {
	offset := (req.Page - 1) * req.PageSize

	var documents []*db.GetTrashedMerchantDocumentsRow
	err := r.db.SelectContext(ctx, &documents, getTrashedMerchantDocuments,
		req.Search,
		int32(req.PageSize),
		int32(offset),
	)
	if err != nil {
		return nil, nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	totalCount := 0
	if len(documents) > 0 {
		totalCount = int(documents[0].TotalCount)
	}

	return documents, &totalCount, nil
}

func (r *merchantDocumentQueryRepository) FindById(ctx context.Context, id int) (*db.MerchantDocument, error) {
	var document db.MerchantDocument
	if err := r.db.GetContext(ctx, &document, getMerchantDocument, int32(id)); err != nil {
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	return &document, nil
}
