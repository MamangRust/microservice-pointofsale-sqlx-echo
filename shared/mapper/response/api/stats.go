package response_api

import (
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pbtransaction "github.com/MamangRust/microservice-point-of-sale-pb/transaction"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
)

// StatsResponseMapper maps the stats-reader (ClickHouse) gRPC responses to the
// REST shapes exposed by the API gateway (F4 §7.4). The REST shapes are shared
// with the domain services' own stats endpoints (domain/response).
type StatsResponseMapper interface {
	ToApiResponseOrderMonthlyTotalRevenue(res *pborder.ApiResponseOrderMonthlyTotalRevenue) *response.ApiResponseOrderMonthlyTotalRevenue
	ToApiResponseOrderYearlyTotalRevenue(res *pborder.ApiResponseOrderYearlyTotalRevenue) *response.ApiResponseOrderYearlyTotalRevenue
	ToApiResponseOrderMonthly(res *pborder.ApiResponseOrderMonthly) *response.ApiResponseOrderMonthly
	ToApiResponseOrderYearly(res *pborder.ApiResponseOrderYearly) *response.ApiResponseOrderYearly

	ToApiResponseCashierMonthlyTotalSales(res *pbcashier.ApiResponseCashierMonthlyTotalSales) *response.ApiResponseCashierMonthlyTotalSales
	ToApiResponseCashierYearlyTotalSales(res *pbcashier.ApiResponseCashierYearlyTotalSales) *response.ApiResponseCashierYearlyTotalSales
	ToApiResponseCashierMonthSales(res *pbcashier.ApiResponseCashierMonthSales) *response.ApiResponseCashierMonthSales
	ToApiResponseCashierYearSales(res *pbcashier.ApiResponseCashierYearSales) *response.ApiResponseCashierYearSales

	ToApiResponseCategoryMonthlyTotalPrice(res *pbcategory.ApiResponseCategoryMonthlyTotalPrice) *response.ApiResponseCategoryMonthlyTotalPrice
	ToApiResponseCategoryYearlyTotalPrice(res *pbcategory.ApiResponseCategoryYearlyTotalPrice) *response.ApiResponseCategoryYearlyTotalPrice
	ToApiResponseCategoryMonthPrice(res *pbcategory.ApiResponseCategoryMonthPrice) *response.ApiResponseCategoryMonthPrice
	ToApiResponseCategoryYearPrice(res *pbcategory.ApiResponseCategoryYearPrice) *response.ApiResponseCategoryYearPrice

	ToApiResponsesTransactionMonthSuccess(res *pbtransaction.ApiResponseTransactionMonthAmountSuccess) *response.ApiResponsesTransactionMonthSuccess
	ToApiResponsesTransactionYearSuccess(res *pbtransaction.ApiResponseTransactionYearAmountSuccess) *response.ApiResponsesTransactionYearSuccess
	ToApiResponsesTransactionMonthFailed(res *pbtransaction.ApiResponseTransactionMonthAmountFailed) *response.ApiResponsesTransactionMonthFailed
	ToApiResponsesTransactionYearFailed(res *pbtransaction.ApiResponseTransactionYearAmountFailed) *response.ApiResponsesTransactionYearFailed

	ToApiResponsesTransactionMonthMethod(res *pbtransaction.ApiResponseTransactionMonthPaymentMethod) *response.ApiResponsesTransactionMonthMethod
	ToApiResponsesTransactionYearMethod(res *pbtransaction.ApiResponseTransactionYearPaymentmethod) *response.ApiResponsesTransactionYearMethod
}

type statsResponseMapper struct{}

func NewStatsResponseMapper() *statsResponseMapper {
	return &statsResponseMapper{}
}

// --- Order stats ---

