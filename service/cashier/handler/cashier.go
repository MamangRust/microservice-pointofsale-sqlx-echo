package handler

import (
	"context"
	"math"

	db "github.com/MamangRust/microservice-point-of-sale-cashier/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-cashier/service"
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbcommon "github.com/MamangRust/microservice-point-of-sale-pb/common"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/cashier_errors"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type cashierHandleGrpc struct {
	pbcashier.UnimplementedCashierServiceServer
	cashierQuery   service.CashierQueryService
	cashierCommand service.CashierCommandService
	logger         logger.LoggerInterface
}

func NewCashierHandleGrpc(
	service *service.Service,
	logger logger.LoggerInterface,
) pbcashier.CashierServiceServer {
	return &cashierHandleGrpc{
		cashierQuery:   service.CashierQuery,
		cashierCommand: service.CashierCommand,
		logger:         logger,
	}
}

func (s *cashierHandleGrpc) FindAll(ctx context.Context, request *pbcashier.FindAllCashierRequest) (*pbcashier.ApiResponsePaginationCashier, error) {
	s.logger.Info("FindAll cashier called", zap.Int32("page", request.GetPage()), zap.Int32("pageSize", request.GetPageSize()))

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllCashiers{
		Search:   search,
		Page:     page,
		PageSize: pageSize,
	}

	cashier, totalRecords, err := s.cashierQuery.FindAll(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindAll cashier failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindAll cashier success", zap.Int("count", len(cashier)))

	return &pbcashier.ApiResponsePaginationCashier{
		Status:     "success",
		Message:    "Successfully fetched cashier",
		Data:       mapResponsesCashier(cashier),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *cashierHandleGrpc) FindById(ctx context.Context, request *pbcashier.FindByIdCashierRequest) (*pbcashier.ApiResponseCashier, error) {
	s.logger.Info("FindById cashier called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, cashier_errors.ErrGrpcFailedInvalidId
	}

	cashier, err := s.cashierQuery.FindById(ctx, id)
	if err != nil {
		s.logger.Error("FindById cashier failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("FindById cashier success", zap.Int("id", id))

	return &pbcashier.ApiResponseCashier{
		Status:  "success",
		Message: "Successfully fetched categories",
		Data:    mapResponseCashier(cashier),
	}, nil
}

func (s *cashierHandleGrpc) FindByActive(ctx context.Context, request *pbcashier.FindAllCashierRequest) (*pbcashier.ApiResponsePaginationCashierDeleteAt, error) {
	s.logger.Info("FindByActive cashier called", zap.Int32("page", request.GetPage()))

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllCashiers{
		Search:   search,
		Page:     page,
		PageSize: pageSize,
	}

	cashier, totalRecords, err := s.cashierQuery.FindByActive(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindByActive cashier failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindByActive cashier success", zap.Int("count", len(cashier)))

	return &pbcashier.ApiResponsePaginationCashierDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched active cashier",
		Data:       mapResponsesCashierActive(cashier),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *cashierHandleGrpc) FindByTrashed(ctx context.Context, request *pbcashier.FindAllCashierRequest) (*pbcashier.ApiResponsePaginationCashierDeleteAt, error) {
	s.logger.Info("FindByTrashed cashier called")

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllCashiers{
		Search:   search,
		Page:     page,
		PageSize: pageSize,
	}

	users, totalRecords, err := s.cashierQuery.FindByTrashed(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindByTrashed cashier failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindByTrashed cashier success", zap.Int("count", len(users)))

	return &pbcashier.ApiResponsePaginationCashierDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched trashed cashier",
		Data:       mapResponsesCashierTrashed(users),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *cashierHandleGrpc) FindByMerchant(ctx context.Context, request *pbcashier.FindByMerchantCashierRequest) (*pbcashier.ApiResponsePaginationCashier, error) {
	s.logger.Info("FindByMerchant cashier called", zap.Int32("merchantId", request.GetMerchantId()))

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()
	merchant_id := int(request.GetMerchantId())

	if merchant_id <= 0 {
		return nil, cashier_errors.ErrGrpcFailedInvalidMerchantId
	}

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllCashierMerchant{
		Search:     search,
		Page:       page,
		PageSize:   pageSize,
		MerchantID: merchant_id,
	}

	cashier, totalRecords, err := s.cashierQuery.FindByMerchant(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindByMerchant cashier failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindByMerchant cashier success", zap.Int("count", len(cashier)))

	return &pbcashier.ApiResponsePaginationCashier{
		Status:     "success",
		Message:    "Successfully fetched cashier",
		Data:       mapResponsesCashierByMerchant(cashier),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *cashierHandleGrpc) CreateCashier(ctx context.Context, request *pbcashier.CreateCashierRequest) (*pbcashier.ApiResponseCashier, error) {
	s.logger.Info("CreateCashier called", zap.String("name", request.GetName()))

	req := &requests.CreateCashierRequest{
		Name:       request.GetName(),
		MerchantID: int(request.GetMerchantId()),
		UserID:     int(request.GetUserId()),
	}

	if err := req.Validate(); err != nil {
		return nil, cashier_errors.ErrGrpcValidateCreateCashier
	}

	cashier, err := s.cashierCommand.CreateCashier(ctx, req)
	if err != nil {
		s.logger.Error("CreateCashier failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("CreateCashier success")

	return &pbcashier.ApiResponseCashier{
		Status:  "success",
		Message: "Successfully created cashier",
		Data:    mapResponseCashier(cashier),
	}, nil
}

func (s *cashierHandleGrpc) UpdateCashier(ctx context.Context, request *pbcashier.UpdateCashierRequest) (*pbcashier.ApiResponseCashier, error) {
	s.logger.Info("UpdateCashier called", zap.Int32("id", request.GetCashierId()))

	id := int(request.GetCashierId())
	if id <= 0 {
		return nil, cashier_errors.ErrGrpcFailedInvalidId
	}

	req := &requests.UpdateCashierRequest{
		CashierID: &id,
		Name:      request.GetName(),
	}

	if err := req.Validate(); err != nil {
		return nil, cashier_errors.ErrGrpcValidateUpdateCashier
	}

	cashier, err := s.cashierCommand.UpdateCashier(ctx, req)
	if err != nil {
		s.logger.Error("UpdateCashier failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("UpdateCashier success")

	return &pbcashier.ApiResponseCashier{
		Status:  "success",
		Message: "Successfully updated cashier",
		Data:    mapResponseCashier(cashier),
	}, nil
}

func (s *cashierHandleGrpc) TrashedCashier(ctx context.Context, request *pbcashier.FindByIdCashierRequest) (*pbcashier.ApiResponseCashierDeleteAt, error) {
	s.logger.Info("TrashedCashier called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, cashier_errors.ErrGrpcFailedInvalidId
	}

	cashier, err := s.cashierCommand.TrashedCashier(ctx, id)
	if err != nil {
		s.logger.Error("TrashedCashier failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("TrashedCashier success")

	return &pbcashier.ApiResponseCashierDeleteAt{
		Status:  "success",
		Message: "Successfully trashed cashier",
		Data:    mapResponseCashierDeleteAt(cashier),
	}, nil
}

func (s *cashierHandleGrpc) RestoreCashier(ctx context.Context, request *pbcashier.FindByIdCashierRequest) (*pbcashier.ApiResponseCashierDeleteAt, error) {
	s.logger.Info("RestoreCashier called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, cashier_errors.ErrGrpcFailedInvalidId
	}

	cashier, err := s.cashierCommand.RestoreCashier(ctx, id)
	if err != nil {
		s.logger.Error("RestoreCashier failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("RestoreCashier success")

	return &pbcashier.ApiResponseCashierDeleteAt{
		Status:  "success",
		Message: "Successfully restored cashier",
		Data:    mapResponseCashierDeleteAt(cashier),
	}, nil
}

func (s *cashierHandleGrpc) DeleteCashierPermanent(ctx context.Context, request *pbcashier.FindByIdCashierRequest) (*pbcashier.ApiResponseCashierDelete, error) {
	s.logger.Info("DeleteCashierPermanent called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, cashier_errors.ErrGrpcFailedInvalidId
	}

	_, err := s.cashierCommand.DeleteCashierPermanent(ctx, id)
	if err != nil {
		s.logger.Error("DeleteCashierPermanent failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("DeleteCashierPermanent success")

	return &pbcashier.ApiResponseCashierDelete{
		Status:  "success",
		Message: "Successfully deleted cashier permanently",
	}, nil
}

func (s *cashierHandleGrpc) RestoreAllCashier(ctx context.Context, _ *emptypb.Empty) (*pbcashier.ApiResponseCashierAll, error) {
	s.logger.Info("RestoreAllCashier called")

	_, err := s.cashierCommand.RestoreAllCashier(ctx)
	if err != nil {
		s.logger.Error("RestoreAllCashier failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("RestoreAllCashier success")

	return &pbcashier.ApiResponseCashierAll{
		Status:  "success",
		Message: "Successfully restore all cashier",
	}, nil
}

func (s *cashierHandleGrpc) DeleteAllCashierPermanent(ctx context.Context, _ *emptypb.Empty) (*pbcashier.ApiResponseCashierAll, error) {
	s.logger.Info("DeleteAllCashierPermanent called")

	_, err := s.cashierCommand.DeleteAllCashierPermanent(ctx)
	if err != nil {
		s.logger.Error("DeleteAllCashierPermanent failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("DeleteAllCashierPermanent success")

	return &pbcashier.ApiResponseCashierAll{
		Status:  "success",
		Message: "Successfully delete cashier permanen",
	}, nil
}

// Map helpers
func mapPaginationMeta(meta *pbcommon.PaginationMeta) *pbcommon.PaginationMeta {
	if meta == nil {
		return nil
	}
	return &pbcommon.PaginationMeta{
		CurrentPage:  meta.CurrentPage,
		PageSize:     meta.PageSize,
		TotalPages:   meta.TotalPages,
		TotalRecords: meta.TotalRecords,
	}
}

func mapResponseCashier(cashier *db.Cashier) *pbcashier.CashierResponse {
	if cashier == nil {
		return nil
	}
	var createdAtStr, updatedAtStr string
	if cashier.CreatedAt.Valid {
		createdAtStr = cashier.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if cashier.UpdatedAt.Valid {
		updatedAtStr = cashier.UpdatedAt.Time.Format("2006-01-02 15:04:05")
	}
	return &pbcashier.CashierResponse{
		Id:         int32(cashier.CashierID),
		MerchantId: int32(cashier.MerchantID),
		Name:       cashier.Name,
		CreatedAt:  createdAtStr,
		UpdatedAt:  updatedAtStr,
	}
}

func mapResponsesCashier(cashiers []*db.GetCashiersRow) []*pbcashier.CashierResponse {
	var mappedCashiers []*pbcashier.CashierResponse
	for _, cashier := range cashiers {
		var createdAtStr, updatedAtStr string
		if cashier.CreatedAt.Valid {
			createdAtStr = cashier.CreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		if cashier.UpdatedAt.Valid {
			updatedAtStr = cashier.UpdatedAt.Time.Format("2006-01-02 15:04:05")
		}
		mappedCashiers = append(mappedCashiers, &pbcashier.CashierResponse{
			Id:         int32(cashier.CashierID),
			MerchantId: int32(cashier.MerchantID),
			Name:       cashier.Name,
			CreatedAt:  createdAtStr,
			UpdatedAt:  updatedAtStr,
		})
	}
	return mappedCashiers
}

func mapResponsesCashierByMerchant(cashiers []*db.GetCashiersByMerchantRow) []*pbcashier.CashierResponse {
	var mappedCashiers []*pbcashier.CashierResponse
	for _, cashier := range cashiers {
		var createdAtStr, updatedAtStr string
		if cashier.CreatedAt.Valid {
			createdAtStr = cashier.CreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		if cashier.UpdatedAt.Valid {
			updatedAtStr = cashier.UpdatedAt.Time.Format("2006-01-02 15:04:05")
		}
		mappedCashiers = append(mappedCashiers, &pbcashier.CashierResponse{
			Id:         int32(cashier.CashierID),
			MerchantId: int32(cashier.MerchantID),
			Name:       cashier.Name,
			CreatedAt:  createdAtStr,
			UpdatedAt:  updatedAtStr,
		})
	}
	return mappedCashiers
}

func mapResponseCashierDeleteAt(cashier *db.Cashier) *pbcashier.CashierResponseDeleteAt {
	if cashier == nil {
		return nil
	}
	var createdAtStr, updatedAtStr string
	if cashier.CreatedAt.Valid {
		createdAtStr = cashier.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if cashier.UpdatedAt.Valid {
		updatedAtStr = cashier.UpdatedAt.Time.Format("2006-01-02 15:04:05")
	}
	var deletedAt *wrapperspb.StringValue
	if cashier.DeletedAt.Valid {
		deletedAt = wrapperspb.String(cashier.DeletedAt.Time.Format("2006-01-02 15:04:05"))
	}

	return &pbcashier.CashierResponseDeleteAt{
		Id:         int32(cashier.CashierID),
		MerchantId: int32(cashier.MerchantID),
		Name:       cashier.Name,
		CreatedAt:  createdAtStr,
		UpdatedAt:  updatedAtStr,
		DeletedAt:  deletedAt,
	}
}

func mapResponsesCashierActive(cashiers []*db.GetCashiersActiveRow) []*pbcashier.CashierResponseDeleteAt {
	var mappedCashiers []*pbcashier.CashierResponseDeleteAt
	for _, cashier := range cashiers {
		var createdAtStr, updatedAtStr string
		if cashier.CreatedAt.Valid {
			createdAtStr = cashier.CreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		if cashier.UpdatedAt.Valid {
			updatedAtStr = cashier.UpdatedAt.Time.Format("2006-01-02 15:04:05")
		}
		var deletedAt *wrapperspb.StringValue
		if cashier.DeletedAt.Valid {
			deletedAt = wrapperspb.String(cashier.DeletedAt.Time.Format("2006-01-02 15:04:05"))
		}
		mappedCashiers = append(mappedCashiers, &pbcashier.CashierResponseDeleteAt{
			Id:         int32(cashier.CashierID),
			MerchantId: int32(cashier.MerchantID),
			Name:       cashier.Name,
			CreatedAt:  createdAtStr,
			UpdatedAt:  updatedAtStr,
			DeletedAt:  deletedAt,
		})
	}
	return mappedCashiers
}

func mapResponsesCashierTrashed(cashiers []*db.GetCashiersTrashedRow) []*pbcashier.CashierResponseDeleteAt {
	var mappedCashiers []*pbcashier.CashierResponseDeleteAt
	for _, cashier := range cashiers {
		var createdAtStr, updatedAtStr string
		if cashier.CreatedAt.Valid {
			createdAtStr = cashier.CreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		if cashier.UpdatedAt.Valid {
			updatedAtStr = cashier.UpdatedAt.Time.Format("2006-01-02 15:04:05")
		}
		var deletedAt *wrapperspb.StringValue
		if cashier.DeletedAt.Valid {
			deletedAt = wrapperspb.String(cashier.DeletedAt.Time.Format("2006-01-02 15:04:05"))
		}
		mappedCashiers = append(mappedCashiers, &pbcashier.CashierResponseDeleteAt{
			Id:         int32(cashier.CashierID),
			MerchantId: int32(cashier.MerchantID),
			Name:       cashier.Name,
			CreatedAt:  createdAtStr,
			UpdatedAt:  updatedAtStr,
			DeletedAt:  deletedAt,
		})
	}
	return mappedCashiers
}
