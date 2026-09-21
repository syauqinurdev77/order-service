package repositories

import (
	"order-service/src/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OrderRepository interface {
	Begin() *gorm.DB
	Commit(tx *gorm.DB) error
	Rollback(tx *gorm.DB)
	FindCustomer(tx *gorm.DB, id uint) (*models.Customer, error)
	LockProduct(tx *gorm.DB, id uint) (*models.Product, error)
	DecreaseStock(tx *gorm.DB, productID uint, qty int) error
	RestoreStock(tx *gorm.DB, productID uint, qty int) error
	CreateOrder(tx *gorm.DB, order *models.Order) error
	CreateOrderItems(tx *gorm.DB, items []models.OrderItem) error
	LockOrder(tx *gorm.DB, id uint) (*models.Order, error)
	UpdateStatus(tx *gorm.DB, orderID uint, status string) error
	FindByID(id uint) (*models.Order, error)
	FindAll(page, limit int, customerID *uint) ([]models.Order, int64, error)
}

type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Begin() *gorm.DB {
	return r.db.Begin()
}

func (r *orderRepository) Commit(tx *gorm.DB) error {
	return tx.Commit().Error
}

func (r *orderRepository) Rollback(tx *gorm.DB) {
	tx.Rollback()
}

func (r *orderRepository) FindCustomer(tx *gorm.DB, id uint) (*models.Customer, error) {
	var customer models.Customer
	err := tx.First(&customer, id).Error
	if err != nil {
		return nil, err
	}
	return &customer, nil
}

func (r *orderRepository) LockProduct(tx *gorm.DB, id uint) (*models.Product, error) {
	var product models.Product
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, id).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *orderRepository) DecreaseStock(tx *gorm.DB, productID uint, qty int) error {
	return tx.Model(&models.Product{}).
		Where("id = ?", productID).
		Update("stock", gorm.Expr("stock - ?", qty)).Error
}

func (r *orderRepository) RestoreStock(tx *gorm.DB, productID uint, qty int) error {
	return tx.Model(&models.Product{}).
		Where("id = ?", productID).
		Update("stock", gorm.Expr("stock + ?", qty)).Error
}

func (r *orderRepository) CreateOrder(tx *gorm.DB, order *models.Order) error {
	return tx.Create(order).Error
}

func (r *orderRepository) CreateOrderItems(tx *gorm.DB, items []models.OrderItem) error {
	return tx.Create(&items).Error
}

func (r *orderRepository) LockOrder(tx *gorm.DB, id uint) (*models.Order, error) {
	var order models.Order
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Preload("Items").
		First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) UpdateStatus(tx *gorm.DB, orderID uint, status string) error {
	return tx.Model(&models.Order{}).
		Where("id = ?", orderID).
		Update("status", status).Error
}

func (r *orderRepository) FindByID(id uint) (*models.Order, error) {
	var order models.Order
	err := r.db.Preload("Items.Product").Preload("Customer").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindAll(page, limit int, customerID *uint) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.Model(&models.Order{})
	if customerID != nil {
		query = query.Where("customer_id = ?", *customerID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err := query.Preload("Items.Product").Preload("Customer").
		Order("id desc").
		Offset(offset).Limit(limit).
		Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}
