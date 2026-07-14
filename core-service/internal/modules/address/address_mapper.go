package address

import "github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"

func toResponse(addr *models.Address) *AddressResponse {
	return &AddressResponse{
		ID:          addr.ID,
		UserID:      addr.UserID,
		Label:       addr.Label,
		FullAddress: addr.FullAddress,
		City:        addr.City,
		District:    addr.District,
		PostalCode:  addr.PostalCode,
		IsPrimary:   addr.IsPrimary,
	}
}
