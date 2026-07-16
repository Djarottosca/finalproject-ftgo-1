package address

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

func (h *Handler) Create(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)

	var req CreateAddressRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	res, err := h.service.Create(c.Request().Context(), userID, req)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to create address")
	}

	return response.SuccessResponse(c, http.StatusCreated, "address created", res)
}

func (h *Handler) List(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)

	res, err := h.service.List(c.Request().Context(), userID)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to list addresses")
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}

func (h *Handler) Update(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid id")
	}

	var req UpdateAddressRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	res, err := h.service.Update(c.Request().Context(), userID, id, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrForbidden):
			return response.ErrorResponse(c, http.StatusForbidden, err.Error())
		default:
			return response.ErrorResponse(c, http.StatusInternalServerError, "failed to update address")
		}
	}

	return response.SuccessResponse(c, http.StatusOK, "address updated", res)
}

func (h *Handler) Delete(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid id")
	}

	if err := h.service.Delete(c.Request().Context(), userID, id); err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrForbidden):
			return response.ErrorResponse(c, http.StatusForbidden, err.Error())
		default:
			return response.ErrorResponse(c, http.StatusInternalServerError, "failed to delete address")
		}
	}

	return response.SuccessResponse(c, http.StatusOK, "address deleted", nil)
}
