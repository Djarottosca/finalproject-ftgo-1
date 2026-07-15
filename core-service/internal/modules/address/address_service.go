package address

import (
	"errors"

	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
)

var (
	ErrNotFound  = errors.New("address not found")
	ErrForbidden = errors.New("address does not belong to this user")
)

// Service defines the address use cases exposed to the handler layer.
type Service interface {
	Create(userID int, req CreateAddressRequest) (*AddressResponse, error)
	List(userID int) ([]AddressResponse, error)
	Update(userID, addressID int, req UpdateAddressRequest) (*AddressResponse, error)
	Delete(userID, addressID int) error
}

type service struct {
	repo Repository
}

// NewService returns the Service implementation backed by the given Repository.
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) Create(userID int, req CreateAddressRequest) (*AddressResponse, error) {
	if req.IsPrimary {
		if err := s.repo.ClearPrimary(userID); err != nil {
			return nil, err
		}
	}

	addr := &models.Address{
		UserID:      userID,
		Label:       req.Label,
		FullAddress: req.FullAddress,
		City:        req.City,
		District:    req.District,
		PostalCode:  req.PostalCode,
		IsPrimary:   req.IsPrimary,
	}
	if err := s.repo.Create(addr); err != nil {
		return nil, err
	}

	return toResponse(addr), nil
}

func (s *service) List(userID int) ([]AddressResponse, error) {
	addrs, err := s.repo.ListByUserID(userID)
	if err != nil {
		return nil, err
	}

	res := make([]AddressResponse, 0, len(addrs))
	for _, addr := range addrs {
		res = append(res, *toResponse(&addr))
	}
	return res, nil
}

func (s *service) Update(userID, addressID int, req UpdateAddressRequest) (*AddressResponse, error) {
	addr, err := s.repo.FindByID(addressID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if addr.UserID != userID {
		return nil, ErrForbidden
	}

	if req.IsPrimary && !addr.IsPrimary {
		if err := s.repo.ClearPrimary(userID); err != nil {
			return nil, err
		}
	}

	addr.Label = req.Label
	addr.FullAddress = req.FullAddress
	addr.City = req.City
	addr.District = req.District
	addr.PostalCode = req.PostalCode
	addr.IsPrimary = req.IsPrimary

	if err := s.repo.Update(addr); err != nil {
		return nil, err
	}

	return toResponse(addr), nil
}

func (s *service) Delete(userID, addressID int) error {
	addr, err := s.repo.FindByID(addressID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if addr.UserID != userID {
		return ErrForbidden
	}
	return s.repo.Delete(addressID)
}
