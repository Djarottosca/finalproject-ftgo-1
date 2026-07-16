package admin

import (
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

// StockReport handles GET /admin/reports/stock?threshold=N
func (h *Handler) StockReport(c echo.Context) error {
	threshold, _ := strconv.Atoi(c.QueryParam("threshold")) // 0 if absent/invalid; service applies default

	res, err := h.service.StockReport(c.Request().Context(), threshold)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to generate stock report")
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}

// SalesReport handles GET /admin/reports/sales
func (h *Handler) SalesReport(c echo.Context) error {
	res, err := h.service.SalesReport(c.Request().Context())
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to generate sales report")
	}
	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}

// Transactions handles GET /admin/transactions
func (h *Handler) Transactions(c echo.Context) error {
	res, err := h.service.Transactions(c.Request().Context())
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to list transactions")
	}
	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}
