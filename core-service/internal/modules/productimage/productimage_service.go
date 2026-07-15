package productimage

import (
	"errors"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/product"
)

var (
	ErrNotFound        = errors.New("image not found")
	ErrProductNotFound = errors.New("product not found")
	ErrForbidden       = errors.New("product does not belong to this supplier")
)

// Service defines the product image use cases exposed to the handler layer.
type Service interface {
	Add(supplierID, productID int, req AddImageRequest) (*ImageResponse, error)
	List(productID int) ([]ImageResponse, error)
	Delete(supplierID, imageID int) error
}

type service struct {
	repo        Repository
	productRepo product.Repository
}

// NewService returns the Service implementation backed by the given repositories.
func NewService(repo Repository, productRepo product.Repository) Service {
	return &service{repo: repo, productRepo: productRepo}
}

func (s *service) Add(supplierID, productID int, req AddImageRequest) (*ImageResponse, error) {
	prod, err := s.productRepo.FindByID(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}
	if prod.SupplierID != supplierID {
		return nil, ErrForbidden
	}

	image := &models.ProductImage{
		ProductID: productID,
		ImageURL:  req.ImageURL,
	}
	if err := s.repo.Create(image); err != nil {
		return nil, err
	}

	return toResponse(image), nil
}

func (s *service) List(productID int) ([]ImageResponse, error) {
	images, err := s.repo.ListByProductID(productID)
	if err != nil {
		return nil, err
	}

	res := make([]ImageResponse, 0, len(images))
	for _, image := range images {
		res = append(res, *toResponse(&image))
	}
	return res, nil
}

func (s *service) Delete(supplierID, imageID int) error {
	image, err := s.repo.FindByID(imageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}

	prod, err := s.productRepo.FindByID(image.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProductNotFound
		}
		return err
	}
	if prod.SupplierID != supplierID {
		return ErrForbidden
	}

	return s.repo.Delete(imageID)
}
