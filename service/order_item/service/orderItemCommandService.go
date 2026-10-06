package service

import (
	"context"

	mencache "github.com/MamangRust/microservice-point-of-sale-order-item/cache"
	db "github.com/MamangRust/microservice-point-of-sale-order-item/database/schema"
	"github.com/MamangRust/microservice-point-of-sale-order-item/repository"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	sharederrorhandler "github.com/MamangRust/microservice-point-of-sale-shared/errorhandler"
	orderitem_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/order_item_errors"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	"go.opentelemetry.io/otel/attribute"
	"go.uber.org/zap"
)

type orderItemCommandDeps struct {
	Cache         mencache.OrderItemCommandCache
	Repo          repository.OrderItemCommandRepository
	Logger        logger.LoggerInterface
	Observability observability.TraceLoggerObservability
}

type orderItemCommandService struct {
	mencache      mencache.OrderItemCommandCache
	repo          repository.OrderItemCommandRepository
	logger        logger.LoggerInterface
	observability observability.TraceLoggerObservability
}

func NewOrderItemCommandService(params *orderItemCommandDeps) OrderItemCommandService {
	return &orderItemCommandService{
		mencache:      params.Cache,
		repo:          params.Repo,
		logger:        params.Logger,
		observability: params.Observability,
	}
}

func (s *orderItemCommandService) CreateOrderItem(ctx context.Context, req *requests.CreateOrderItemRecordRequest) (*db.OrderItem, error) {
	const method = "CreateOrderItem"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("order.id", req.OrderID),
		attribute.Int("product.id", req.ProductID),
		attribute.Int("quantity", req.Quantity),
	)
	defer func() {
		end(status)
	}()

	orderItem, err := s.repo.CreateOrderItem(ctx, req)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*db.OrderItem](
			s.logger,
			orderitem_errors.ErrFailedCreateOrderItem.WithInternal(err),
			method,
			span,
			zap.Error(err),
		)
	}

	s.mencache.DeleteOrderItemCache(ctx, int(orderItem.OrderID))
	logSuccess("Successfully created order item", zap.Int32("order_item.id", orderItem.OrderItemID))
	return orderItem, nil
}

func (s *orderItemCommandService) UpdateOrderItem(ctx context.Context, req *requests.UpdateOrderItemRecordRequest) (*db.OrderItem, error) {
	const method = "UpdateOrderItem"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("order_item.id", req.OrderItemID),
		attribute.Int("order.id", req.OrderID),
	)
	defer func() {
		end(status)
	}()

	orderItem, err := s.repo.UpdateOrderItem(ctx, req)
	if err != nil {
		status = "error"
		return sharederrorhandler.HandleError[*db.OrderItem](
			s.logger,
			orderitem_errors.ErrFailedUpdateOrderItem.WithInternal(err),
			method,
			span,
			zap.Error(err),
		)
	}

	s.mencache.DeleteOrderItemCache(ctx, int(orderItem.OrderID))
	logSuccess("Successfully updated order item", zap.Int32("order_item.id", orderItem.OrderItemID))
	return orderItem, nil
}

func (s *orderItemCommandService) DeleteOrderItem(ctx context.Context, orderItemID int) error {
	const method = "DeleteOrderItem"

	ctx, span, end, status, logSuccess := s.observability.StartTracingAndLogging(ctx, method,
		attribute.Int("order_item.id", orderItemID),
	)
	defer func() {
		end(status)
	}()

	if err := s.repo.DeleteOrderItem(ctx, orderItemID); err != nil {
		status = "error"
		_, mappedErr := sharederrorhandler.HandleError[any](
			s.logger,
			orderitem_errors.ErrFailedDeleteOrderItem.WithInternal(err),
			method,
			span,
			zap.Error(err),
		)
		return mappedErr
	}

	s.mencache.DeleteOrderItemAllCache(ctx)
	logSuccess("Successfully deleted order item", zap.Int("order_item.id", orderItemID))
	return nil
}
