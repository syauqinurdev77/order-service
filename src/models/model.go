package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	Id        uint           `gorm:"primaryKey" json:"id"`
	Sku       string         `gorm:"uniqueIndex;not null" json:"sku"`
	Name      string         `gorm:"not null" json:"name"`
	Price     float64        `gorm:"not null" json:"price"`
	Stock     int            `gorm:"not null" json:"stock"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type Customer struct {
	Id        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"not null" json:"name"`
	Email     string         `gorm:"uniqueIndex;not null" json:"email"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

type Order struct {
	Id          uint           `gorm:"primaryKey" json:"id"`
	CustomerId  uint           `gorm:"not null" json:"customer_id"`
	Customer    Customer       `gorm:"foreignKey:CustomerID" json:"-"`
	Status      string         `gorm:"not null;default:'pending'" json:"status"`
	TotalAmount float64        `gorm:"not null" json:"total_amount"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	Items []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
}

type OrderItem struct {
	Id           uint           `gorm:"primaryKey" json:"id"`
	OrderId      uint           `gorm:"not null" json:"order_id"`
	ProductId    uint           `gorm:"not null" json:"product_id"`
	Product      Product        `gorm:"foreignKey:ProductID" json:"product"`
	Qty          int            `gorm:"not null" json:"qty"`
	PriceAtOrder float64        `gorm:"not null" json:"price_at_order"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}
