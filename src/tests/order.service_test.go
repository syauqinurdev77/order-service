package services_test

import (
	"errors"
	"testing"

	"order-service/src/models"
	"order-service/src/services"
)

func TestCreateOrderSuccess(t *testing.T) {
	repo := newFakeOrderRepo()
	repo.seed()
	svc := services.NewOrderService(repo)

	order, err := svc.CreateOrder(services.CreateOrderInput{
		CustomerID: 1,
		Items: []services.OrderItemInput{
			{ProductID: 1, Qty: 2},
			{ProductID: 2, Qty: 1},
		},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if order.Status != services.StatusPending {
		t.Errorf("expected status pending, got %s", order.Status)
	}
	if order.TotalAmount != 40000 {
		t.Errorf("expected total 40000 (2*10000 + 1*20000), got %.2f", order.TotalAmount)
	}
	if len(order.Items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(order.Items))
	}
	if order.Items[0].PriceAtOrder != 10000 {
		t.Errorf("expected price_at_order 10000, got %.2f", order.Items[0].PriceAtOrder)
	}
	if repo.products[1].Stock != 3 {
		t.Errorf("expected product 1 stock 3, got %d", repo.products[1].Stock)
	}
	if repo.products[2].Stock != 9 {
		t.Errorf("expected product 2 stock 9, got %d", repo.products[2].Stock)
	}
	if repo.commits != 1 {
		t.Errorf("expected 1 commit, got %d", repo.commits)
	}
}

func TestInsufficientStock(t *testing.T) {
	repo := newFakeOrderRepo()
	repo.seed()
	repo.products[2].Stock = 1
	svc := services.NewOrderService(repo)

	_, err := svc.CreateOrder(services.CreateOrderInput{
		CustomerID: 1,
		Items: []services.OrderItemInput{
			{ProductID: 1, Qty: 2},
			{ProductID: 2, Qty: 5},
		},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var stockErr *services.InsufficientStockError
	if !errors.As(err, &stockErr) {
		t.Fatalf("expected InsufficientStockError, got %T", err)
	}
	if stockErr.Sku != "CEL-01" || stockErr.Remaining != 1 || stockErr.Requested != 5 {
		t.Errorf("unexpected stock error detail: %+v", stockErr)
	}
}

func TestAtomicity(t *testing.T) {
	repo := newFakeOrderRepo()
	repo.seed()
	repo.products[2].Stock = 1
	svc := services.NewOrderService(repo)

	_, err := svc.CreateOrder(services.CreateOrderInput{
		CustomerID: 1,
		Items: []services.OrderItemInput{
			{ProductID: 1, Qty: 2},
			{ProductID: 2, Qty: 5},
		},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if repo.products[1].Stock != 5 {
		t.Errorf("expected product 1 stock unchanged at 5, got %d", repo.products[1].Stock)
	}
	if repo.products[2].Stock != 1 {
		t.Errorf("expected product 2 stock unchanged at 1, got %d", repo.products[2].Stock)
	}
	if len(repo.orders) != 0 {
		t.Errorf("expected no order stored, got %d", len(repo.orders))
	}
	if repo.rollbacks == 0 {
		t.Error("expected transaction to be rolled back")
	}
}

func TestInvalidTransition(t *testing.T) {
	repo := newFakeOrderRepo()
	repo.seed()
	repo.orders[1] = &models.Order{
		Id:         1,
		CustomerId: 1,
		Status:     services.StatusPending,
		Items: []models.OrderItem{
			{Id: 1, OrderId: 1, ProductId: 1, Qty: 2, PriceAtOrder: 10000},
		},
	}
	svc := services.NewOrderService(repo)

	_, err := svc.UpdateStatus(1, "shipped")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	var transitionErr *services.InvalidTransitionError
	if !errors.As(err, &transitionErr) {
		t.Fatalf("expected InvalidTransitionError, got %T", err)
	}
	if transitionErr.Current != services.StatusPending || transitionErr.Target != services.StatusShipped {
		t.Errorf("unexpected transition detail: %+v", transitionErr)
	}
	if repo.orders[1].Status != services.StatusPending {
		t.Errorf("expected order status unchanged at pending, got %s", repo.orders[1].Status)
	}
}

func TestCancelRestoresStock(t *testing.T) {
	repo := newFakeOrderRepo()
	repo.seed()
	repo.products[1].Stock = 5
	repo.orders[1] = &models.Order{
		Id:         1,
		CustomerId: 1,
		Status:     services.StatusPaid,
		Items: []models.OrderItem{
			{Id: 1, OrderId: 1, ProductId: 1, Qty: 2, PriceAtOrder: 10000},
		},
	}
	svc := services.NewOrderService(repo)

	order, err := svc.UpdateStatus(1, "cancelled")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if order.Status != services.StatusCancelled {
		t.Errorf("expected status cancelled, got %s", order.Status)
	}
	if repo.products[1].Stock != 7 {
		t.Errorf("expected stock restored to 7, got %d", repo.products[1].Stock)
	}
	if repo.commits != 1 {
		t.Errorf("expected 1 commit, got %d", repo.commits)
	}
}

func TestShippedCanBeCancelled(t *testing.T) {
	repo := newFakeOrderRepo()
	repo.seed()
	repo.products[1].Stock = 5
	repo.orders[1] = &models.Order{
		Id:         1,
		CustomerId: 1,
		Status:     services.StatusShipped,
		Items: []models.OrderItem{
			{Id: 1, OrderId: 1, ProductId: 1, Qty: 2, PriceAtOrder: 10000},
		},
	}
	svc := services.NewOrderService(repo)

	order, err := svc.UpdateStatus(1, "cancelled")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if order.Status != services.StatusCancelled {
		t.Errorf("expected status cancelled, got %s", order.Status)
	}
	if repo.products[1].Stock != 7 {
		t.Errorf("expected stock restored to 7, got %d", repo.products[1].Stock)
	}
}

func TestOrderNotFound(t *testing.T) {
	repo := newFakeOrderRepo()
	svc := services.NewOrderService(repo)

	_, err := svc.GetOrder(999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, services.ErrOrderNotFound) {
		t.Errorf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestCustomerNotFound(t *testing.T) {
	repo := newFakeOrderRepo()
	repo.seed()
	delete(repo.customers, 1)
	svc := services.NewOrderService(repo)

	_, err := svc.CreateOrder(services.CreateOrderInput{
		CustomerID: 1,
		Items: []services.OrderItemInput{
			{ProductID: 1, Qty: 1},
		},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, services.ErrCustomerNotFound) {
		t.Errorf("expected ErrCustomerNotFound, got %v", err)
	}
}
