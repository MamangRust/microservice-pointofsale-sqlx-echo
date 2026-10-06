package handler

import (
	"context"

	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	orderitem_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/order_item_errors"
	"go.uber.org/zap"
)

func (s *orderItemHandleGrpc) CreateOrderItem(ctx context.Context, request *pborderitem.CreateOrderItemRecordRequest) (*pborderitem.ApiResponseOrderItem, error) {
	s.logger.Info("CreateOrderItem called",
		zap.Int32("order.id", request.GetOrderId()),
		zap.Int32("product.id", request.GetProductId()),
	)

	if request.GetOrderId() <= 0 || request.GetProductId() <= 0 {
		return nil, orderitem_errors.ErrGrpcInvalidID
	}
	if request.GetQuantity() <= 0 {
		return nil, orderitem_errors.ErrFailedInvalidQuantity
	}

	req := &requests.CreateOrderItemRecordRequest{
		OrderID:   int(request.GetOrderId()),
		ProductID: int(request.GetProductId()),
		Quantity:  int(request.GetQuantity()),
		Price:     int(request.GetPrice()),
	}

	orderItem, err := s.orderItemCommandService.CreateOrderItem(ctx, req)
	if err != nil {
		s.logger.Error("CreateOrderItem failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("CreateOrderItem success", zap.Int32("order_item.id", orderItem.OrderItemID))

	return &pborderitem.ApiResponseOrderItem{
		Status:  "success",
		Message: "Successfully created order item",
		Data:    mapOrderItemToProto(orderItem),
	}, nil
}

func (s *orderItemHandleGrpc) UpdateOrderItem(ctx context.Context, request *pborderitem.UpdateOrderItemRecordRequest) (*pborderitem.ApiResponseOrderItem, error) {
	s.logger.Info("UpdateOrderItem called", zap.Int32("order_item.id", request.GetOrderItemId()))

	if request.GetOrderItemId() <= 0 {
		return nil, orderitem_errors.ErrGrpcInvalidID
	}
	if request.GetQuantity() <= 0 {
		return nil, orderitem_errors.ErrFailedInvalidQuantity
	}

	req := &requests.UpdateOrderItemRecordRequest{
		OrderItemID: int(request.GetOrderItemId()),
		OrderID:     int(request.GetOrderId()),
		ProductID:   int(request.GetProductId()),
		Quantity:    int(request.GetQuantity()),
		Price:       int(request.GetPrice()),
	}

	orderItem, err := s.orderItemCommandService.UpdateOrderItem(ctx, req)
	if err != nil {
		s.logger.Error("UpdateOrderItem failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("UpdateOrderItem success", zap.Int32("order_item.id", orderItem.OrderItemID))

	return &pborderitem.ApiResponseOrderItem{
		Status:  "success",
		Message: "Successfully updated order item",
		Data:    mapOrderItemToProto(orderItem),
	}, nil
}

func (s *orderItemHandleGrpc) DeleteOrderItem(ctx context.Context, request *pborderitem.DeleteOrderItemRecordRequest) (*pborderitem.ApiResponseOrderItemDelete, error) {
	s.logger.Info("DeleteOrderItem called", zap.Int32("order_item.id", request.GetOrderItemId()))

	if request.GetOrderItemId() <= 0 {
		return nil, orderitem_errors.ErrGrpcInvalidID
	}

	if err := s.orderItemCommandService.DeleteOrderItem(ctx, int(request.GetOrderItemId())); err != nil {
		s.logger.Error("DeleteOrderItem failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	s.logger.Info("DeleteOrderItem success", zap.Int32("order_item.id", request.GetOrderItemId()))

	return &pborderitem.ApiResponseOrderItemDelete{
		Status:  "success",
		Message: "Successfully deleted order item",
	}, nil
}
