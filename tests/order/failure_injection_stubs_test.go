package order_test

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

// stubCashierRepo implements cashieradapter.QueryRepository. In the
// failure-injection suite it is never reached (the merchant dependency fails
// first), so it returns a minimal valid model to keep the flow honest if the
// ordering ever changes.
type stubCashierRepo struct {
	cashierID int
}

func (s *stubCashierRepo) FindById(ctx context.Context, cashierID int) (*models.Cashier, error) {
	return &models.Cashier{CashierID: int32(cashierID), MerchantID: int32(s.cashierID), Name: "stub"}, nil
}

// stubProductRepo implements productadapter.QueryRepository and
// productadapter.CommandRepository. Never reached in the failure-injection
// suite (merchant fails first).
type stubProductRepo struct{}

func (s *stubProductRepo) FindById(ctx context.Context, productID int) (*models.Product, error) {
	return &models.Product{ProductID: int32(productID), Price: 10000, CountInStock: 100}, nil
}

func (s *stubProductRepo) DecrementProductCountStock(ctx context.Context, productID int, quantity int) (*models.Product, error) {
	return &models.Product{ProductID: int32(productID), Price: 10000, CountInStock: 100}, nil
}

func (s *stubProductRepo) IncrementProductCountStock(ctx context.Context, productID int, quantity int) (*models.Product, error) {
	return &models.Product{ProductID: int32(productID), Price: 10000, CountInStock: 100}, nil
}

// stubOrderItemRepo implements orderitemadapter.QueryRepository and
// orderitemadapter.CommandRepository. Never reached in the failure-injection
// suite (merchant fails first).
type stubOrderItemRepo struct{}

func (s *stubOrderItemRepo) FindOrderItemByOrder(ctx context.Context, orderID int) ([]*models.OrderItem, error) {
	return nil, nil
}

func (s *stubOrderItemRepo) CalculateTotalPrice(ctx context.Context, orderID int) (*int32, error) {
	var total int32
	return &total, nil
}

func (s *stubOrderItemRepo) CreateOrderItem(ctx context.Context, req *requests.CreateOrderItemRecordRequest) (*models.OrderItem, error) {
	return &models.OrderItem{
		OrderID:   int32(req.OrderID),
		ProductID: int32(req.ProductID),
		Quantity:  int32(req.Quantity),
		Price:     int64(req.Price),
	}, nil
}

func (s *stubOrderItemRepo) UpdateOrderItem(ctx context.Context, req *requests.UpdateOrderItemRecordRequest) (*models.OrderItem, error) {
	return &models.OrderItem{
		OrderItemID: int32(req.OrderItemID),
		OrderID:     int32(req.OrderID),
		ProductID:   int32(req.ProductID),
		Quantity:    int32(req.Quantity),
		Price:       int64(req.Price),
	}, nil
}

func (s *stubOrderItemRepo) DeleteOrderItem(ctx context.Context, orderItemID int) error {
	return nil
}
