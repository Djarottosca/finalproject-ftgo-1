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

	res, err := h.service.StockReport(threshold)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to generate stock report")
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}
