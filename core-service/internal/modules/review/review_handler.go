package review

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

// Create handles POST /products/:id/reviews.
func (h *Handler) Create(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)

	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid product id")
	}

	var req CreateReviewRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	res, err := h.service.Create(c.Request().Context(), userID, productID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrProductNotFound):
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrAlreadyReviewed):
			return response.ErrorResponse(c, http.StatusConflict, err.Error())
		default:
			return response.ErrorResponse(c, http.StatusInternalServerError, "failed to create review")
		}
	}

	return response.SuccessResponse(c, http.StatusCreated, "review created", res)
}

// List handles GET /products/:id/reviews.
func (h *Handler) List(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid product id")
	}

	res, err := h.service.ListByProduct(c.Request().Context(), productID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to list reviews")
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}
