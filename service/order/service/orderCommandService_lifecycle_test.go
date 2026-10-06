package service

import (
	"context"
	"errors"
	"testing"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
)

type lifecycleStockCall struct {
	productID int
	quantity  int
	increment bool
}

type lifecycleStockRepository struct {
	calls       []lifecycleStockCall
	failProduct int
	failOn      int
}

func (r *lifecycleStockRepository) DecrementProductCountStock(_ context.Context, productID, quantity int) (*models.Product, error) {
	r.calls = append(r.calls, lifecycleStockCall{productID: productID, quantity: quantity})
	if !r.shouldFail(productID) {
		return &models.Product{ProductID: int32(productID)}, nil
	}
	return nil, errors.New("decrement stock failed")
}

func (r *lifecycleStockRepository) IncrementProductCountStock(_ context.Context, productID, quantity int) (*models.Product, error) {
	r.calls = append(r.calls, lifecycleStockCall{productID: productID, quantity: quantity, increment: true})
	if !r.shouldFail(productID) {
		return &models.Product{ProductID: int32(productID)}, nil
	}
	return nil, errors.New("increment stock failed")
}

func (r *lifecycleStockRepository) shouldFail(productID int) bool {
	if r.failProduct != productID {
		return false
	}
	return r.failOn == 0 || len(r.calls) == r.failOn
}

func TestIncrementStockForOrderItemsCompensatesPartialMutation(t *testing.T) {
	repo := &lifecycleStockRepository{failProduct: 2}
	service := &orderCommandService{productCommandRepository: repo}
	items := []*models.OrderItem{
		{ProductID: 1, Quantity: 3},
		{ProductID: 2, Quantity: 4},
	}

	_, err := service.incrementStockForOrderItems(context.Background(), items)
	if err == nil {
		t.Fatal("expected increment failure")
	}

	want := []lifecycleStockCall{
		{productID: 1, quantity: 3, increment: true},
		{productID: 2, quantity: 4, increment: true},
		{productID: 1, quantity: 3, increment: false},
	}
	assertLifecycleStockCalls(t, want, repo.calls)
}

func TestDecrementStockForOrderItemsCompensatesPartialMutation(t *testing.T) {
	repo := &lifecycleStockRepository{failProduct: 2}
	service := &orderCommandService{productCommandRepository: repo}
	items := []*models.OrderItem{
		{ProductID: 1, Quantity: 2},
		{ProductID: 2, Quantity: 5},
	}

	_, err := service.decrementStockForOrderItems(context.Background(), items)
	if err == nil {
		t.Fatal("expected decrement failure")
	}

	want := []lifecycleStockCall{
		{productID: 1, quantity: 2, increment: false},
		{productID: 2, quantity: 5, increment: false},
		{productID: 1, quantity: 2, increment: true},
	}
	assertLifecycleStockCalls(t, want, repo.calls)
}

func assertLifecycleStockCalls(t *testing.T, want, got []lifecycleStockCall) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("unexpected call count: got %d, want %d; calls=%v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("unexpected call at index %d: got %+v, want %+v", i, got[i], want[i])
		}
	}
}
