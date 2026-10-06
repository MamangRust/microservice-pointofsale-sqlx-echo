package handler

import (
	"context"
	"math"

	db "github.com/MamangRust/microservice-point-of-sale-merchant/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-merchant/service"
	pbcommon "github.com/MamangRust/microservice-point-of-sale-pb/common"
	pbmerchantdoc "github.com/MamangRust/microservice-point-of-sale-pb/merchant_document"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	merchantdocument_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/merchant_document_errors"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type merchantDocumentHandleGrpc struct {
	pbmerchantdoc.UnimplementedMerchantDocumentServiceServer
	merchantDocumentQuery   service.MerchantDocumentQueryService
	merchantDocumentCommand service.MerchantDocumentCommandService
	logger                  logger.LoggerInterface
}

func NewMerchantDocumentHandleGrpc(
	service *service.Service,
	logger logger.LoggerInterface,
) pbmerchantdoc.MerchantDocumentServiceServer {
	return &merchantDocumentHandleGrpc{
		merchantDocumentQuery:   service.MerchantDocumentQuery,
		merchantDocumentCommand: service.MerchantDocumentCommand,
		logger:                  logger,
	}
}

func (s *merchantDocumentHandleGrpc) FindAll(ctx context.Context, req *pbmerchantdoc.FindAllMerchantDocumentsRequest) (*pbmerchantdoc.ApiResponsePaginationMerchantDocument, error) {
	s.logger.Info("FindAll merchant documents called", zap.Int32("page", req.GetPage()))

	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllMerchantDocuments{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	documents, totalRecords, err := s.merchantDocumentQuery.FindAll(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindAll merchant documents failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindAll merchant documents success")

	return &pbmerchantdoc.ApiResponsePaginationMerchantDocument{
		Status:     "success",
		Message:    "Successfully fetched merchant documents",
		Data:       mapResponsesGetMerchantDocumentsRow(documents),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *merchantDocumentHandleGrpc) FindById(ctx context.Context, req *pbmerchantdoc.FindMerchantDocumentByIdRequest) (*pbmerchantdoc.ApiResponseMerchantDocument, error) {
	s.logger.Info("FindById merchant document called", zap.Int32("id", req.GetDocumentId()))

	id := int(req.GetDocumentId())
	if id <= 0 {
		return nil, merchantdocument_errors.ErrGrpcMerchantInvalidID
	}

	document, err := s.merchantDocumentQuery.FindById(ctx, id)
	if err != nil {
		s.logger.Error("FindById merchant document failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("FindById merchant document success")

	return &pbmerchantdoc.ApiResponseMerchantDocument{
		Status:  "success",
		Message: "Successfully fetched merchant document",
		Data:    mapMerchantDocument(document),
	}, nil
}

func (s *merchantDocumentHandleGrpc) FindAllActive(ctx context.Context, req *pbmerchantdoc.FindAllMerchantDocumentsRequest) (*pbmerchantdoc.ApiResponsePaginationMerchantDocument, error) {
	s.logger.Info("FindAllActive merchant documents called", zap.Int32("page", req.GetPage()))

	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllMerchantDocuments{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	documents, totalRecords, err := s.merchantDocumentQuery.FindByActive(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindAllActive merchant documents failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindAllActive merchant documents success")

	return &pbmerchantdoc.ApiResponsePaginationMerchantDocument{
		Status:     "success",
		Message:    "Successfully fetched active merchant documents",
		Data:       mapResponsesGetActiveMerchantDocumentsRow(documents),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *merchantDocumentHandleGrpc) FindAllTrashed(ctx context.Context, req *pbmerchantdoc.FindAllMerchantDocumentsRequest) (*pbmerchantdoc.ApiResponsePaginationMerchantDocumentAt, error) {
	s.logger.Info("FindAllTrashed merchant documents called")

	page := int(req.GetPage())
	pageSize := int(req.GetPageSize())
	search := req.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllMerchantDocuments{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	documents, totalRecords, err := s.merchantDocumentQuery.FindByTrashed(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindAllTrashed merchant documents failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindAllTrashed merchant documents success")

	return &pbmerchantdoc.ApiResponsePaginationMerchantDocumentAt{
		Status:     "success",
		Message:    "Successfully fetched trashed merchant documents",
		Data:       mapResponsesGetTrashedMerchantDocumentsRow(documents),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *merchantDocumentHandleGrpc) Create(ctx context.Context, req *pbmerchantdoc.CreateMerchantDocumentRequest) (*pbmerchantdoc.ApiResponseMerchantDocument, error) {
	s.logger.Info("Create merchant document called", zap.Int32("merchantId", req.GetMerchantId()))

	request := requests.CreateMerchantDocumentRequest{
		MerchantID:   int(req.GetMerchantId()),
		DocumentType: req.GetDocumentType(),
		DocumentUrl:  req.GetDocumentUrl(),
	}

	if err := request.Validate(); err != nil {
		return nil, merchantdocument_errors.ErrGrpcValidateCreateMerchantDocument
	}

	document, err := s.merchantDocumentCommand.CreateMerchantDocument(ctx, &request)
	if err != nil {
		s.logger.Error("Create merchant document failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("Create merchant document success")

	return &pbmerchantdoc.ApiResponseMerchantDocument{
		Status:  "success",
		Message: "Successfully created merchant document",
		Data:    mapMerchantDocument(document),
	}, nil
}

func (s *merchantDocumentHandleGrpc) Update(ctx context.Context, req *pbmerchantdoc.UpdateMerchantDocumentRequest) (*pbmerchantdoc.ApiResponseMerchantDocument, error) {
	s.logger.Info("Update merchant document called", zap.Int32("id", req.GetDocumentId()))

	id := int(req.GetDocumentId())
	if id <= 0 {
		return nil, merchantdocument_errors.ErrGrpcMerchantInvalidID
	}

	request := requests.UpdateMerchantDocumentRequest{
		DocumentID:   &id,
		MerchantID:   int(req.GetMerchantId()),
		DocumentType: req.GetDocumentType(),
		DocumentUrl:  req.GetDocumentUrl(),
		Status:       req.GetStatus(),
		Note:         req.GetNote(),
	}

	if err := request.Validate(); err != nil {
		return nil, merchantdocument_errors.ErrGrpcFailedUpdateMerchantDocument
	}

	document, err := s.merchantDocumentCommand.UpdateMerchantDocument(ctx, &request)
	if err != nil {
		s.logger.Error("Update merchant document failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("Update merchant document success")

	return &pbmerchantdoc.ApiResponseMerchantDocument{
		Status:  "success",
		Message: "Successfully updated merchant document",
		Data:    mapMerchantDocument(document),
	}, nil
}

func (s *merchantDocumentHandleGrpc) UpdateStatus(ctx context.Context, req *pbmerchantdoc.UpdateMerchantDocumentStatusRequest) (*pbmerchantdoc.ApiResponseMerchantDocument, error) {
	s.logger.Info("UpdateStatus merchant document called", zap.Int32("id", req.GetDocumentId()))

	id := int(req.GetDocumentId())
	if id <= 0 {
		return nil, merchantdocument_errors.ErrGrpcMerchantInvalidID
	}

	request := requests.UpdateMerchantDocumentStatusRequest{
		DocumentID: &id,
		MerchantID: int(req.GetMerchantId()),
		Status:     req.GetStatus(),
		Note:       req.GetNote(),
	}

	if err := request.Validate(); err != nil {
		return nil, merchantdocument_errors.ErrGrpcFailedUpdateMerchantDocument
	}

	document, err := s.merchantDocumentCommand.UpdateMerchantDocumentStatus(ctx, &request)
	if err != nil {
		s.logger.Error("UpdateStatus merchant document failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("UpdateStatus merchant document success")

	return &pbmerchantdoc.ApiResponseMerchantDocument{
		Status:  "success",
		Message: "Successfully updated merchant document status",
		Data:    mapMerchantDocument(document),
	}, nil
}

func (s *merchantDocumentHandleGrpc) Trashed(ctx context.Context, req *pbmerchantdoc.TrashedMerchantDocumentRequest) (*pbmerchantdoc.ApiResponseMerchantDocument, error) {
	s.logger.Info("Trashed merchant document called", zap.Int32("id", req.GetDocumentId()))

	id := int(req.GetDocumentId())
	if id <= 0 {
		return nil, merchantdocument_errors.ErrGrpcMerchantInvalidID
	}

	document, err := s.merchantDocumentCommand.TrashedMerchantDocument(ctx, id)
	if err != nil {
		s.logger.Error("Trashed merchant document failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("Trashed merchant document success")

	return &pbmerchantdoc.ApiResponseMerchantDocument{
		Status:  "success",
		Message: "Successfully trashed merchant document",
		Data:    mapMerchantDocument(document),
	}, nil
}

func (s *merchantDocumentHandleGrpc) Restore(ctx context.Context, req *pbmerchantdoc.RestoreMerchantDocumentRequest) (*pbmerchantdoc.ApiResponseMerchantDocument, error) {
	s.logger.Info("Restore merchant document called", zap.Int32("id", req.GetDocumentId()))

	id := int(req.GetDocumentId())
	if id <= 0 {
		return nil, merchantdocument_errors.ErrGrpcMerchantInvalidID
	}

	document, err := s.merchantDocumentCommand.RestoreMerchantDocument(ctx, id)
	if err != nil {
		s.logger.Error("Restore merchant document failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("Restore merchant document success")

	return &pbmerchantdoc.ApiResponseMerchantDocument{
		Status:  "success",
		Message: "Successfully restored merchant document",
		Data:    mapMerchantDocument(document),
	}, nil
}

func (s *merchantDocumentHandleGrpc) DeletePermanent(ctx context.Context, req *pbmerchantdoc.DeleteMerchantDocumentPermanentRequest) (*pbmerchantdoc.ApiResponseMerchantDocumentDelete, error) {
	s.logger.Info("DeletePermanent merchant document called", zap.Int32("id", req.GetDocumentId()))

	id := int(req.GetDocumentId())
	if id <= 0 {
		return nil, merchantdocument_errors.ErrGrpcMerchantInvalidID
	}

	_, err := s.merchantDocumentCommand.DeleteMerchantDocumentPermanent(ctx, id)
	if err != nil {
		s.logger.Error("DeletePermanent merchant document failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("DeletePermanent merchant document success")

	return &pbmerchantdoc.ApiResponseMerchantDocumentDelete{
		Status:  "success",
		Message: "Successfully permanently deleted merchant document",
	}, nil
}

func (s *merchantDocumentHandleGrpc) RestoreAll(ctx context.Context, _ *emptypb.Empty) (*pbmerchantdoc.ApiResponseMerchantDocumentAll, error) {
	s.logger.Info("RestoreAll merchant documents called")

	_, err := s.merchantDocumentCommand.RestoreAllMerchantDocument(ctx)
	if err != nil {
		s.logger.Error("RestoreAll merchant documents failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("RestoreAll merchant documents success")

	return &pbmerchantdoc.ApiResponseMerchantDocumentAll{
		Status:  "success",
		Message: "Successfully restored all merchant documents",
	}, nil
}

func (s *merchantDocumentHandleGrpc) DeleteAllPermanent(ctx context.Context, _ *emptypb.Empty) (*pbmerchantdoc.ApiResponseMerchantDocumentAll, error) {
	s.logger.Info("DeleteAllPermanent merchant documents called")

	_, err := s.merchantDocumentCommand.DeleteAllMerchantDocumentPermanent(ctx)
	if err != nil {
		s.logger.Error("DeleteAllPermanent merchant documents failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("DeleteAllPermanent merchant documents success")

	return &pbmerchantdoc.ApiResponseMerchantDocumentAll{
		Status:  "success",
		Message: "Successfully permanently deleted all merchant documents",
	}, nil
}

// Map helpers
func mapMerchantDocument(doc *db.MerchantDocument) *pbmerchantdoc.MerchantDocument {
	if doc == nil {
		return nil
	}
	return &pbmerchantdoc.MerchantDocument{
		DocumentId:   int32(doc.DocumentID),
		MerchantId:   int32(doc.MerchantID),
		DocumentType: doc.DocumentType,
		DocumentUrl:  doc.DocumentUrl,
		Status:       doc.Status,
		Note:         mapSqlNullString(doc.Note),
		UploadedAt:   mapSqlNullTime(doc.CreatedAt),
		UpdatedAt:    mapSqlNullTime(doc.UpdatedAt),
	}
}

func mapResponsesGetMerchantDocumentsRow(docs []*db.GetMerchantDocumentsRow) []*pbmerchantdoc.MerchantDocument {
	var res []*pbmerchantdoc.MerchantDocument
	for _, doc := range docs {
		res = append(res, &pbmerchantdoc.MerchantDocument{
			DocumentId:   int32(doc.DocumentID),
			MerchantId:   int32(doc.MerchantID),
			DocumentType: doc.DocumentType,
			DocumentUrl:  doc.DocumentUrl,
			Status:       doc.Status,
			Note:         mapSqlNullString(doc.Note),
			UploadedAt:   mapSqlNullTime(doc.CreatedAt),
			UpdatedAt:    mapSqlNullTime(doc.UpdatedAt),
		})
	}
	return res
}

func mapResponsesGetActiveMerchantDocumentsRow(docs []*db.GetActiveMerchantDocumentsRow) []*pbmerchantdoc.MerchantDocument {
	var res []*pbmerchantdoc.MerchantDocument
	for _, doc := range docs {
		res = append(res, &pbmerchantdoc.MerchantDocument{
			DocumentId:   int32(doc.DocumentID),
			MerchantId:   int32(doc.MerchantID),
			DocumentType: doc.DocumentType,
			DocumentUrl:  doc.DocumentUrl,
			Status:       doc.Status,
			Note:         mapSqlNullString(doc.Note),
			UploadedAt:   mapSqlNullTime(doc.CreatedAt),
			UpdatedAt:    mapSqlNullTime(doc.UpdatedAt),
		})
	}
	return res
}

func mapMerchantDocumentDeleteAt(doc *db.MerchantDocument) *pbmerchantdoc.MerchantDocumentDeleteAt {
	if doc == nil {
		return nil
	}
	var deletedAt *wrapperspb.StringValue
	if doc.DeletedAt.Valid {
		deletedAt = wrapperspb.String(doc.DeletedAt.Time.Format("2006-01-02 15:04:05"))
	}

	return &pbmerchantdoc.MerchantDocumentDeleteAt{
		DocumentId:   int32(doc.DocumentID),
		MerchantId:   int32(doc.MerchantID),
		DocumentType: doc.DocumentType,
		DocumentUrl:  doc.DocumentUrl,
		Status:       doc.Status,
		Note:         mapSqlNullString(doc.Note),
		UploadedAt:   mapSqlNullTime(doc.CreatedAt),
		UpdatedAt:    mapSqlNullTime(doc.UpdatedAt),
		DeletedAt:    deletedAt,
	}
}

func mapResponsesGetTrashedMerchantDocumentsRow(docs []*db.GetTrashedMerchantDocumentsRow) []*pbmerchantdoc.MerchantDocumentDeleteAt {
	var res []*pbmerchantdoc.MerchantDocumentDeleteAt
	for _, doc := range docs {
		var deletedAt *wrapperspb.StringValue
		if doc.DeletedAt.Valid {
			deletedAt = wrapperspb.String(doc.DeletedAt.Time.Format("2006-01-02 15:04:05"))
		}
		res = append(res, &pbmerchantdoc.MerchantDocumentDeleteAt{
			DocumentId:   int32(doc.DocumentID),
			MerchantId:   int32(doc.MerchantID),
			DocumentType: doc.DocumentType,
			DocumentUrl:  doc.DocumentUrl,
			Status:       doc.Status,
			Note:         mapSqlNullString(doc.Note),
			UploadedAt:   mapSqlNullTime(doc.CreatedAt),
			UpdatedAt:    mapSqlNullTime(doc.UpdatedAt),
			DeletedAt:    deletedAt,
		})
	}
	return res
}
