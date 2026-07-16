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

// NewHandler returns Handler. supplierRepo resolves the caller's supplier
// ID from user ID, since JWT only carries user_id/role.
func NewHandler(service Service, supplierRepo supplier.Repository) *Handler {
	return &Handler{service: service, supplierRepo: supplierRepo}
}

func (h *Handler) List(c echo.Context) error {
	var req ListRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid query parameter")
	}
	if err := c.Validate(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}
	req.Normalize()

	res, err := h.service.List(c.Request().Context(), req)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to fetch products")
	}

	return response.SuccessResponse(c, http.StatusOK, "products fetched", res)
}

func (h *Handler) Detail(c echo.Context) error {
	slug := c.Param("slug")
	if slug == "" {
		return response.ErrorResponse(c, http.StatusBadRequest, "slug is required")
	}

	res, err := h.service.Detail(c.Request().Context(), slug)
	if err != nil {
		if errors.Is(err, ErrProductNotFound) {
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to fetch product detail")
	}

	return response.SuccessResponse(c, http.StatusOK, "product detail fetched", res)
}

func (h *Handler) currentSupplierID(c echo.Context) (int, error) {
	userID, _ := c.Get("user_id").(int)

	sup, err := h.supplierRepo.FindByUserID(c.Request().Context(), userID)
	if err != nil {
		return 0, errors.New("you must be a registered supplier to perform this action")
	}
	return sup.ID, nil
}

func (h *Handler) Create(c echo.Context) error {
	supplierID, err := h.currentSupplierID(c)
	if err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, err.Error())
	}

	var req CreateProductRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	res, err := h.service.Create(c.Request().Context(), supplierID, req)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to create product")
	}

	return response.SuccessResponse(c, http.StatusCreated, "product created", res)
}

func (h *Handler) ListMine(c echo.Context) error {
	supplierID, err := h.currentSupplierID(c)
	if err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, err.Error())
	}

	res, err := h.service.ListMine(c.Request().Context(), supplierID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to list products")
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}

func (h *Handler) Update(c echo.Context) error {
	supplierID, err := h.currentSupplierID(c)
	if err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, err.Error())
	}

	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid product id")
	}

	var req UpdateProductRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	res, err := h.service.Update(c.Request().Context(), supplierID, productID, req)
	if err != nil {
		return h.mutationError(c, err, "failed to update product")
	}

	return response.SuccessResponse(c, http.StatusOK, "product updated", res)
}

func (h *Handler) SetDiscount(c echo.Context) error {
	supplierID, err := h.currentSupplierID(c)
	if err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, err.Error())
	}

	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid product id")
	}

	var req DiscountRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	res, err := h.service.SetDiscount(c.Request().Context(), supplierID, productID, req)
	if err != nil {
		return h.mutationError(c, err, "failed to set discount")
	}

	return response.SuccessResponse(c, http.StatusOK, "discount updated", res)
}

func (h *Handler) AdjustStock(c echo.Context) error {
	supplierID, err := h.currentSupplierID(c)
	if err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, err.Error())
	}

	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid product id")
	}

	var req StockAdjustRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	res, err := h.service.AdjustStock(c.Request().Context(), supplierID, productID, req)
	if err != nil {
		return h.mutationError(c, err, "failed to adjust stock")
	}

	return response.SuccessResponse(c, http.StatusOK, "stock updated", res)
}

func (h *Handler) Delete(c echo.Context) error {
	supplierID, err := h.currentSupplierID(c)
	if err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, err.Error())
	}

	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid product id")
	}

	if err := h.service.Delete(c.Request().Context(), supplierID, productID); err != nil {
		return h.mutationError(c, err, "failed to delete product")
	}

	return response.SuccessResponse(c, http.StatusOK, "product deleted", nil)
}

// AdminList handles GET /admin/products. Returns every product across all
// suppliers, unscoped — admin oversight, not the supplier's own listing.
func (h *Handler) AdminList(c echo.Context) error {
	res, err := h.service.AdminList(c.Request().Context())
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to list products")
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}

func (h *Handler) mutationError(c echo.Context, err error, fallback string) error {
	switch {
	case errors.Is(err, ErrProductNotFound):
		return response.ErrorResponse(c, http.StatusNotFound, err.Error())
	case errors.Is(err, ErrForbidden):
		return response.ErrorResponse(c, http.StatusForbidden, err.Error())
	case errors.Is(err, ErrInvalidDiscount), errors.Is(err, ErrInvalidStock):
		return response.ErrorResponse(c, http.StatusBadRequest, err.Error())
	default:
		return response.ErrorResponse(c, http.StatusInternalServerError, fallback)
	}
}
