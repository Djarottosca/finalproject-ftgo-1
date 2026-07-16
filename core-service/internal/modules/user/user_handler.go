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

// Me handles GET /users/me. Returns the caller's own profile.
func (h *Handler) Me(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)

	res, err := h.service.Me(c.Request().Context(), userID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to get profile")
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}

// UpdateMe handles PUT /users/me. Lets the caller edit their own profile.
func (h *Handler) UpdateMe(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)

	var req UpdateProfileRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	res, err := h.service.UpdateMe(c.Request().Context(), userID, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to update profile")
	}

	return response.SuccessResponse(c, http.StatusOK, "profile updated", res)
}

// AdminList handles GET /admin/users.
func (h *Handler) AdminList(c echo.Context) error {
	res, err := h.service.AdminList(c.Request().Context())
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to list users")
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}

// AdminGetByID handles GET /admin/users/:id.
func (h *Handler) AdminGetByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid id")
	}

	res, err := h.service.AdminGetByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to get user")
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}

// AdminUpdate handles PUT /admin/users/:id.
func (h *Handler) AdminUpdate(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid id")
	}

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	res, err := h.service.AdminUpdate(c.Request().Context(), id, req)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to update user")
	}

	return response.SuccessResponse(c, http.StatusOK, "user updated", res)
}

// AdminDelete handles DELETE /admin/users/:id.
func (h *Handler) AdminDelete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid id")
	}

	if err := h.service.AdminDelete(c.Request().Context(), id); err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to delete user")
	}

	return response.SuccessResponse(c, http.StatusOK, "user deleted", nil)
}
