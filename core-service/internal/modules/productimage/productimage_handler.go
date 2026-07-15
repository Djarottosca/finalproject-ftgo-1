package productimage

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

func (h *Handler) Add(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)
	sup, err := h.supplierRepo.FindByUserID(userID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, "you must be a registered supplier to perform this action")
	}

	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid product id")
	}

	var req AddImageRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	res, err := h.service.Add(sup.ID, productID, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrProductNotFound):
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrForbidden):
			return response.ErrorResponse(c, http.StatusForbidden, err.Error())
		default:
			return response.ErrorResponse(c, http.StatusInternalServerError, "failed to add image")
		}
	}

	return response.SuccessResponse(c, http.StatusCreated, "image added", res)
}

func (h *Handler) List(c echo.Context) error {
	productID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid product id")
	}

	res, err := h.service.List(productID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to list images")
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}

func (h *Handler) Delete(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)
	sup, err := h.supplierRepo.FindByUserID(userID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusForbidden, "you must be a registered supplier to perform this action")
	}

	imageID, err := strconv.Atoi(c.Param("imageId"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid image id")
	}

	if err := h.service.Delete(sup.ID, imageID); err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrProductNotFound):
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrForbidden):
			return response.ErrorResponse(c, http.StatusForbidden, err.Error())
		default:
			return response.ErrorResponse(c, http.StatusInternalServerError, "failed to delete image")
		}
	}

	return response.SuccessResponse(c, http.StatusOK, "image deleted", nil)
}
