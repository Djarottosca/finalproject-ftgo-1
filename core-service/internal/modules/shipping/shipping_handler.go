package shipping

import (
	"net/http"

	echo "github.com/labstack/echo/v4"

	"github.com/Djarottosca/finalproject-ftgo-1/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// SearchDestinations godoc
// GET /shipping/destinations?search=...
func (h *Handler) SearchDestinations(c echo.Context) error {
	search := c.QueryParam("search")

	res, err := h.service.SearchDestinations(c.Request().Context(), search)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to fetch destinations")
	}
	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}

// CalculateCost godoc
// POST /shipping/cost
func (h *Handler) CalculateCost(c echo.Context) error {
	var req CalculateCostRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	res, err := h.service.CalculateCost(c.Request().Context(), req)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to calculate shipping cost")
	}
	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}
