package supplier

import "github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"

func toResponse(supplier *models.Supplier) *SupplierResponse {
	return &SupplierResponse{
		ID:        supplier.ID,
		UserID:    supplier.UserID,
		StoreName: supplier.StoreName,
		Slug:      supplier.SupplierSlug,
		Address:   supplier.Address,
		Status:    supplier.Status,
	}
}
