package services_test

import (
	"sync"

	"order-service/src/models"

	"gorm.io/gorm"
)

type txState struct {
	stockDelta    map[uint]int
	newOrder      *models.Order
	newItems      []models.OrderItem
	statusUpdates map[uint]string
}

type fakeOrderRepo struct {
	mu        sync.Mutex
	customers map[uint]*models.Customer
	products  map[uint]*models.Product
	orders    map[uint]*models.Order
	nextID    uint
	pending   *txState
	commits   int
	rollbacks int
}

func newFakeOrderRepo() *fakeOrderRepo {
	return &fakeOrderRepo{
		customers: map[uint]*models.Customer{},
		products:  map[uint]*models.Product{},
		orders:    map[uint]*models.Order{},
		nextID:    1,
	}
}

func (f *fakeOrderRepo) seed() {
	f.customers[1] = &models.Customer{Id: 1, Name: "Budi", Email: "budi@mail.com"}
	f.products[1] = &models.Product{Id: 1, Sku: "KAOS-01", Name: "Kaos", Price: 10000, Stock: 5}
	f.products[2] = &models.Product{Id: 2, Sku: "CEL-01", Name: "Celana", Price: 20000, Stock: 10}
}

func (f *fakeOrderRepo) Begin() *gorm.DB {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pending = &txState{
		stockDelta:    map[uint]int{},
		statusUpdates: map[uint]string{},
	}
	return nil
}

func (f *fakeOrderRepo) Commit(tx *gorm.DB) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.pending == nil {
		return nil
	}
	for id, delta := range f.pending.stockDelta {
		f.products[id].Stock += delta
	}
	if f.pending.newOrder != nil {
		order := f.pending.newOrder
		order.Items = f.pending.newItems
		f.orders[order.Id] = order
	}
	for id, status := range f.pending.statusUpdates {
		if order, ok := f.orders[id]; ok {
			order.Status = status
		}
	}
	f.pending = nil
	f.commits++
	return nil
}

func (f *fakeOrderRepo) Rollback(tx *gorm.DB) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pending = nil
	f.rollbacks++
}

func (f *fakeOrderRepo) FindCustomer(tx *gorm.DB, id uint) (*models.Customer, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	customer, ok := f.customers[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return customer, nil
}

func (f *fakeOrderRepo) LockProduct(tx *gorm.DB, id uint) (*models.Product, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	product, ok := f.products[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return product, nil
}

func (f *fakeOrderRepo) DecreaseStock(tx *gorm.DB, productID uint, qty int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pending.stockDelta[productID] -= qty
	return nil
}

func (f *fakeOrderRepo) RestoreStock(tx *gorm.DB, productID uint, qty int) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pending.stockDelta[productID] += qty
	return nil
}

func (f *fakeOrderRepo) CreateOrder(tx *gorm.DB, order *models.Order) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	order.Id = f.nextID
	f.nextID++
	f.pending.newOrder = order
	return nil
}

func (f *fakeOrderRepo) CreateOrderItems(tx *gorm.DB, items []models.OrderItem) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := range items {
		items[i].Id = f.nextID
		f.nextID++
	}
	f.pending.newItems = append(f.pending.newItems, items...)
	return nil
}

func (f *fakeOrderRepo) LockOrder(tx *gorm.DB, id uint) (*models.Order, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	order, ok := f.orders[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return order, nil
}

func (f *fakeOrderRepo) UpdateStatus(tx *gorm.DB, orderID uint, status string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.pending.statusUpdates[orderID] = status
	return nil
}

func (f *fakeOrderRepo) FindByID(id uint) (*models.Order, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	order, ok := f.orders[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return order, nil
}

func (f *fakeOrderRepo) FindAll(page, limit int, customerID *uint) ([]models.Order, int64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var filtered []models.Order
	for _, order := range f.orders {
		if customerID != nil && order.CustomerId != *customerID {
			continue
		}
		filtered = append(filtered, *order)
	}
	total := int64(len(filtered))
	offset := (page - 1) * limit
	if offset > len(filtered) {
		offset = len(filtered)
	}
	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}
	return filtered[offset:end], total, nil
}
