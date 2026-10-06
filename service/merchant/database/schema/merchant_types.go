// Hand-written package db — Row/Params structs migrated from sqlc-generated
// sources (merchant.sql, merchant_document.sql). db tags were added for sqlx
// scanning while json tags are preserved.

package db

import (
	"github.com/jackc/pgx/v5/pgtype"
)

// ----- merchant.sql -----

type CreateMerchantParams struct {
	UserID       int32   `db:"user_id" json:"user_id"`
	Name         string  `db:"name" json:"name"`
	Description  *string `db:"description" json:"description"`
	Address      *string `db:"address" json:"address"`
	ContactEmail *string `db:"contact_email" json:"contact_email"`
	ContactPhone *string `db:"contact_phone" json:"contact_phone"`
	Status       string  `db:"status" json:"status"`
}

type GetMerchantsRow struct {
	MerchantID   int32            `db:"merchant_id" json:"merchant_id"`
	UserID       int32            `db:"user_id" json:"user_id"`
	Name         string           `db:"name" json:"name"`
	Description  *string          `db:"description" json:"description"`
	Address      *string          `db:"address" json:"address"`
	ContactEmail *string          `db:"contact_email" json:"contact_email"`
	ContactPhone *string          `db:"contact_phone" json:"contact_phone"`
	Status       string           `db:"status" json:"status"`
	CreatedAt    pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt    pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount   int64            `db:"total_count" json:"total_count"`
}

type GetMerchantsActiveRow struct {
	MerchantID   int32            `db:"merchant_id" json:"merchant_id"`
	UserID       int32            `db:"user_id" json:"user_id"`
	Name         string           `db:"name" json:"name"`
	Description  *string          `db:"description" json:"description"`
	Address      *string          `db:"address" json:"address"`
	ContactEmail *string          `db:"contact_email" json:"contact_email"`
	ContactPhone *string          `db:"contact_phone" json:"contact_phone"`
	Status       string           `db:"status" json:"status"`
	CreatedAt    pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt    pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount   int64            `db:"total_count" json:"total_count"`
}

type GetMerchantsTrashedRow struct {
	MerchantID   int32            `db:"merchant_id" json:"merchant_id"`
	UserID       int32            `db:"user_id" json:"user_id"`
	Name         string           `db:"name" json:"name"`
	Description  *string          `db:"description" json:"description"`
	Address      *string          `db:"address" json:"address"`
	ContactEmail *string          `db:"contact_email" json:"contact_email"`
	ContactPhone *string          `db:"contact_phone" json:"contact_phone"`
	Status       string           `db:"status" json:"status"`
	CreatedAt    pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt    pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount   int64            `db:"total_count" json:"total_count"`
}

type UpdateMerchantParams struct {
	MerchantID   int32   `db:"merchant_id" json:"merchant_id"`
	Name         string  `db:"name" json:"name"`
	Description  *string `db:"description" json:"description"`
	Address      *string `db:"address" json:"address"`
	ContactEmail *string `db:"contact_email" json:"contact_email"`
	ContactPhone *string `db:"contact_phone" json:"contact_phone"`
	Status       string  `db:"status" json:"status"`
}

type UpdateMerchantStatusParams struct {
	MerchantID int32  `db:"merchant_id" json:"merchant_id"`
	Status     string `db:"status" json:"status"`
}

// ----- merchant_document.sql -----

type CreateMerchantDocumentParams struct {
	MerchantID   int32   `db:"merchant_id" json:"merchant_id"`
	DocumentType string  `db:"document_type" json:"document_type"`
	DocumentUrl  string  `db:"document_url" json:"document_url"`
	Status       string  `db:"status" json:"status"`
	Note         *string `db:"note" json:"note"`
}

type GetMerchantDocumentsRow struct {
	DocumentID   int32            `db:"document_id" json:"document_id"`
	MerchantID   int32            `db:"merchant_id" json:"merchant_id"`
	DocumentType string           `db:"document_type" json:"document_type"`
	DocumentUrl  string           `db:"document_url" json:"document_url"`
	Status       string           `db:"status" json:"status"`
	Note         *string          `db:"note" json:"note"`
	UploadedAt   pgtype.Timestamp `db:"uploaded_at" json:"uploaded_at"`
	CreatedAt    pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt    pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount   int64            `db:"total_count" json:"total_count"`
}

type GetActiveMerchantDocumentsRow struct {
	DocumentID   int32            `db:"document_id" json:"document_id"`
	MerchantID   int32            `db:"merchant_id" json:"merchant_id"`
	DocumentType string           `db:"document_type" json:"document_type"`
	DocumentUrl  string           `db:"document_url" json:"document_url"`
	Status       string           `db:"status" json:"status"`
	Note         *string          `db:"note" json:"note"`
	UploadedAt   pgtype.Timestamp `db:"uploaded_at" json:"uploaded_at"`
	CreatedAt    pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt    pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount   int64            `db:"total_count" json:"total_count"`
}

type GetTrashedMerchantDocumentsRow struct {
	DocumentID   int32            `db:"document_id" json:"document_id"`
	MerchantID   int32            `db:"merchant_id" json:"merchant_id"`
	DocumentType string           `db:"document_type" json:"document_type"`
	DocumentUrl  string           `db:"document_url" json:"document_url"`
	Status       string           `db:"status" json:"status"`
	Note         *string          `db:"note" json:"note"`
	UploadedAt   pgtype.Timestamp `db:"uploaded_at" json:"uploaded_at"`
	CreatedAt    pgtype.Timestamp `db:"created_at" json:"created_at"`
	UpdatedAt    pgtype.Timestamp `db:"updated_at" json:"updated_at"`
	DeletedAt    pgtype.Timestamp `db:"deleted_at" json:"deleted_at"`
	TotalCount   int64            `db:"total_count" json:"total_count"`
}

type UpdateMerchantDocumentParams struct {
	DocumentID   int32   `db:"document_id" json:"document_id"`
	DocumentType string  `db:"document_type" json:"document_type"`
	DocumentUrl  string  `db:"document_url" json:"document_url"`
	Status       string  `db:"status" json:"status"`
	Note         *string `db:"note" json:"note"`
}

type UpdateMerchantDocumentStatusParams struct {
	DocumentID int32   `db:"document_id" json:"document_id"`
	Status     string  `db:"status" json:"status"`
	Note       *string `db:"note" json:"note"`
}
