package user

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
	var req CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	res, err := h.service.Create(c.Request().Context(), req)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to create user")
	}

	return response.SuccessResponse(c, http.StatusCreated, "user created", res)
}

func (h *Handler) List(c echo.Context) error {
	res, err := h.service.List(c.Request().Context())
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to list users")
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
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to get user")
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}

func (h *Handler) Update(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid id")
	}

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	res, err := h.service.Update(c.Request().Context(), id, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to update user")
	}

	return response.SuccessResponse(c, http.StatusOK, "user updated", res)
}

func (h *Handler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid id")
	}

	if err := h.service.Delete(c.Request().Context(), id); err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to delete user")
	}

	return response.SuccessResponse(c, http.StatusOK, "user deleted", nil)
}