func (s *statsResponseMapper) ToApiResponseOrderMonthlyTotalRevenue(res *pborder.ApiResponseOrderMonthlyTotalRevenue) *response.ApiResponseOrderMonthlyTotalRevenue {
	data := make([]*response.OrderMonthlyTotalRevenueResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.OrderMonthlyTotalRevenueResponse{
			Year:           row.GetYear(),
			Month:          row.GetMonth(),
			TotalRevenue:   int(row.GetTotalRevenue()),
			TotalItemsSold: int(row.GetTotalItemsSold()),
		})
	}
	return &response.ApiResponseOrderMonthlyTotalRevenue{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseOrderYearlyTotalRevenue(res *pborder.ApiResponseOrderYearlyTotalRevenue) *response.ApiResponseOrderYearlyTotalRevenue {
	data := make([]*response.OrderYearlyTotalRevenueResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.OrderYearlyTotalRevenueResponse{
			Year:         row.GetYear(),
			TotalRevenue: int(row.GetTotalRevenue()),
		})
	}
	return &response.ApiResponseOrderYearlyTotalRevenue{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseOrderMonthly(res *pborder.ApiResponseOrderMonthly) *response.ApiResponseOrderMonthly {
	data := make([]*response.OrderMonthlyResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.OrderMonthlyResponse{
			Month:          row.GetMonth(),
			OrderCount:     int(row.GetOrderCount()),
			TotalRevenue:   int(row.GetTotalRevenue()),
			TotalItemsSold: int(row.GetTotalItemsSold()),
		})
	}
	return &response.ApiResponseOrderMonthly{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseOrderYearly(res *pborder.ApiResponseOrderYearly) *response.ApiResponseOrderYearly {
	data := make([]*response.OrderYearlyResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.OrderYearlyResponse{
			Year:               row.GetYear(),
			OrderCount:         int(row.GetOrderCount()),
			TotalRevenue:       int(row.GetTotalRevenue()),
			TotalItemsSold:     int(row.GetTotalItemsSold()),
			ActiveCashiers:     int(row.GetActiveCashiers()),
			UniqueProductsSold: int(row.GetUniqueProductsSold()),
		})
	}
	return &response.ApiResponseOrderYearly{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

// --- Cashier stats ---

func (s *statsResponseMapper) ToApiResponseCashierMonthlyTotalSales(res *pbcashier.ApiResponseCashierMonthlyTotalSales) *response.ApiResponseCashierMonthlyTotalSales {
	data := make([]*response.CashierResponseMonthTotalSales, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.CashierResponseMonthTotalSales{
			Year:       row.GetYear(),
			Month:      row.GetMonth(),
			TotalSales: int(row.GetTotalSales()),
		})
	}
	return &response.ApiResponseCashierMonthlyTotalSales{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseCashierYearlyTotalSales(res *pbcashier.ApiResponseCashierYearlyTotalSales) *response.ApiResponseCashierYearlyTotalSales {
	data := make([]*response.CashierResponseYearTotalSales, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.CashierResponseYearTotalSales{
			Year:       row.GetYear(),
			TotalSales: int(row.GetTotalSales()),
		})
	}
	return &response.ApiResponseCashierYearlyTotalSales{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseCashierMonthSales(res *pbcashier.ApiResponseCashierMonthSales) *response.ApiResponseCashierMonthSales {
	data := make([]*response.CashierResponseMonthSales, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.CashierResponseMonthSales{
			Month:      row.GetMonth(),
			CashierID:  int(row.GetCashierId()),
			OrderCount: int(row.GetOrderCount()),
			TotalSales: int(row.GetTotalSales()),
		})
	}
	return &response.ApiResponseCashierMonthSales{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseCashierYearSales(res *pbcashier.ApiResponseCashierYearSales) *response.ApiResponseCashierYearSales {
	data := make([]*response.CashierResponseYearSales, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.CashierResponseYearSales{
			Year:       row.GetYear(),
			CashierID:  int(row.GetCashierId()),
			OrderCount: int(row.GetOrderCount()),
			TotalSales: int(row.GetTotalSales()),
		})
	}
	return &response.ApiResponseCashierYearSales{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

// --- Category stats ---

func (s *statsResponseMapper) ToApiResponseCategoryMonthlyTotalPrice(res *pbcategory.ApiResponseCategoryMonthlyTotalPrice) *response.ApiResponseCategoryMonthlyTotalPrice {
	data := make([]*response.CategoriesMonthlyTotalPriceResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.CategoriesMonthlyTotalPriceResponse{
			Year:         row.GetYear(),
			Month:        row.GetMonth(),
			TotalRevenue: int(row.GetTotalRevenue()),
		})
	}
	return &response.ApiResponseCategoryMonthlyTotalPrice{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseCategoryYearlyTotalPrice(res *pbcategory.ApiResponseCategoryYearlyTotalPrice) *response.ApiResponseCategoryYearlyTotalPrice {
	data := make([]*response.CategoriesYearlyTotalPriceResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.CategoriesYearlyTotalPriceResponse{
			Year:         row.GetYear(),
			TotalRevenue: int(row.GetTotalRevenue()),
		})
	}
	return &response.ApiResponseCategoryYearlyTotalPrice{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseCategoryMonthPrice(res *pbcategory.ApiResponseCategoryMonthPrice) *response.ApiResponseCategoryMonthPrice {
	data := make([]*response.CategoryMonthPriceResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.CategoryMonthPriceResponse{
			Month:        row.GetMonth(),
			CategoryID:   int(row.GetCategoryId()),
			OrderCount:   int(row.GetOrderCount()),
			ItemsSold:    int(row.GetItemsSold()),
			TotalRevenue: int(row.GetTotalRevenue()),
		})
	}
	return &response.ApiResponseCategoryMonthPrice{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponseCategoryYearPrice(res *pbcategory.ApiResponseCategoryYearPrice) *response.ApiResponseCategoryYearPrice {
	data := make([]*response.CategoryYearPriceResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.CategoryYearPriceResponse{
			Year:               row.GetYear(),
			CategoryID:         int(row.GetCategoryId()),
			OrderCount:         int(row.GetOrderCount()),
			ItemsSold:          int(row.GetItemsSold()),
			TotalRevenue:       int(row.GetTotalRevenue()),
			UniqueProductsSold: int(row.GetUniqueProductsSold()),
		})
	}
	return &response.ApiResponseCategoryYearPrice{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

// --- Transaction stats ---

func (s *statsResponseMapper) ToApiResponsesTransactionMonthSuccess(res *pbtransaction.ApiResponseTransactionMonthAmountSuccess) *response.ApiResponsesTransactionMonthSuccess {
	data := make([]*response.TransactionMonthlyAmountSuccessResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.TransactionMonthlyAmountSuccessResponse{
			Year:         row.GetYear(),
			Month:        row.GetMonth(),
			TotalSuccess: int(row.GetTotalSuccess()),
			TotalAmount:  int(row.GetTotalAmount()),
		})
	}
	return &response.ApiResponsesTransactionMonthSuccess{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponsesTransactionYearSuccess(res *pbtransaction.ApiResponseTransactionYearAmountSuccess) *response.ApiResponsesTransactionYearSuccess {
	data := make([]*response.TransactionYearlyAmountSuccessResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.TransactionYearlyAmountSuccessResponse{
			Year:         row.GetYear(),
			TotalSuccess: int(row.GetTotalSuccess()),
			TotalAmount:  int(row.GetTotalAmount()),
		})
	}
	return &response.ApiResponsesTransactionYearSuccess{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponsesTransactionMonthFailed(res *pbtransaction.ApiResponseTransactionMonthAmountFailed) *response.ApiResponsesTransactionMonthFailed {
	data := make([]*response.TransactionMonthlyAmountFailedResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.TransactionMonthlyAmountFailedResponse{
			Year:        row.GetYear(),
			Month:       row.GetMonth(),
			TotalFailed: int(row.GetTotalFailed()),
			TotalAmount: int(row.GetTotalAmount()),
		})
	}
	return &response.ApiResponsesTransactionMonthFailed{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponsesTransactionYearFailed(res *pbtransaction.ApiResponseTransactionYearAmountFailed) *response.ApiResponsesTransactionYearFailed {
	data := make([]*response.TransactionYearlyAmountFailedResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.TransactionYearlyAmountFailedResponse{
			Year:        row.GetYear(),
			TotalFailed: int(row.GetTotalFailed()),
			TotalAmount: int(row.GetTotalAmount()),
		})
	}
	return &response.ApiResponsesTransactionYearFailed{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponsesTransactionMonthMethod(res *pbtransaction.ApiResponseTransactionMonthPaymentMethod) *response.ApiResponsesTransactionMonthMethod {
	data := make([]*response.TransactionMonthlyMethodResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.TransactionMonthlyMethodResponse{
			Month:             row.GetMonth(),
			PaymentMethod:     row.GetPaymentMethod(),
			TotalTransactions: int(row.GetTotalTransactions()),
			TotalAmount:       int(row.GetTotalAmount()),
		})
	}
	return &response.ApiResponsesTransactionMonthMethod{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}

func (s *statsResponseMapper) ToApiResponsesTransactionYearMethod(res *pbtransaction.ApiResponseTransactionYearPaymentmethod) *response.ApiResponsesTransactionYearMethod {
	data := make([]*response.TransactionYearlyMethodResponse, 0, len(res.GetData()))
	for _, row := range res.GetData() {
		data = append(data, &response.TransactionYearlyMethodResponse{
			Year:              row.GetYear(),
			PaymentMethod:     row.GetPaymentMethod(),
			TotalTransactions: int(row.GetTotalTransactions()),
			TotalAmount:       int(row.GetTotalAmount()),
		})
	}
	return &response.ApiResponsesTransactionYearMethod{Status: res.GetStatus(), Message: res.GetMessage(), Data: data}
}
