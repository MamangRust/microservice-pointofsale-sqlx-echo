package response_api

import (
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/response"
)

type orderItemResponseMapper struct {
}

func NewOrderItemResponseMapper() *orderItemResponseMapper {
	return &orderItemResponseMapper{}
}

func (o *orderItemResponseMapper) ToResponseOrderItem(orderItem *pborderitem.OrderItemResponse) *response.OrderItemResponse {
	return &response.OrderItemResponse{
		ID:        int(orderItem.Id),
		OrderID:   int(orderItem.OrderId),
		ProductID: int(orderItem.ProductId),
		Quantity:  int(orderItem.Quantity),
		Price:     int(orderItem.Price),
		CreatedAt: orderItem.CreatedAt,
		UpdatedAt: orderItem.UpdatedAt,
	}
}

func (o *orderItemResponseMapper) ToResponsesOrderItem(orderItems []*pborderitem.OrderItemResponse) []*response.OrderItemResponse {
	var mappedOrderItems []*response.OrderItemResponse

	for _, orderItem := range orderItems {
		mappedOrderItems = append(mappedOrderItems, o.ToResponseOrderItem(orderItem))
	}

	return mappedOrderItems
}

func (o *orderItemResponseMapper) ToResponseOrderItemDeleteAt(orderItem *pborderitem.OrderItemResponseDeleteAt) *response.OrderItemResponseDeleteAt {
	var deletedAt string
	if orderItem.DeletedAt != nil {
		deletedAt = orderItem.DeletedAt.Value
	}
	return &response.OrderItemResponseDeleteAt{
		ID:        int(orderItem.Id),
		OrderID:   int(orderItem.OrderId),
		ProductID: int(orderItem.ProductId),
		Quantity:  int(orderItem.Quantity),
		Price:     int(orderItem.Price),
		CreatedAt: orderItem.CreatedAt,
		UpdatedAt: orderItem.UpdatedAt,
		DeleteAt:  &deletedAt,
	}
}

func (o *orderItemResponseMapper) ToResponsesOrderItemDeleteAt(orderItems []*pborderitem.OrderItemResponseDeleteAt) []*response.OrderItemResponseDeleteAt {
	var mappedOrderItems []*response.OrderItemResponseDeleteAt

	for _, orderItem := range orderItems {
		mappedOrderItems = append(mappedOrderItems, o.ToResponseOrderItemDeleteAt(orderItem))
	}

	return mappedOrderItems
}

func (o *orderItemResponseMapper) ToApiResponseOrderItem(pbResponse *pborderitem.ApiResponseOrderItem) *response.ApiResponseOrderItem {
	return &response.ApiResponseOrderItem{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    o.ToResponseOrderItem(pbResponse.Data),
	}
}

func (o *orderItemResponseMapper) ToApiResponsesOrderItem(pbResponse *pborderitem.ApiResponsesOrderItem) *response.ApiResponsesOrderItem {
	return &response.ApiResponsesOrderItem{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
		Data:    o.ToResponsesOrderItem(pbResponse.Data),
	}
}

func (o *orderItemResponseMapper) ToApiResponseOrderItemDelete(pbResponse *pborderitem.ApiResponseOrderItemDelete) *response.ApiResponseOrderItemDelete {
	return &response.ApiResponseOrderItemDelete{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (o *orderItemResponseMapper) ToApiResponseOrderItemAll(pbResponse *pborderitem.ApiResponseOrderItemAll) *response.ApiResponseOrderItemAll {
	return &response.ApiResponseOrderItemAll{
		Status:  pbResponse.Status,
		Message: pbResponse.Message,
	}
}

func (o *orderItemResponseMapper) ToApiResponsePaginationOrderItemDeleteAt(pbResponse *pborderitem.ApiResponsePaginationOrderItemDeleteAt) *response.ApiResponsePaginationOrderItemDeleteAt {
	return &response.ApiResponsePaginationOrderItemDeleteAt{
		Status:     pbResponse.Status,
		Message:    pbResponse.Message,
		Data:       o.ToResponsesOrderItemDeleteAt(pbResponse.Data),
		Pagination: *mapPaginationMeta(pbResponse.Pagination),
	}
}

func (o *orderItemResponseMapper) ToApiResponsePaginationOrderItem(pbResponse *pborderitem.ApiResponsePaginationOrderItem) *response.ApiResponsePaginationOrderItem {
	return &response.ApiResponsePaginationOrderItem{
		Status:     pbResponse.Status,
		Message:    pbResponse.Message,
		Data:       o.ToResponsesOrderItem(pbResponse.Data),
		Pagination: *mapPaginationMeta(pbResponse.Pagination),
	}
}
