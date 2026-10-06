package response_api

import (
	pbmerchantdoc "github.com/MamangRust/microservice-point-of-sale-pb/merchant_document"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
)

type merchantDocumentResponse struct{}

func NewMerchantDocumentResponseMapper() *merchantDocumentResponse {
	return &merchantDocumentResponse{}
}

func (m *merchantDocumentResponse) ToApiResponseMerchantDocument(doc *pbmerchantdoc.ApiResponseMerchantDocument) *response.ApiResponseMerchantDocument {
	return &response.ApiResponseMerchantDocument{
		Status:  doc.Status,
		Message: doc.Message,
		Data:    m.mapMerchantDocument(doc.Data),
	}
}

func (m *merchantDocumentResponse) ToApiResponsesMerchantDocument(docs *pbmerchantdoc.ApiResponsesMerchantDocument) *response.ApiResponsesMerchantDocument {
	return &response.ApiResponsesMerchantDocument{
		Status:  docs.Status,
		Message: docs.Message,
		Data:    m.mapMerchantDocuments(docs.Data),
	}
}

func (m *merchantDocumentResponse) ToApiResponsePaginationMerchantDocument(docs *pbmerchantdoc.ApiResponsePaginationMerchantDocument) *response.ApiResponsePaginationMerchantDocument {
	return &response.ApiResponsePaginationMerchantDocument{
		Status:     docs.Status,
		Message:    docs.Message,
		Data:       m.mapMerchantDocuments(docs.Data),
		Pagination: mapPaginationMeta(docs.Pagination),
	}
}

func (m *merchantDocumentResponse) ToApiResponsePaginationMerchantDocumentDeleteAt(docs *pbmerchantdoc.ApiResponsePaginationMerchantDocumentAt) *response.ApiResponsePaginationMerchantDocumentDeleteAt {
	return &response.ApiResponsePaginationMerchantDocumentDeleteAt{
		Status:     docs.Status,
		Message:    docs.Message,
		Data:       m.mapMerchantDocumentsDeletedAt(docs.Data),
		Pagination: mapPaginationMeta(docs.Pagination),
	}
}

func (m *merchantDocumentResponse) ToApiResponseMerchantDocumentAll(resp *pbmerchantdoc.ApiResponseMerchantDocumentAll) *response.ApiResponseMerchantDocumentAll {
	return &response.ApiResponseMerchantDocumentAll{
		Status:  resp.Status,
		Message: resp.Message,
	}
}

func (m *merchantDocumentResponse) ToApiResponseMerchantDocumentDeleteAt(resp *pbmerchantdoc.ApiResponseMerchantDocumentDelete) *response.ApiResponseMerchantDocumentDelete {
	return &response.ApiResponseMerchantDocumentDelete{
		Status:  resp.Status,
		Message: resp.Message,
	}
}

func (m *merchantDocumentResponse) mapMerchantDocument(doc *pbmerchantdoc.MerchantDocument) *response.MerchantDocumentResponse {
	if doc == nil {
		return nil
	}
	return &response.MerchantDocumentResponse{
		ID:           int(doc.DocumentId),
		MerchantID:   int(doc.MerchantId),
		DocumentType: doc.DocumentType,
		DocumentURL:  doc.DocumentUrl,
		Status:       doc.Status,
		Note:         doc.Note,
		CreatedAt:    doc.UploadedAt,
		UpdatedAt:    doc.UpdatedAt,
	}
}

func (m *merchantDocumentResponse) mapMerchantDocuments(docs []*pbmerchantdoc.MerchantDocument) []*response.MerchantDocumentResponse {
	var responses []*response.MerchantDocumentResponse
	for _, doc := range docs {
		responses = append(responses, m.mapMerchantDocument(doc))
	}
	return responses
}

func (m *merchantDocumentResponse) mapMerchantDocumentDeletedAt(doc *pbmerchantdoc.MerchantDocumentDeleteAt) *response.MerchantDocumentResponseDeleteAt {
	if doc == nil {
		return nil
	}
	var deletedAt *string

	if doc.DeletedAt != nil {
		deletedAt = &doc.DeletedAt.Value
	}

	return &response.MerchantDocumentResponseDeleteAt{
		ID:           int(doc.DocumentId),
		MerchantID:   int(doc.MerchantId),
		DocumentType: doc.DocumentType,
		DocumentURL:  doc.DocumentUrl,
		Status:       doc.Status,
		Note:         doc.Note,
		CreatedAt:    doc.UploadedAt,
		UpdatedAt:    doc.UpdatedAt,
		DeletedAt:    deletedAt,
	}
}

func (m *merchantDocumentResponse) mapMerchantDocumentsDeletedAt(docs []*pbmerchantdoc.MerchantDocumentDeleteAt) []*response.MerchantDocumentResponseDeleteAt {
	var responses []*response.MerchantDocumentResponseDeleteAt
	for _, doc := range docs {
		responses = append(responses, m.mapMerchantDocumentDeletedAt(doc))
	}
	return responses
}
