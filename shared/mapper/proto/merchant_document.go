package protomapper

import (
	pbcommon "github.com/MamangRust/microservice-point-of-sale-pb/common"
	pbmerchantdoc "github.com/MamangRust/microservice-point-of-sale-pb/merchant_document"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type merchantDocumentProtoMapper struct{}

func NewMerchantDocumentProtoMapper() *merchantDocumentProtoMapper {
	return &merchantDocumentProtoMapper{}
}

func (m *merchantDocumentProtoMapper) ToProtoResponseMerchantDocument(status string, message string, doc *response.MerchantDocumentResponse) *pbmerchantdoc.ApiResponseMerchantDocument {
	return &pbmerchantdoc.ApiResponseMerchantDocument{
		Status:  status,
		Message: message,
		Data:    m.mapMerchantDocument(doc),
	}
}

func (m *merchantDocumentProtoMapper) ToProtoResponsesMerchantDocument(status string, message string, docs []*response.MerchantDocumentResponse) *pbmerchantdoc.ApiResponsesMerchantDocument {
	return &pbmerchantdoc.ApiResponsesMerchantDocument{
		Status:  status,
		Message: message,
		Data:    m.mapMerchantDocuments(docs),
	}
}

func (m *merchantDocumentProtoMapper) ToProtoResponsePaginationMerchantDocument(pagination *pbcommon.PaginationMeta, status string, message string, docs []*response.MerchantDocumentResponse) *pbmerchantdoc.ApiResponsePaginationMerchantDocument {
	return &pbmerchantdoc.ApiResponsePaginationMerchantDocument{
		Status:     status,
		Message:    message,
		Data:       m.mapMerchantDocuments(docs),
		Pagination: mapPaginationMeta(pagination),
	}
}

func (m *merchantDocumentProtoMapper) ToProtoResponsePaginationMerchantDocumentDeleteAt(pagination *pbcommon.PaginationMeta, status string, message string, docs []*response.MerchantDocumentResponseDeleteAt) *pbmerchantdoc.ApiResponsePaginationMerchantDocumentAt {
	return &pbmerchantdoc.ApiResponsePaginationMerchantDocumentAt{
		Status:     status,
		Message:    message,
		Data:       m.mapMerchantDocumentsDeleteAt(docs),
		Pagination: mapPaginationMeta(pagination),
	}
}

func (m *merchantDocumentProtoMapper) ToProtoResponseMerchantDocumentDelete(status string, message string) *pbmerchantdoc.ApiResponseMerchantDocumentDelete {
	return &pbmerchantdoc.ApiResponseMerchantDocumentDelete{
		Status:  status,
		Message: message,
	}
}

func (m *merchantDocumentProtoMapper) ToProtoResponseMerchantDocumentAll(status string, message string) *pbmerchantdoc.ApiResponseMerchantDocumentAll {
	return &pbmerchantdoc.ApiResponseMerchantDocumentAll{
		Status:  status,
		Message: message,
	}
}

func (m *merchantDocumentProtoMapper) mapMerchantDocument(doc *response.MerchantDocumentResponse) *pbmerchantdoc.MerchantDocument {
	return &pbmerchantdoc.MerchantDocument{
		DocumentId:   int32(doc.ID),
		MerchantId:   int32(doc.MerchantID),
		DocumentType: doc.DocumentType,
		DocumentUrl:  doc.DocumentURL,
		Status:       doc.Status,
		Note:         doc.Note,
		UploadedAt:   doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
	}
}

func (m *merchantDocumentProtoMapper) mapMerchantDocuments(docs []*response.MerchantDocumentResponse) []*pbmerchantdoc.MerchantDocument {
	var res []*pbmerchantdoc.MerchantDocument
	for _, doc := range docs {
		res = append(res, m.mapMerchantDocument(doc))
	}
	return res
}

func (m *merchantDocumentProtoMapper) mapMerchantDocumentDeleteAt(doc *response.MerchantDocumentResponseDeleteAt) *pbmerchantdoc.MerchantDocumentDeleteAt {
	var deletedAt *wrapperspb.StringValue
	if doc.DeletedAt != nil {
		deletedAt = wrapperspb.String(*doc.DeletedAt)
	}

	return &pbmerchantdoc.MerchantDocumentDeleteAt{
		DocumentId:   int32(doc.ID),
		MerchantId:   int32(doc.MerchantID),
		DocumentType: doc.DocumentType,
		DocumentUrl:  doc.DocumentURL,
		Status:       doc.Status,
		Note:         doc.Note,
		UploadedAt:   doc.CreatedAt,
		UpdatedAt:    doc.UpdatedAt,
		DeletedAt:    deletedAt,
	}
}

func (m *merchantDocumentProtoMapper) mapMerchantDocumentsDeleteAt(docs []*response.MerchantDocumentResponseDeleteAt) []*pbmerchantdoc.MerchantDocumentDeleteAt {
	var res []*pbmerchantdoc.MerchantDocumentDeleteAt
	for _, doc := range docs {
		res = append(res, m.mapMerchantDocumentDeleteAt(doc))
	}
	return res
}
