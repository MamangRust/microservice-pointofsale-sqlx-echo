package handler

import (
	"context"
	"math"

	pbcommon "github.com/MamangRust/microservice-point-of-sale-pb/common"
	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/transaction_errors"
	db "github.com/MamangRust/microservice-point-of-sale-transacton/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-transacton/service"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type transactionHandleGrpc struct {
	pbtransaction.UnimplementedTransactionServiceServer
	transactionQuery   service.TransactionQueryService
	transactionCommand service.TransactionCommandService
	logger             logger.LoggerInterface
}

func NewTransactionHandleGrpc(
	service *service.Service,
	logger logger.LoggerInterface,
) pbtransaction.TransactionServiceServer {
	return &transactionHandleGrpc{
		transactionQuery:   service.TransactionQuery,
		transactionCommand: service.TransactionCommand,
		logger:             logger,
	}
}

func (s *transactionHandleGrpc) FindAll(ctx context.Context, request *pbtransaction.FindAllTransactionRequest) (*pbtransaction.ApiResponsePaginationTransaction, error) {
	s.logger.Info("FindAll transactions called", zap.Int32("page", request.GetPage()))

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllTransaction{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	transaction, totalRecords, err := s.transactionQuery.FindAllTransactions(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindAll transactions failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindAll transactions success")

	return &pbtransaction.ApiResponsePaginationTransaction{
		Status:     "success",
		Message:    "Successfully fetched transaction",
		Data:       mapResponsesTransaction(transaction),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *transactionHandleGrpc) FindByMerchant(ctx context.Context, request *pbtransaction.FindAllTransactionMerchantRequest) (*pbtransaction.ApiResponsePaginationTransaction, error) {
	s.logger.Info("FindByMerchant transactions called", zap.Int32("merchantId", request.GetMerchantId()))

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()
	merchant_id := int(request.GetMerchantId())

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllTransactionByMerchant{
		MerchantID: merchant_id,
		Page:       page,
		PageSize:   pageSize,
		Search:     search,
	}

	transaction, totalRecords, err := s.transactionQuery.FindByMerchant(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindByMerchant transactions failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindByMerchant transactions success")

	return &pbtransaction.ApiResponsePaginationTransaction{
		Status:     "success",
		Message:    "Successfully fetched transaction",
		Data:       mapResponsesTransactionByMerchant(transaction),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *transactionHandleGrpc) FindById(ctx context.Context, request *pbtransaction.FindByIdTransactionRequest) (*pbtransaction.ApiResponseTransaction, error) {
	s.logger.Info("FindById transaction called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, transaction_errors.ErrGrpcInvalidID
	}

	transaction, err := s.transactionQuery.FindById(ctx, id)
	if err != nil {
		s.logger.Error("FindById transaction failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("FindById transaction success")

	return &pbtransaction.ApiResponseTransaction{
		Status:  "success",
		Message: "Successfully fetched transaction",
		Data:    mapResponseTransaction(transaction),
	}, nil
}

func (s *transactionHandleGrpc) FindByActive(ctx context.Context, request *pbtransaction.FindAllTransactionRequest) (*pbtransaction.ApiResponsePaginationTransactionDeleteAt, error) {
	s.logger.Info("FindByActive transactions called", zap.Int32("page", request.GetPage()))

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllTransaction{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	transaction, totalRecords, err := s.transactionQuery.FindByActive(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindByActive transactions failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindByActive transactions success")

	return &pbtransaction.ApiResponsePaginationTransactionDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched active transaction",
		Data:       mapResponsesTransactionActive(transaction),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *transactionHandleGrpc) FindByTrashed(ctx context.Context, request *pbtransaction.FindAllTransactionRequest) (*pbtransaction.ApiResponsePaginationTransactionDeleteAt, error) {
	s.logger.Info("FindByTrashed transactions called")

	page := int(request.GetPage())
	pageSize := int(request.GetPageSize())
	search := request.GetSearch()

	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	reqService := requests.FindAllTransaction{
		Page:     page,
		PageSize: pageSize,
		Search:   search,
	}

	transaction, totalRecords, err := s.transactionQuery.FindByTrashed(ctx, &reqService)
	if err != nil {
		s.logger.Error("FindByTrashed transactions failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	totalPages := int(math.Ceil(float64(*totalRecords) / float64(pageSize)))

	paginationMeta := &pbcommon.PaginationMeta{
		CurrentPage:  int32(page),
		PageSize:     int32(pageSize),
		TotalPages:   int32(totalPages),
		TotalRecords: int32(*totalRecords),
	}

	s.logger.Info("FindByTrashed transactions success")

	return &pbtransaction.ApiResponsePaginationTransactionDeleteAt{
		Status:     "success",
		Message:    "Successfully fetched trashed transaction",
		Data:       mapResponsesTransactionTrashed(transaction),
		Pagination: mapPaginationMeta(paginationMeta),
	}, nil
}

func (s *transactionHandleGrpc) Create(ctx context.Context, request *pbtransaction.CreateTransactionRequest) (*pbtransaction.ApiResponseTransaction, error) {
	s.logger.Info("Create transaction called", zap.Int32("orderId", request.GetOrderId()))

	req := &requests.CreateTransactionRequest{
		CashierID:     int(request.GetCashierId()),
		OrderID:       int(request.GetOrderId()),
		PaymentMethod: request.GetPaymentMethod(),
		Amount:        int(request.GetAmount()),
	}

	if err := req.Validate(); err != nil {
		return nil, transaction_errors.ErrGrpcValidateCreateTransaction
	}

	transaction, err := s.transactionCommand.CreateTransaction(ctx, req)
	if err != nil {
		s.logger.Error("Create transaction failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("Create transaction success")

	return &pbtransaction.ApiResponseTransaction{
		Status:  "success",
		Message: "Successfully created transaction",
		Data:    mapResponseTransaction(transaction),
	}, nil
}

func (s *transactionHandleGrpc) Update(ctx context.Context, request *pbtransaction.UpdateTransactionRequest) (*pbtransaction.ApiResponseTransaction, error) {
	s.logger.Info("Update transaction called", zap.Int32("id", request.GetTransactionId()))

	id := int(request.GetTransactionId())
	if id <= 0 {
		return nil, transaction_errors.ErrGrpcInvalidID
	}

	req := &requests.UpdateTransactionRequest{
		TransactionID: &id,
		OrderID:       int(request.GetOrderId()),
		CashierID:     int(request.GetCashierId()),
		PaymentMethod: request.GetPaymentMethod(),
		Amount:        int(request.GetAmount()),
	}

	if err := req.Validate(); err != nil {
		return nil, transaction_errors.ErrGrpcValidateUpdateTransaction
	}

	transaction, err := s.transactionCommand.UpdateTransaction(ctx, req)
	if err != nil {
		s.logger.Error("Update transaction failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("Update transaction success")

	return &pbtransaction.ApiResponseTransaction{
		Status:  "success",
		Message: "Successfully updated transaction",
		Data:    mapResponseTransaction(transaction),
	}, nil
}

func (s *transactionHandleGrpc) TrashedTransaction(ctx context.Context, request *pbtransaction.FindByIdTransactionRequest) (*pbtransaction.ApiResponseTransactionDeleteAt, error) {
	s.logger.Info("TrashedTransaction called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, transaction_errors.ErrGrpcInvalidID
	}

	transaction, err := s.transactionCommand.TrashedTransaction(ctx, id)
	if err != nil {
		s.logger.Error("TrashedTransaction failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("TrashedTransaction success")

	return &pbtransaction.ApiResponseTransactionDeleteAt{
		Status:  "success",
		Message: "Successfully trashed transaction",
		Data:    mapResponseTransactionDeleteAt(transaction),
	}, nil
}

func (s *transactionHandleGrpc) RestoreTransaction(ctx context.Context, request *pbtransaction.FindByIdTransactionRequest) (*pbtransaction.ApiResponseTransactionDeleteAt, error) {
	s.logger.Info("RestoreTransaction called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, transaction_errors.ErrGrpcInvalidID
	}

	transaction, err := s.transactionCommand.RestoreTransaction(ctx, id)
	if err != nil {
		s.logger.Error("RestoreTransaction failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("RestoreTransaction success")

	return &pbtransaction.ApiResponseTransactionDeleteAt{
		Status:  "success",
		Message: "Successfully restored transaction",
		Data:    mapResponseTransactionDeleteAt(transaction),
	}, nil
}

func (s *transactionHandleGrpc) DeleteTransactionPermanent(ctx context.Context, request *pbtransaction.FindByIdTransactionRequest) (*pbtransaction.ApiResponseTransactionDelete, error) {
	s.logger.Info("DeleteTransactionPermanent called", zap.Int32("id", request.GetId()))

	id := int(request.GetId())
	if id <= 0 {
		return nil, transaction_errors.ErrGrpcInvalidID
	}

	_, err := s.transactionCommand.DeleteTransactionPermanently(ctx, id)
	if err != nil {
		s.logger.Error("DeleteTransactionPermanent failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("DeleteTransactionPermanent success")

	return &pbtransaction.ApiResponseTransactionDelete{
		Status:  "success",
		Message: "Successfully deleted Transaction permanently",
	}, nil
}

func (s *transactionHandleGrpc) RestoreAllTransaction(ctx context.Context, _ *emptypb.Empty) (*pbtransaction.ApiResponseTransactionAll, error) {
	s.logger.Info("RestoreAllTransaction called")

	_, err := s.transactionCommand.RestoreAllTransactions(ctx)
	if err != nil {
		s.logger.Error("RestoreAllTransaction failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("RestoreAllTransaction success")

	return &pbtransaction.ApiResponseTransactionAll{
		Status:  "success",
		Message: "Successfully restore all Transaction",
	}, nil
}

func (s *transactionHandleGrpc) DeleteAllTransactionPermanent(ctx context.Context, _ *emptypb.Empty) (*pbtransaction.ApiResponseTransactionAll, error) {
	s.logger.Info("DeleteAllTransactionPermanent called")

	_, err := s.transactionCommand.DeleteAllTransactionPermanent(ctx)
	if err != nil {
		s.logger.Error("DeleteAllTransactionPermanent failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("DeleteAllTransactionPermanent success")

	return &pbtransaction.ApiResponseTransactionAll{
		Status:  "success",
		Message: "Successfully delete Transaction permanen",
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

func mapResponseTransaction(transaction *db.Transaction) *pbtransaction.TransactionResponse {
	if transaction == nil {
		return nil
	}
	var createdAtStr, updatedAtStr string
	if transaction.CreatedAt.Valid {
		createdAtStr = transaction.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if transaction.UpdatedAt.Valid {
		updatedAtStr = transaction.UpdatedAt.Time.Format("2006-01-02 15:04:05")
	}
	var changeAmount int32
	if transaction.ChangeAmount != nil {
		changeAmount = *transaction.ChangeAmount
	}
	return &pbtransaction.TransactionResponse{
		Id:            int32(transaction.TransactionID),
		OrderId:       int32(transaction.OrderID),
		MerchantId:    int32(transaction.MerchantID),
		PaymentMethod: transaction.PaymentMethod,
		Amount:        transaction.Amount,
		ChangeAmount:  changeAmount,
		PaymentStatus: transaction.PaymentStatus,
		CreatedAt:     createdAtStr,
		UpdatedAt:     updatedAtStr,
	}
}

func mapResponsesTransaction(transactions []*db.GetTransactionsRow) []*pbtransaction.TransactionResponse {
	var mappedTransactions []*pbtransaction.TransactionResponse
	for _, t := range transactions {
		var createdAtStr, updatedAtStr string
		if t.CreatedAt.Valid {
			createdAtStr = t.CreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		if t.UpdatedAt.Valid {
			updatedAtStr = t.UpdatedAt.Time.Format("2006-01-02 15:04:05")
		}
		var changeAmount int32
		if t.ChangeAmount != nil {
			changeAmount = *t.ChangeAmount
		}
		mappedTransactions = append(mappedTransactions, &pbtransaction.TransactionResponse{
			Id:            int32(t.TransactionID),
			OrderId:       int32(t.OrderID),
			MerchantId:    int32(t.MerchantID),
			PaymentMethod: t.PaymentMethod,
			Amount:        t.Amount,
			ChangeAmount:  changeAmount,
			PaymentStatus: t.PaymentStatus,
			CreatedAt:     createdAtStr,
			UpdatedAt:     updatedAtStr,
		})
	}
	return mappedTransactions
}

func mapResponsesTransactionByMerchant(transactions []*db.GetTransactionByMerchantRow) []*pbtransaction.TransactionResponse {
	var mappedTransactions []*pbtransaction.TransactionResponse
	for _, t := range transactions {
		var createdAtStr, updatedAtStr string
		if t.CreatedAt.Valid {
			createdAtStr = t.CreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		if t.UpdatedAt.Valid {
			updatedAtStr = t.UpdatedAt.Time.Format("2006-01-02 15:04:05")
		}
		var changeAmount int32
		if t.ChangeAmount != nil {
			changeAmount = *t.ChangeAmount
		}
		mappedTransactions = append(mappedTransactions, &pbtransaction.TransactionResponse{
			Id:            int32(t.TransactionID),
			OrderId:       int32(t.OrderID),
			MerchantId:    int32(t.MerchantID),
			PaymentMethod: t.PaymentMethod,
			Amount:        t.Amount,
			ChangeAmount:  changeAmount,
			PaymentStatus: t.PaymentStatus,
			CreatedAt:     createdAtStr,
			UpdatedAt:     updatedAtStr,
		})
	}
	return mappedTransactions
}

func mapResponsesTransactionActive(transactions []*db.GetTransactionsActiveRow) []*pbtransaction.TransactionResponseDeleteAt {
	var mappedTransactions []*pbtransaction.TransactionResponseDeleteAt
	for _, t := range transactions {
		var createdAtStr, updatedAtStr string
		if t.CreatedAt.Valid {
			createdAtStr = t.CreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		if t.UpdatedAt.Valid {
			updatedAtStr = t.UpdatedAt.Time.Format("2006-01-02 15:04:05")
		}
		var deletedAt *wrapperspb.StringValue
		if t.DeletedAt.Valid {
			deletedAt = wrapperspb.String(t.DeletedAt.Time.Format("2006-01-02 15:04:05"))
		}
		var changeAmount int32
		if t.ChangeAmount != nil {
			changeAmount = *t.ChangeAmount
		}
		mappedTransactions = append(mappedTransactions, &pbtransaction.TransactionResponseDeleteAt{
			Id:            int32(t.TransactionID),
			OrderId:       int32(t.OrderID),
			MerchantId:    int32(t.MerchantID),
			PaymentMethod: t.PaymentMethod,
			Amount:        t.Amount,
			ChangeAmount:  changeAmount,
			PaymentStatus: t.PaymentStatus,
			CreatedAt:     createdAtStr,
			UpdatedAt:     updatedAtStr,
			DeletedAt:     deletedAt,
		})
	}
	return mappedTransactions
}

func mapResponsesTransactionTrashed(transactions []*db.GetTransactionsTrashedRow) []*pbtransaction.TransactionResponseDeleteAt {
	var mappedTransactions []*pbtransaction.TransactionResponseDeleteAt
	for _, t := range transactions {
		var createdAtStr, updatedAtStr string
		if t.CreatedAt.Valid {
			createdAtStr = t.CreatedAt.Time.Format("2006-01-02 15:04:05")
		}
		if t.UpdatedAt.Valid {
			updatedAtStr = t.UpdatedAt.Time.Format("2006-01-02 15:04:05")
		}
		var deletedAt *wrapperspb.StringValue
		if t.DeletedAt.Valid {
			deletedAt = wrapperspb.String(t.DeletedAt.Time.Format("2006-01-02 15:04:05"))
		}
		var changeAmount int32
		if t.ChangeAmount != nil {
			changeAmount = *t.ChangeAmount
		}
		mappedTransactions = append(mappedTransactions, &pbtransaction.TransactionResponseDeleteAt{
			Id:            int32(t.TransactionID),
			OrderId:       int32(t.OrderID),
			MerchantId:    int32(t.MerchantID),
			PaymentMethod: t.PaymentMethod,
			Amount:        t.Amount,
			ChangeAmount:  changeAmount,
			PaymentStatus: t.PaymentStatus,
			CreatedAt:     createdAtStr,
			UpdatedAt:     updatedAtStr,
			DeletedAt:     deletedAt,
		})
	}
	return mappedTransactions
}

func mapResponseTransactionDeleteAt(transaction *db.Transaction) *pbtransaction.TransactionResponseDeleteAt {
	if transaction == nil {
		return nil
	}
	var createdAtStr, updatedAtStr string
	if transaction.CreatedAt.Valid {
		createdAtStr = transaction.CreatedAt.Time.Format("2006-01-02 15:04:05")
	}
	if transaction.UpdatedAt.Valid {
		updatedAtStr = transaction.UpdatedAt.Time.Format("2006-01-02 15:04:05")
	}
	var deletedAt *wrapperspb.StringValue
	if transaction.DeletedAt.Valid {
		deletedAt = wrapperspb.String(transaction.DeletedAt.Time.Format("2006-01-02 15:04:05"))
	}
	var changeAmount int32
	if transaction.ChangeAmount != nil {
		changeAmount = *transaction.ChangeAmount
	}
	return &pbtransaction.TransactionResponseDeleteAt{
		Id:            int32(transaction.TransactionID),
		OrderId:       int32(transaction.OrderID),
		MerchantId:    int32(transaction.MerchantID),
		PaymentMethod: transaction.PaymentMethod,
		Amount:        transaction.Amount,
		ChangeAmount:  changeAmount,
		PaymentStatus: transaction.PaymentStatus,
		CreatedAt:     createdAtStr,
		UpdatedAt:     updatedAtStr,
		DeletedAt:     deletedAt,
	}
}
