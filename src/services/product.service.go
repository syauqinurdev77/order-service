package services

import (
	"errors"
	"fmt"
	"order-service/src/models"
	"order-service/src/repositories"

	"gorm.io/gorm"
)

var (
	ErrDuplicateSKU    = errors.New("duplicate sku")
	ErrProductNotFound = errors.New("product not found")
)

type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string {
	return e.Message
}

func newValidationError(message string) error {
	return ValidationError{Message: message}
}

type ProductService interface {
	CreateProduct(input models.Product) (*models.Product, error)
	ListProducts(page, limit int, q string) ([]models.Product, int64, error)
	GetProduct(id uint) (*models.Product, error)
	UpdateProduct(id uint, input models.Product) (*models.Product, error)
}

type productService struct {
	repo repositories.ProductRepository
}

func NewProductService(repo repositories.ProductRepository) ProductService {
	return &productService{repo: repo}
}

func (s *productService) CreateProduct(input models.Product) (*models.Product, error) {
	if err := validateProductInput(input, true); err != nil {
		return nil, err
	}

	_, err := s.repo.FindBySKU(input.Sku)
	switch {
	case err == nil:
		return nil, ErrDuplicateSKU
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, err
	}

	product := &models.Product{
		Sku:   input.Sku,
		Name:  input.Name,
		Price: input.Price,
		Stock: input.Stock,
	}

	if err := s.repo.Create(product); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *productService) ListProducts(page, limit int, q string) ([]models.Product, int64, error) {
	return s.repo.FindAll(page, limit, q)
}

func (s *productService) GetProduct(id uint) (*models.Product, error) {
	product, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	return product, nil
}

func (s *productService) UpdateProduct(id uint, input models.Product) (*models.Product, error) {
	if err := validateProductInput(input, false); err != nil {
		return nil, err
	}

	product, err := s.GetProduct(id)
	if err != nil {
		return nil, err
	}

	product.Name = input.Name
	product.Price = input.Price
	product.Stock = input.Stock

	if err := s.repo.Update(product); err != nil {
		return nil, err
	}

	return product, nil
}

func validateProductInput(input models.Product, requireSKU bool) error {
	if requireSKU && input.Sku == "" {
		return newValidationError("sku is required")
	}
	if input.Name == "" {
		return newValidationError("name is required")
	}
	if input.Price <= 0 {
		return newValidationError(fmt.Sprintf("price must be greater than 0, got %.2f", input.Price))
	}
	if input.Stock < 0 {
		return newValidationError(fmt.Sprintf("stock cannot be negative, got %d", input.Stock))
	}
	return nil
}
