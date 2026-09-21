package services

import (
	"errors"
	"fmt"
	"math"
	"order-service/src/models"
	"order-service/src/repositories"
	"strings"

	"gorm.io/gorm"
)

const (
	StatusPending   = "pending"
	StatusPaid      = "paid"
	StatusShipped   = "shipped"
	StatusCompleted = "completed"
	StatusCancelled = "cancelled"
)

var validTransitions = map[string][]string{
	StatusPending:   {StatusPaid, StatusCancelled},
	StatusPaid:      {StatusShipped, StatusCancelled},
	StatusShipped:   {StatusCompleted, StatusCancelled},
	StatusCompleted: {},
	StatusCancelled: {},
}

var ErrOrderNotFound = errors.New("order not found")

type InsufficientStockError struct {
	Sku       string
	Remaining int
	Requested int
}

func (e *InsufficientStockError) Error() string {
	return fmt.Sprintf("product '%s' remaining %d, requested %d", e.Sku, e.Remaining, e.Requested)
}

type InvalidTransitionError struct {
	Current string
	Target  string
}

func (e *InvalidTransitionError) Error() string {
	return fmt.Sprintf("cannot transition from '%s' to '%s'", e.Current, e.Target)
}

type OrderItemInput struct {
	ProductID uint
	Qty       int
}

type CreateOrderInput struct {
	CustomerID uint
	Items      []OrderItemInput
}

type OrderService interface {
	CreateOrder(input CreateOrderInput) (*models.Order, error)
	ListOrders(page, limit int, customerID *uint) ([]models.Order, int64, error)
	GetOrder(id uint) (*models.Order, error)
	UpdateStatus(id uint, status string) (*models.Order, error)
}

type orderService struct {
	repo repositories.OrderRepository
}

func NewOrderService(repo repositories.OrderRepository) OrderService {
	return &orderService{repo: repo}
}

func (s *orderService) CreateOrder(input CreateOrderInput) (*models.Order, error) {
	if input.CustomerID == 0 {
		return nil, newValidationError("customer_id is required")
	}
	if len(input.Items) == 0 {
		return nil, newValidationError("at least one item is required")
	}
	for _, item := range input.Items {
		if item.ProductID == 0 {
			return nil, newValidationError("product_id is required")
		}
		if item.Qty <= 0 {
			return nil, newValidationError("qty must be greater than 0")
		}
	}

	tx := s.repo.Begin()
	defer s.repo.Rollback(tx)

	if _, err := s.repo.FindCustomer(tx, input.CustomerID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}

	var total float64
	items := make([]models.OrderItem, 0, len(input.Items))

	for _, item := range input.Items {
		product, err := s.repo.LockProduct(tx, item.ProductID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrProductNotFound
			}
			return nil, err
		}

		if product.Stock < item.Qty {
			return nil, &InsufficientStockError{
				Sku:       product.Sku,
				Remaining: product.Stock,
				Requested: item.Qty,
			}
		}

		if err := s.repo.DecreaseStock(tx, product.Id, item.Qty); err != nil {
			return nil, err
		}

		total += float64(item.Qty) * product.Price
		items = append(items, models.OrderItem{
			ProductId:    product.Id,
			Qty:          item.Qty,
			PriceAtOrder: product.Price,
		})
	}

	order := &models.Order{
		CustomerId:  input.CustomerID,
		Status:      StatusPending,
		TotalAmount: round2(total),
	}

	if err := s.repo.CreateOrder(tx, order); err != nil {
		return nil, err
	}

	for i := range items {
		items[i].OrderId = order.Id
	}

	if err := s.repo.CreateOrderItems(tx, items); err != nil {
		return nil, err
	}

	if err := s.repo.Commit(tx); err != nil {
		return nil, err
	}

	return s.repo.FindByID(order.Id)
}

func (s *orderService) ListOrders(page, limit int, customerID *uint) ([]models.Order, int64, error) {
	return s.repo.FindAll(page, limit, customerID)
}

func (s *orderService) GetOrder(id uint) (*models.Order, error) {
	order, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}
	return order, nil
}

func (s *orderService) UpdateStatus(id uint, status string) (*models.Order, error) {
	target := strings.ToLower(strings.TrimSpace(status))
	if target == "" {
		return nil, newValidationError("status is required")
	}
	if !isValidStatus(target) {
		return nil, newValidationError("invalid status, must be one of: pending, paid, shipped, completed, cancelled")
	}

	tx := s.repo.Begin()
	defer s.repo.Rollback(tx)

	order, err := s.repo.LockOrder(tx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrOrderNotFound
		}
		return nil, err
	}

	if !isValidTransition(order.Status, target) {
		return nil, &InvalidTransitionError{
			Current: order.Status,
			Target:  target,
		}
	}

	if target == StatusCancelled {
		for _, item := range order.Items {
			if err := s.repo.RestoreStock(tx, item.ProductId, item.Qty); err != nil {
				return nil, err
			}
		}
	}

	if err := s.repo.UpdateStatus(tx, id, target); err != nil {
		return nil, err
	}

	if err := s.repo.Commit(tx); err != nil {
		return nil, err
	}

	return s.repo.FindByID(id)
}

func isValidStatus(status string) bool {
	_, ok := validTransitions[status]
	return ok
}

func isValidTransition(from, to string) bool {
	for _, next := range validTransitions[from] {
		if next == to {
			return true
		}
	}
	return false
}

func round2(value float64) float64 {
	return math.Round(value*100) / 100
}
