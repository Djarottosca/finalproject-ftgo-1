package payment

// CreatePaymentRequest is the body of POST /payments. The user has already
// checked out (an order exists); this starts the payment for that order.
type CreatePaymentRequest struct {
	OrderID int `json:"order_id"`
}

// PaymentResponse is what the client gets back. PaymentLink is where the user
// is redirected to pay (real Xendit link, or a placeholder in simulation).
type PaymentResponse struct {
	ID               int     `json:"id"`
	OrderID          int     `json:"order_id"`
	Amount           float64 `json:"amount"`
	Status           string  `json:"status"`
	PaymentLink      string  `json:"payment_link,omitempty"`
	PaymentReference string  `json:"payment_reference,omitempty"`
}
