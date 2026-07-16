package payment

import (
	"errors"
	"net/http"
	"strconv"

	echo "github.com/labstack/echo/v4"

	"github.com/Djarottosca/finalproject-ftgo-1/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// Create handles POST /payments. Body: {"order_id": N}. Starts payment for an
// order the caller owns and returns the payment link to redirect them to.
func (h *Handler) Create(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)

	var req CreatePaymentRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if req.OrderID == 0 {
		return response.ErrorResponse(c, http.StatusBadRequest, "order_id is required")
	}

	res, err := h.service.CreateForOrder(c.Request().Context(), req.OrderID, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrOrderNotFound):
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrForbidden):
			return response.ErrorResponse(c, http.StatusForbidden, err.Error())
		case errors.Is(err, ErrOrderNotPayable):
			return response.ErrorResponse(c, http.StatusConflict, err.Error())
		default:
			return response.ErrorResponse(c, http.StatusInternalServerError, "failed to create payment")
		}
	}

	return response.SuccessResponse(c, http.StatusCreated, "payment created", res)
}

// GetStatus handles GET /payments/:orderId. Returns the stored payment,
// refreshing from payment-service if it's still pending.
func (h *Handler) GetStatus(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)

	orderID, err := strconv.Atoi(c.Param("orderId"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid order id")
	}

	res, err := h.service.GetStatus(c.Request().Context(), orderID, userID)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound), errors.Is(err, ErrOrderNotFound):
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrForbidden):
			return response.ErrorResponse(c, http.StatusForbidden, err.Error())
		default:
			return response.ErrorResponse(c, http.StatusInternalServerError, "failed to get payment status")
		}
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}