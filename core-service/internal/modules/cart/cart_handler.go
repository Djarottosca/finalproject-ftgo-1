package cart

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

func (h *Handler) Add(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)

	var req AddToCartRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	if err := h.service.Add(userID, req); err != nil {
		switch {
		case errors.Is(err, ErrInvalidQty):
			return response.ErrorResponse(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrProductNotFound):
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		default:
			return response.ErrorResponse(c, http.StatusInternalServerError, "failed to add to cart")
		}
	}

	return response.SuccessResponse(c, http.StatusCreated, "added to cart", nil)
}

func (h *Handler) List(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)

	res, err := h.service.List(userID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to list cart items")
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}

func (h *Handler) Update(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)

	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid product id")
	}

	var req UpdateCartRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	if err := h.service.Update(userID, productID, req); err != nil {
		switch {
		case errors.Is(err, ErrInvalidQty):
			return response.ErrorResponse(c, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrNotFound):
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		default:
			return response.ErrorResponse(c, http.StatusInternalServerError, "failed to update cart item")
		}
	}

	return response.SuccessResponse(c, http.StatusOK, "cart item updated", nil)
}

func (h *Handler) Delete(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)

	productID, err := strconv.Atoi(c.Param("productId"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid product id")
	}

	if err := h.service.Delete(userID, productID); err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to delete cart item")
	}

	return response.SuccessResponse(c, http.StatusOK, "cart item deleted", nil)
}
