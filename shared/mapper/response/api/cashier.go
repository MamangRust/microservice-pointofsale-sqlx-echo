package response_api

import (
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
)

type cashierResponseMapper struct{}

func NewCashierResponseMapper() *cashierResponseMapper {
	return &cashierResponseMapper{}
}

func (c *cashierResponseMapper) ToResponseCashier(cashier *pbcashier.CashierResponse) *response.CashierResponse {
	return &response.CashierResponse{
		ID:         int(cashier.Id),
		MerchantID: int(cashier.MerchantId),
		Name:       cashier.Name,
		CreatedAt:  cashier.CreatedAt,
		UpdatedAt:  cashier.UpdatedAt,
	}
}

func (c *cashierResponseMapper) ToResponsesCashier(cashiers []*pbcashier.CashierResponse) []*response.CashierResponse {
	var mappedCashiers []*response.CashierResponse

	for _, cashier := range cashiers {
		mappedCashiers = append(mappedCashiers, c.ToResponseCashier(cashier))
	}

	return mappedCashiers
}

func (c *cashierResponseMapper) ToResponseCashierDeleteAt(cashier *pbcashier.CashierResponseDeleteAt) *response.CashierResponseDeleteAt {
	var deletedAt string
	if cashier.DeletedAt != nil {
		deletedAt = cashier.DeletedAt.Value
	}

	return &response.CashierResponseDeleteAt{
		ID:         int(cashier.Id),
		MerchantID: int(cashier.MerchantId),
		Name:       cashier.Name,
		CreatedAt:  cashier.CreatedAt,
		UpdatedAt:  cashier.UpdatedAt,
		DeletedAt:  &deletedAt,
	}
}

func (c *cashierResponseMapper) ToResponsesCashierDeleteAt(cashiers []*pbcashier.CashierResponseDeleteAt) []*response.CashierResponseDeleteAt {
	var mappedCashiers []*response.CashierResponseDeleteAt

	for _, cashier := range cashiers {
		mappedCashiers = append(mappedCashiers, c.ToResponseCashierDeleteAt(cashier))
	}

	return mappedCashiers
}

func (c *cashierResponseMapper) ToApiResponseCashier(pbResponse *pbcashier.ApiResponseCashier) *response.ApiResponseCashier {
	return &response.ApiResponseCashier{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    c.ToResponseCashier(pbResponse.Data),
	}
}

func (c *cashierResponseMapper) ToApiResponsesCashier(pbResponse *pbcashier.ApiResponsesCashier) *response.ApiResponsesCashier {
	return &response.ApiResponsesCashier{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    c.ToResponsesCashier(pbResponse.Data),
	}
}

func (c *cashierResponseMapper) ToApiResponseCashierDeleteAt(pbResponse *pbcashier.ApiResponseCashierDeleteAt) *response.ApiResponseCashierDeleteAt {
	return &response.ApiResponseCashierDeleteAt{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    c.ToResponseCashierDeleteAt(pbResponse.Data),
	}
}

func (c *cashierResponseMapper) ToApiResponseCashierDelete(pbResponse *pbcashier.ApiResponseCashierDelete) *response.ApiResponseCashierDelete {
	return &response.ApiResponseCashierDelete{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (c *cashierResponseMapper) ToApiResponseCashierAll(pbResponse *pbcashier.ApiResponseCashierAll) *response.ApiResponseCashierAll {
	return &response.ApiResponseCashierAll{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (c *cashierResponseMapper) ToApiResponsePaginationCashierDeleteAt(pbResponse *pbcashier.ApiResponsePaginationCashierDeleteAt) *response.ApiResponsePaginationCashierDeleteAt {
	return &response.ApiResponsePaginationCashierDeleteAt{
		Status:     pbResponse.Status,
		Message:    pbResponse.Message,
		Data:       c.ToResponsesCashierDeleteAt(pbResponse.Data),
		Pagination: *mapPaginationMeta(pbResponse.Pagination),
	}
}

func (c *cashierResponseMapper) ToApiResponsePaginationCashier(pbResponse *pbcashier.ApiResponsePaginationCashier) *response.ApiResponsePaginationCashier {
	return &response.ApiResponsePaginationCashier{
		Status:     pbResponse.Status,
		Message:    pbResponse.Message,
		Data:       c.ToResponsesCashier(pbResponse.Data),
		Pagination: *mapPaginationMeta(pbResponse.Pagination),
	}
}
