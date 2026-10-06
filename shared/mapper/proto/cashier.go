package protomapper

import (
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbcommon "github.com/MamangRust/microservice-point-of-sale-pb/common"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

type cashierProtoMapper struct {
}

func NewCashierProtoMapper() *cashierProtoMapper {
	return &cashierProtoMapper{}
}

func (c *cashierProtoMapper) ToProtoResponseCashier(status string, message string, pbResponse *response.CashierResponse) *pbcashier.ApiResponseCashier {
	return &pbcashier.ApiResponseCashier{
		Status:  status,
		Message: message,
		Data:    c.mapResponseCashier(pbResponse),
	}
}

func (c *cashierProtoMapper) ToProtoResponsesCashier(status string, message string, pbResponse []*response.CashierResponse) *pbcashier.ApiResponsesCashier {
	return &pbcashier.ApiResponsesCashier{
		Status:  status,
		Message: message,
		Data:    c.mapResponsesCashier(pbResponse),
	}
}

func (c *cashierProtoMapper) ToProtoResponseCashierDeleteAt(status string, message string, pbResponse *response.CashierResponseDeleteAt) *pbcashier.ApiResponseCashierDeleteAt {
	return &pbcashier.ApiResponseCashierDeleteAt{
		Status:  status,
		Message: message,
		Data:    c.mapResponseCashierDeleteAt(pbResponse),
	}
}

func (c *cashierProtoMapper) ToProtoResponseCashierDelete(status string, message string) *pbcashier.ApiResponseCashierDelete {
	return &pbcashier.ApiResponseCashierDelete{
		Status:  status,
		Message: message,
	}
}

func (u *cashierProtoMapper) ToProtoResponseCashierAll(status string, message string) *pbcashier.ApiResponseCashierAll {
	return &pbcashier.ApiResponseCashierAll{
		Status:  status,
		Message: message,
	}
}

func (u *cashierProtoMapper) ToProtoResponsePaginationCashierDeleteAt(pagination *pbcommon.PaginationMeta, status string, message string, users []*response.CashierResponseDeleteAt) *pbcashier.ApiResponsePaginationCashierDeleteAt {
	return &pbcashier.ApiResponsePaginationCashierDeleteAt{
		Status:     status,
		Message:    message,
		Data:       u.mapResponsesCashierDeleteAt(users),
		Pagination: mapPaginationMeta(pagination),
	}
}

func (u *cashierProtoMapper) ToProtoResponsePaginationCashier(pagination *pbcommon.PaginationMeta, status string, message string, users []*response.CashierResponse) *pbcashier.ApiResponsePaginationCashier {
	return &pbcashier.ApiResponsePaginationCashier{
		Status:     status,
		Message:    message,
		Data:       u.mapResponsesCashier(users),
		Pagination: mapPaginationMeta(pagination),
	}
}

func (c *cashierProtoMapper) mapResponseCashier(cashier *response.CashierResponse) *pbcashier.CashierResponse {
	return &pbcashier.CashierResponse{
		Id:         int32(cashier.ID),
		MerchantId: int32(cashier.MerchantID),
		Name:       cashier.Name,
		CreatedAt:  cashier.CreatedAt,
		UpdatedAt:  cashier.UpdatedAt,
	}
}

func (c *cashierProtoMapper) mapResponsesCashier(cashiers []*response.CashierResponse) []*pbcashier.CashierResponse {
	var mappedCashiers []*pbcashier.CashierResponse

	for _, cashier := range cashiers {
		mappedCashiers = append(mappedCashiers, c.mapResponseCashier(cashier))
	}

	return mappedCashiers
}

func (c *cashierProtoMapper) mapResponseCashierDeleteAt(cashier *response.CashierResponseDeleteAt) *pbcashier.CashierResponseDeleteAt {
	var deletedAt *wrapperspb.StringValue
	if cashier.DeletedAt != nil {
		deletedAt = wrapperspb.String(*cashier.DeletedAt)
	}

	return &pbcashier.CashierResponseDeleteAt{
		Id:         int32(cashier.ID),
		MerchantId: int32(cashier.MerchantID),
		Name:       cashier.Name,
		CreatedAt:  cashier.CreatedAt,
		UpdatedAt:  cashier.UpdatedAt,
		DeletedAt:  deletedAt,
	}
}

func (c *cashierProtoMapper) mapResponsesCashierDeleteAt(cashiers []*response.CashierResponseDeleteAt) []*pbcashier.CashierResponseDeleteAt {
	var mappedCashiers []*pbcashier.CashierResponseDeleteAt

	for _, cashier := range cashiers {
		mappedCashiers = append(mappedCashiers, c.mapResponseCashierDeleteAt(cashier))
	}

	return mappedCashiers
}
