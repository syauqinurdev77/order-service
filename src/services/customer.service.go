package services

import (
	"errors"
	"order-service/src/models"
	"order-service/src/repositories"

	"gorm.io/gorm"
)

var (
	ErrDuplicateEmail   = errors.New("duplicate email")
	ErrCustomerNotFound = errors.New("customer not found")
)

type CustomerService interface {
	CreateCustomer(input models.Customer) (*models.Customer, error)
	ListCustomers(page, limit int, q string) ([]models.Customer, int64, error)
	GetCustomer(id uint) (*models.Customer, error)
	UpdateCustomer(id uint, input models.Customer) (*models.Customer, error)
}

type customerService struct {
	repo repositories.CustomerRepository
}

func NewCustomerService(repo repositories.CustomerRepository) CustomerService {
	return &customerService{repo: repo}
}

func (s *customerService) CreateCustomer(input models.Customer) (*models.Customer, error) {
	if err := validateCustomerInput(input); err != nil {
		return nil, err
	}

	_, err := s.repo.FindByEmail(input.Email)
	switch {
	case err == nil:
		return nil, ErrDuplicateEmail
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, err
	}

	customer := &models.Customer{
		Name:  input.Name,
		Email: input.Email,
	}

	if err := s.repo.Create(customer); err != nil {
		return nil, err
	}

	return customer, nil
}

func (s *customerService) ListCustomers(page, limit int, q string) ([]models.Customer, int64, error) {
	return s.repo.FindAll(page, limit, q)
}

func (s *customerService) GetCustomer(id uint) (*models.Customer, error) {
	customer, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCustomerNotFound
		}
		return nil, err
	}
	return customer, nil
}

func (s *customerService) UpdateCustomer(id uint, input models.Customer) (*models.Customer, error) {
	if err := validateCustomerInput(input); err != nil {
		return nil, err
	}

	customer, err := s.GetCustomer(id)
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.FindByEmail(input.Email)
	switch {
	case err == nil && existing.Id != customer.Id:
		return nil, ErrDuplicateEmail
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound):
		return nil, err
	}

	customer.Name = input.Name
	customer.Email = input.Email

	if err := s.repo.Update(customer); err != nil {
		return nil, err
	}

	return customer, nil
}

func validateCustomerInput(input models.Customer) error {
	if input.Name == "" {
		return newValidationError("name is required")
	}
	if input.Email == "" {
		return newValidationError("email is required")
	}
	return nil
}
