package payment

import "github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"

// toResponse converts the stored payment into the API shape. PaymentLink and
// PaymentReference are pointers on the model (nullable columns), so they're
// dereferenced only when present.
func toResponse(p *models.Payment) *PaymentResponse {
	res := &PaymentResponse{
		ID:      p.ID,
		OrderID: p.OrderID,
		Amount:  p.Amount,
		Status:  p.Status,
	}
	if p.PaymentLink != nil {
		res.PaymentLink = *p.PaymentLink
	}
	if p.PaymentReference != nil {
		res.PaymentReference = *p.PaymentReference
	}
	return res
}