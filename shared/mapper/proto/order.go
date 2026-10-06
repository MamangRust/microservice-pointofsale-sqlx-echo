package protomapper

import (
	pbcommon "github.com/MamangRust/microservice-point-of-sale-pb/common"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

type orderProtoMapper struct{}

func NewOrderProtoMapper() *orderProtoMapper {
	return &orderProtoMapper{}
}

func (o *orderProtoMapper) ToProtoResponseOrder(status string, message string, pbResponse *response.OrderResponse) *pborder.ApiResponseOrder {
	return &pborder.ApiResponseOrder{
		Status:  status,
		Message: message,
		Data:    o.mapResponseOrder(pbResponse),
	}
}

func (o *orderProtoMapper) ToProtoResponsesOrder(status string, message string, pbResponse []*response.OrderResponse) *pborder.ApiResponsesOrder {
	return &pborder.ApiResponsesOrder{
		Status:  status,
		Message: message,
		Data:    o.mapResponsesOrder(pbResponse),
	}
}

func (o *orderProtoMapper) ToProtoResponseOrderDeleteAt(status string, message string, pbResponse *response.OrderResponseDeleteAt) *pborder.ApiResponseOrderDeleteAt {
	return &pborder.ApiResponseOrderDeleteAt{
		Status:  status,
		Message: message,
		Data:    o.mapResponseOrderDeleteAt(pbResponse),
	}
}

func (o *orderProtoMapper) ToProtoResponseOrderDelete(status string, message string) *pborder.ApiResponseOrderDelete {
	return &pborder.ApiResponseOrderDelete{
		Status:  status,
		Message: message,
	}
}

func (o *orderProtoMapper) ToProtoResponseOrderAll(status string, message string) *pborder.ApiResponseOrderAll {
	return &pborder.ApiResponseOrderAll{
		Status:  status,
		Message: message,
	}
}

func (o *orderProtoMapper) ToProtoResponsePaginationOrderDeleteAt(pagination *pbcommon.PaginationMeta, status string, message string, orders []*response.OrderResponseDeleteAt) *pborder.ApiResponsePaginationOrderDeleteAt {
	return &pborder.ApiResponsePaginationOrderDeleteAt{
		Status:     status,
		Message:    message,
		Data:       o.mapResponsesOrderDeleteAt(orders),
		Pagination: mapPaginationMeta(pagination),
	}
}

func (o *orderProtoMapper) ToProtoResponsePaginationOrder(pagination *pbcommon.PaginationMeta, status string, message string, orders []*response.OrderResponse) *pborder.ApiResponsePaginationOrder {
	return &pborder.ApiResponsePaginationOrder{
		Status:     status,
		Message:    message,
		Data:       o.mapResponsesOrder(orders),
		Pagination: mapPaginationMeta(pagination),
	}
}

func (o *orderProtoMapper) mapResponseOrder(order *response.OrderResponse) *pborder.OrderResponse {
	return &pborder.OrderResponse{
		Id:         int32(order.ID),
		MerchantId: int32(order.MerchantID),
		CashierId:  int32(order.CashierID),
		TotalPrice: int32(order.TotalPrice),
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
	}
}

func (o *orderProtoMapper) mapResponsesOrder(orders []*response.OrderResponse) []*pborder.OrderResponse {
	var mappedOrders []*pborder.OrderResponse

	for _, order := range orders {
		mappedOrders = append(mappedOrders, o.mapResponseOrder(order))
	}

	return mappedOrders
}

func (o *orderProtoMapper) mapResponseOrderDeleteAt(order *response.OrderResponseDeleteAt) *pborder.OrderResponseDeleteAt {
	var deletedAt *wrapperspb.StringValue

	if order.DeleteAt != nil {
		deletedAt = wrapperspb.String(*order.DeleteAt)
	}

	return &pborder.OrderResponseDeleteAt{
		Id:         int32(order.ID),
		MerchantId: int32(order.MerchantID),
		CashierId:  int32(order.CashierID),
		TotalPrice: int32(order.TotalPrice),
		CreatedAt:  order.CreatedAt,
		UpdatedAt:  order.UpdatedAt,
		DeletedAt:  deletedAt,
	}
}

func (o *orderProtoMapper) mapResponsesOrderDeleteAt(orders []*response.OrderResponseDeleteAt) []*pborder.OrderResponseDeleteAt {
	var mappedOrders []*pborder.OrderResponseDeleteAt

	for _, order := range orders {
		mappedOrders = append(mappedOrders, o.mapResponseOrderDeleteAt(order))
	}

	return mappedOrders
}
