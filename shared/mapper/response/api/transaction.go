package response_api

import (
	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
)

type transactionResponseMapper struct{}

func NewTransactionResponseMapper() *transactionResponseMapper {
	return &transactionResponseMapper{}
}

func (t *transactionResponseMapper) ToResponseTransaction(transaction *pbtransaction.TransactionResponse) *response.TransactionResponse {
	return &response.TransactionResponse{
		ID:            int(transaction.Id),
		OrderID:       int(transaction.OrderId),
		MerchantID:    int(transaction.MerchantId),
		PaymentMethod: transaction.PaymentMethod,
		Amount:        int(transaction.Amount),
		ChangeAmount:  int(transaction.ChangeAmount),
		PaymentStatus: transaction.PaymentStatus,
		CreatedAt:     transaction.CreatedAt,
		UpdatedAt:     transaction.UpdatedAt,
	}
}

func (t *transactionResponseMapper) ToResponsesTransaction(transactions []*pbtransaction.TransactionResponse) []*response.TransactionResponse {
	var mappedTransactions []*response.TransactionResponse

	for _, transaction := range transactions {
		mappedTransactions = append(mappedTransactions, t.ToResponseTransaction(transaction))
	}

	return mappedTransactions
}

func (t *transactionResponseMapper) ToResponseTransactionDeleteAt(transaction *pbtransaction.TransactionResponseDeleteAt) *response.TransactionResponseDeleteAt {
	var deletedAt string
	if transaction.DeletedAt != nil {
		deletedAt = transaction.DeletedAt.Value
	}

	return &response.TransactionResponseDeleteAt{
		ID:            int(transaction.Id),
		OrderID:       int(transaction.OrderId),
		MerchantID:    int(transaction.MerchantId),
		PaymentMethod: transaction.PaymentMethod,
		Amount:        int(transaction.Amount),
		ChangeAmount:  int(transaction.ChangeAmount),
		PaymentStatus: transaction.PaymentStatus,
		CreatedAt:     transaction.CreatedAt,
		UpdatedAt:     transaction.UpdatedAt,
		DeletedAt:     &deletedAt,
	}
}

func (t *transactionResponseMapper) ToResponsesTransactionDeleteAt(transactions []*pbtransaction.TransactionResponseDeleteAt) []*response.TransactionResponseDeleteAt {
	var mappedTransactions []*response.TransactionResponseDeleteAt

	for _, transaction := range transactions {
		mappedTransactions = append(mappedTransactions, t.ToResponseTransactionDeleteAt(transaction))
	}

	return mappedTransactions
}

func (t *transactionResponseMapper) ToApiResponseTransaction(pbResponse *pbtransaction.ApiResponseTransaction) *response.ApiResponseTransaction {
	return &response.ApiResponseTransaction{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    t.ToResponseTransaction(pbResponse.Data),
	}
}

func (t *transactionResponseMapper) ToApiResponseTransactionDeleteAt(pbResponse *pbtransaction.ApiResponseTransactionDeleteAt) *response.ApiResponseTransactionDeleteAt {
	return &response.ApiResponseTransactionDeleteAt{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    t.ToResponseTransactionDeleteAt(pbResponse.Data),
	}
}

func (t *transactionResponseMapper) ToApiResponsesTransaction(pbResponse *pbtransaction.ApiResponsesTransaction) *response.ApiResponsesTransaction {
	return &response.ApiResponsesTransaction{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    t.ToResponsesTransaction(pbResponse.Data),
	}
}

func (t *transactionResponseMapper) ToApiResponseTransactionDelete(pbResponse *pbtransaction.ApiResponseTransactionDelete) *response.ApiResponseTransactionDelete {
	return &response.ApiResponseTransactionDelete{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (t *transactionResponseMapper) ToApiResponseTransactionAll(pbResponse *pbtransaction.ApiResponseTransactionAll) *response.ApiResponseTransactionAll {
	return &response.ApiResponseTransactionAll{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (t *transactionResponseMapper) ToApiResponsePaginationTransactionDeleteAt(pbResponse *pbtransaction.ApiResponsePaginationTransactionDeleteAt) *response.ApiResponsePaginationTransactionDeleteAt {
	return &response.ApiResponsePaginationTransactionDeleteAt{
		Status:     pbResponse.Status,
		Message:    pbResponse.Message,
		Data:       t.ToResponsesTransactionDeleteAt(pbResponse.Data),
		Pagination: *mapPaginationMeta(pbResponse.Pagination),
	}
}

func (t *transactionResponseMapper) ToApiResponsePaginationTransaction(pbResponse *pbtransaction.ApiResponsePaginationTransaction) *response.ApiResponsePaginationTransaction {
	return &response.ApiResponsePaginationTransaction{
		Status:     pbResponse.Status,
		Message:    pbResponse.Message,
		Data:       t.ToResponsesTransaction(pbResponse.Data),
		Pagination: *mapPaginationMeta(pbResponse.Pagination),
	}
}
