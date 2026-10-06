package protomapper

import (
	pbcommon "github.com/MamangRust/microservice-point-of-sale-pb/common"
	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

type transactionProtoMapper struct{}

func NewTransactionProtoMapper() *transactionProtoMapper {
	return &transactionProtoMapper{}
}

func (t *transactionProtoMapper) ToProtoResponseTransaction(status string, message string, trans *response.TransactionResponse) *pbtransaction.ApiResponseTransaction {
	return &pbtransaction.ApiResponseTransaction{
		Status:  status,
		Message: message,
		Data:    t.mapResponseTransaction(trans),
	}
}

func (t *transactionProtoMapper) ToProtoResponsesTransaction(status string, message string, transList []*response.TransactionResponse) *pbtransaction.ApiResponsesTransaction {
	return &pbtransaction.ApiResponsesTransaction{
		Status:  status,
		Message: message,
		Data:    t.mapResponsesTransaction(transList),
	}
}

func (t *transactionProtoMapper) ToProtoResponseTransactionDeleteAt(status string, message string, trans *response.TransactionResponseDeleteAt) *pbtransaction.ApiResponseTransactionDeleteAt {
	return &pbtransaction.ApiResponseTransactionDeleteAt{
		Status:  status,
		Message: message,
		Data:    t.mapResponseTransactionDeleteAt(trans),
	}
}

func (t *transactionProtoMapper) ToProtoResponseTransactionDelete(status string, message string) *pbtransaction.ApiResponseTransactionDelete {
	return &pbtransaction.ApiResponseTransactionDelete{
		Status:  status,
		Message: message,
	}
}

func (t *transactionProtoMapper) ToProtoResponseTransactionAll(status string, message string) *pbtransaction.ApiResponseTransactionAll {
	return &pbtransaction.ApiResponseTransactionAll{
		Status:  status,
		Message: message,
	}
}

func (t *transactionProtoMapper) ToProtoResponsePaginationTransactionDeleteAt(pagination *pbcommon.PaginationMeta, status string, message string, transactions []*response.TransactionResponseDeleteAt) *pbtransaction.ApiResponsePaginationTransactionDeleteAt {
	return &pbtransaction.ApiResponsePaginationTransactionDeleteAt{
		Status:     status,
		Message:    message,
		Data:       t.mapResponsesTransactionDeleteAt(transactions),
		Pagination: mapPaginationMeta(pagination),
	}
}

func (t *transactionProtoMapper) ToProtoResponsePaginationTransaction(pagination *pbcommon.PaginationMeta, status string, message string, transactions []*response.TransactionResponse) *pbtransaction.ApiResponsePaginationTransaction {
	return &pbtransaction.ApiResponsePaginationTransaction{
		Status:     status,
		Message:    message,
		Data:       t.mapResponsesTransaction(transactions),
		Pagination: mapPaginationMeta(pagination),
	}
}

func (t *transactionProtoMapper) mapResponseTransaction(transaction *response.TransactionResponse) *pbtransaction.TransactionResponse {
	return &pbtransaction.TransactionResponse{
		Id:            int32(transaction.ID),
		OrderId:       int32(transaction.OrderID),
		MerchantId:    int32(transaction.MerchantID),
		PaymentMethod: transaction.PaymentMethod,
		Amount:        int32(transaction.Amount),
		ChangeAmount:  int32(transaction.ChangeAmount),
		PaymentStatus: transaction.PaymentStatus,
		CreatedAt:     transaction.CreatedAt,
		UpdatedAt:     transaction.UpdatedAt,
	}
}

func (t *transactionProtoMapper) mapResponsesTransaction(transactions []*response.TransactionResponse) []*pbtransaction.TransactionResponse {
	var mappedTransactions []*pbtransaction.TransactionResponse

	for _, transaction := range transactions {
		mappedTransactions = append(mappedTransactions, t.mapResponseTransaction(transaction))
	}

	return mappedTransactions
}

func (t *transactionProtoMapper) mapResponseTransactionDeleteAt(transaction *response.TransactionResponseDeleteAt) *pbtransaction.TransactionResponseDeleteAt {
	var deletedAt *wrapperspb.StringValue
	if transaction.DeletedAt != nil {
		deletedAt = wrapperspb.String(*transaction.DeletedAt)
	}

	return &pbtransaction.TransactionResponseDeleteAt{
		Id:            int32(transaction.ID),
		OrderId:       int32(transaction.OrderID),
		MerchantId:    int32(transaction.MerchantID),
		PaymentMethod: transaction.PaymentMethod,
		Amount:        int32(transaction.Amount),
		ChangeAmount:  int32(transaction.ChangeAmount),
		PaymentStatus: transaction.PaymentStatus,
		CreatedAt:     transaction.CreatedAt,
		UpdatedAt:     transaction.UpdatedAt,
		DeletedAt:     deletedAt,
	}
}

func (t *transactionProtoMapper) mapResponsesTransactionDeleteAt(transactions []*response.TransactionResponseDeleteAt) []*pbtransaction.TransactionResponseDeleteAt {
	var mappedTransactions []*pbtransaction.TransactionResponseDeleteAt

	for _, transaction := range transactions {
		mappedTransactions = append(mappedTransactions, t.mapResponseTransactionDeleteAt(transaction))
	}

	return mappedTransactions
}
