package product

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

// NewHandler returns a Handler. supplierRepo resolves the caller's supplier
// ID from their user ID, since the JWT only carries user_id/role.
func NewHandler(service Service, supplierRepo supplier.Repository) *Handler {
	return &Handler{service: service, supplierRepo: supplierRepo}
}

func (h *Handler) Create(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)
	sup, err := h.supplierRepo.FindByUserID(userID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, "you must be a registered supplier to perform this action")
	}

	var req CreateProductRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	res, err := h.service.Create(sup.ID, req)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to create product")
	}

	return response.SuccessResponse(c, http.StatusCreated, "product created", res)
}

func (h *Handler) Get(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid id")
	}

	res, err := h.service.Get(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to get product")
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}

func (h *Handler) List(c echo.Context) error {
	filter := ListFilter{Status: c.QueryParam("status")}
	if v, err := strconv.Atoi(c.QueryParam("category_id")); err == nil {
		filter.CategoryID = v
	}
	if v, err := strconv.Atoi(c.QueryParam("supplier_id")); err == nil {
		filter.SupplierID = v
	}

	res, err := h.service.List(filter)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to list products")
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}

// ListMine lists only the caller's own products, for the supplier dashboard.
func (h *Handler) ListMine(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)
	sup, err := h.supplierRepo.FindByUserID(userID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, "you must be a registered supplier to perform this action")
	}

	res, err := h.service.List(ListFilter{SupplierID: sup.ID})
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to list products")
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}

// GetMine gets a single product, but only if it belongs to the caller.
func (h *Handler) GetMine(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)
	sup, err := h.supplierRepo.FindByUserID(userID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, "you must be a registered supplier to perform this action")
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid id")
	}

	res, err := h.service.Get(id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to get product")
	}
	if res.SupplierID != sup.ID {
		return response.ErrorResponse(c, http.StatusForbidden, "product does not belong to this supplier")
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}

func (h *Handler) Update(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)
	sup, err := h.supplierRepo.FindByUserID(userID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, "you must be a registered supplier to perform this action")
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid id")
	}

	var req UpdateProductRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	res, err := h.service.Update(sup.ID, id, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrForbidden):
			return response.ErrorResponse(c, http.StatusForbidden, err.Error())
		default:
			return response.ErrorResponse(c, http.StatusInternalServerError, "failed to update product")
		}
	}

	return response.SuccessResponse(c, http.StatusOK, "product updated", res)
}

func (h *Handler) SetDiscount(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)
	sup, err := h.supplierRepo.FindByUserID(userID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, "you must be a registered supplier to perform this action")
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid id")
	}

	var req DiscountRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	res, err := h.service.SetDiscount(sup.ID, id, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrForbidden):
			return response.ErrorResponse(c, http.StatusForbidden, err.Error())
		case errors.Is(err, ErrInvalidDiscount):
			return response.ErrorResponse(c, http.StatusBadRequest, err.Error())
		default:
			return response.ErrorResponse(c, http.StatusInternalServerError, "failed to set discount")
		}
	}

	return response.SuccessResponse(c, http.StatusOK, "discount updated", res)
}

func (h *Handler) AdjustStock(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)
	sup, err := h.supplierRepo.FindByUserID(userID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, "you must be a registered supplier to perform this action")
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid id")
	}

	var req StockAdjustRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	res, err := h.service.AdjustStock(sup.ID, id, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrForbidden):
			return response.ErrorResponse(c, http.StatusForbidden, err.Error())
		case errors.Is(err, ErrInvalidStock):
			return response.ErrorResponse(c, http.StatusBadRequest, err.Error())
		default:
			return response.ErrorResponse(c, http.StatusInternalServerError, "failed to adjust stock")
		}
	}

	return response.SuccessResponse(c, http.StatusOK, "stock updated", res)
}

func (h *Handler) Delete(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)
	sup, err := h.supplierRepo.FindByUserID(userID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, "you must be a registered supplier to perform this action")
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid id")
	}

	if err := h.service.Delete(sup.ID, id); err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrForbidden):
			return response.ErrorResponse(c, http.StatusForbidden, err.Error())
		default:
			return response.ErrorResponse(c, http.StatusInternalServerError, "failed to delete product")
		}
	}

	return response.SuccessResponse(c, http.StatusOK, "product deleted", nil)
}
