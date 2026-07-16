package order

import (
	"errors"
	"net/http"
	"strconv"

	echo "github.com/labstack/echo/v4"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/supplier"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/response"
)

type Handler struct {
	service      Service
	supplierRepo supplier.Repository
}

// NewHandler returns Handler. supplierRepo resolves the caller's supplier
// ID from user ID, since JWT only carries user_id/role.
func NewHandler(service Service, supplierRepo supplier.Repository) *Handler {
	return &Handler{service: service, supplierRepo: supplierRepo}
}

func (h *Handler) Checkout(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)

	res, err := h.service.Checkout(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, ErrEmptyCart) {
			return response.ErrorResponse(c, http.StatusBadRequest, err.Error())
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to checkout")
	}

	return response.SuccessResponse(c, http.StatusCreated, "order created", res)
}

func (h *Handler) ListMine(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)

	res, err := h.service.ListMine(c.Request().Context(), userID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to list orders")
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}

func (h *Handler) Get(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid id")
	}

	res, err := h.service.Get(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to get order")
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}

func (h *Handler) ListForSupplier(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)
	sup, err := h.supplierRepo.FindByUserID(c.Request().Context(), userID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, "you must be a registered supplier to perform this action")
	}

	res, err := h.service.ListForSupplier(c.Request().Context(), sup.ID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to list orders")
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}

func (h *Handler) UpdateStatus(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)
	sup, err := h.supplierRepo.FindByUserID(c.Request().Context(), userID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, "you must be a registered supplier to perform this action")
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid id")
	}

	var req UpdateOrderStatusRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	res, err := h.service.UpdateStatus(c.Request().Context(), sup.ID, id, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrForbidden):
			return response.ErrorResponse(c, http.StatusForbidden, err.Error())
		case errors.Is(err, ErrInvalidStatus):
			return response.ErrorResponse(c, http.StatusBadRequest, err.Error())
		default:
			return response.ErrorResponse(c, http.StatusInternalServerError, "failed to update order status")
		}
	}

	return response.SuccessResponse(c, http.StatusOK, "order status updated", res)
}
