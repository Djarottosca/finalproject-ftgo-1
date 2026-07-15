package product

import (
	"errors"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/slug"
)

var (
	ErrNotFound        = errors.New("product not found")
	ErrForbidden       = errors.New("product does not belong to this supplier")
	ErrInvalidDiscount = errors.New("discount_type must be percentage, fixed, or empty")
	ErrInvalidStock    = errors.New("resulting stock cannot be negative")
)

// Service defines the product use cases exposed to the handler layer.
type Service interface {
	Create(supplierID int, req CreateProductRequest) (*ProductResponse, error)
	Get(id int) (*ProductResponse, error)
	List(filter ListFilter) ([]ProductResponse, error)
	Update(supplierID, productID int, req UpdateProductRequest) (*ProductResponse, error)
	SetDiscount(supplierID, productID int, req DiscountRequest) (*ProductResponse, error)
	AdjustStock(supplierID, productID int, req StockAdjustRequest) (*ProductResponse, error)
	Delete(supplierID, productID int) error
}

type service struct {
	repo Repository
}

// NewService returns the Service implementation backed by the given Repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(supplierID int, req CreateProductRequest) (*ProductResponse, error) {
	var description *string
	if req.Description != "" {
		description = &req.Description
	}

	product := &models.Product{
		ProductName: req.ProductName,
		ProductSlug: slug.Generate(req.ProductName),
		CategoryID:  req.CategoryID,
		Unit:        req.Unit,
		Stock:       req.Stock,
		SupplierID:  supplierID,
		Price:       req.Price,
		Description: description,
		Status:      models.ProductStatusActive,
	}
	if err := s.repo.Create(product); err != nil {
		return nil, err
	}

	return toResponse(product), nil
}

func (s *service) Get(id int) (*ProductResponse, error) {
	product, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	return toResponse(product), nil
}

func (s *service) List(filter ListFilter) ([]ProductResponse, error) {
	products, err := s.repo.List(filter)
	if err != nil {
		return nil, err
	}

	res := make([]ProductResponse, 0, len(products))
	for _, product := range products {
		res = append(res, *toResponse(&product))
	}
	return res, nil
}

func (s *service) Update(supplierID, productID int, req UpdateProductRequest) (*ProductResponse, error) {
	product, err := s.repo.FindByID(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if product.SupplierID != supplierID {
		return nil, ErrForbidden
	}

	product.ProductName = req.ProductName
	product.CategoryID = req.CategoryID
	product.Unit = req.Unit
	product.Price = req.Price

	var description *string
	if req.Description != "" {
		description = &req.Description
	}
	product.Description = description

	if err := s.repo.Update(product); err != nil {
		return nil, err
	}

	return toResponse(product), nil
}

func (s *service) SetDiscount(supplierID, productID int, req DiscountRequest) (*ProductResponse, error) {
	product, err := s.repo.FindByID(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if product.SupplierID != supplierID {
		return nil, ErrForbidden
	}

	switch req.DiscountType {
	case "":
		product.DiscountType = nil
		product.DiscountAmount = nil
	case models.ProductDiscountPercentage, models.ProductDiscountFixed:
		if req.DiscountType == models.ProductDiscountPercentage && req.DiscountAmount > 100 {
			return nil, ErrInvalidDiscount
		}
		discountType := req.DiscountType
		amount := req.DiscountAmount
		product.DiscountType = &discountType
		product.DiscountAmount = &amount
	default:
		return nil, ErrInvalidDiscount
	}

	if err := s.repo.Update(product); err != nil {
		return nil, err
	}

	return toResponse(product), nil
}

func (s *service) AdjustStock(supplierID, productID int, req StockAdjustRequest) (*ProductResponse, error) {
	product, err := s.repo.FindByID(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if product.SupplierID != supplierID {
		return nil, ErrForbidden
	}

	newStock := product.Stock + req.Delta
	if newStock < 0 {
		return nil, ErrInvalidStock
	}
	product.Stock = newStock

	if err := s.repo.Update(product); err != nil {
		return nil, err
	}

	return toResponse(product), nil
}

func (s *service) Delete(supplierID, productID int) error {
	product, err := s.repo.FindByID(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if product.SupplierID != supplierID {
		return ErrForbidden
	}
	return s.repo.Delete(productID)
}
