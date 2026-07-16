package cart

import (
	"errors"
	"net/http"
	"strconv"

	echo "github.com/labstack/echo/v4"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/middleware"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetCart(c echo.Context) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return response.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
	}

	res, err := h.service.GetCart(c.Request().Context(), userID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to fetch cart")
	}
	return response.SuccessResponse(c, http.StatusOK, "cart fetched", res)
}

func (h *Handler) AddItem(c echo.Context) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return response.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
	}

	var req AddItemRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	res, err := h.service.AddItem(c.Request().Context(), userID, req)
	if err != nil {
		return mapCartError(c, err)
	}
	return response.SuccessResponse(c, http.StatusCreated, "item added to cart", res)
}

func (h *Handler) UpdateItem(c echo.Context) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return response.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
	}

	productID, err := strconv.ParseUint(c.Param("product_id"), 10, 64)
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid product id")
	}

	var req UpdateItemRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	res, err := h.service.UpdateItem(c.Request().Context(), userID, productID, req)
	if err != nil {
		return mapCartError(c, err)
	}
	return response.SuccessResponse(c, http.StatusOK, "cart item updated", res)
}

func (h *Handler) RemoveItem(c echo.Context) error {
	userID, ok := middleware.UserIDFromContext(c)
	if !ok {
		return response.ErrorResponse(c, http.StatusUnauthorized, "unauthorized")
	}

	productID, err := strconv.ParseUint(c.Param("product_id"), 10, 64)
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid product id")
	}

	res, err := h.service.RemoveItem(c.Request().Context(), userID, productID)
	if err != nil {
		return mapCartError(c, err)
	}
	return response.SuccessResponse(c, http.StatusOK, "item removed from cart", res)
}

func mapCartError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, ErrProductNotFound), errors.Is(err, ErrItemNotFound):
		return response.ErrorResponse(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrProductInactive), errors.Is(err, ErrInsufficientStock):
		return response.ErrorResponse(c, http.StatusBadRequest, err.Error())
	default:
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to update cart")
	}
}
