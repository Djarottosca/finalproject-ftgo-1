package auth

import (
	"errors"
	"net/http"

	echo "github.com/labstack/echo/v4"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/models"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) login(c echo.Context, role string) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	res, err := h.service.Login(c.Request().Context(), req, role)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return response.ErrorResponse(c, http.StatusUnauthorized, err.Error())
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to generate token")
	}

	return response.SuccessResponse(c, http.StatusOK, "login successful", res)
}

// UserLogin handles POST /user/login.
func (h *Handler) UserLogin(c echo.Context) error { return h.login(c, models.RoleUser) }

// SupplierLogin handles POST /supplier/login.
func (h *Handler) SupplierLogin(c echo.Context) error { return h.login(c, models.RoleSupplier) }

// AdminLogin handles POST /admin/login.
func (h *Handler) AdminLogin(c echo.Context) error { return h.login(c, models.RoleAdmin) }

// UserRegister handles POST /user/register. Self-signup, always creates a
// "user"-role account. Suppliers register via POST /supplier/register
// (supplier module, creates user+supplier together); admins are seeded.
func (h *Handler) UserRegister(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	res, err := h.service.RegisterUser(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, ErrUsernameTaken) {
			return response.ErrorResponse(c, http.StatusConflict, err.Error())
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to register")
	}

	return response.SuccessResponse(c, http.StatusCreated, "registered", res)
}

// Logout handles POST /{user,supplier,admin}/logout — same handler for all
// three portals since it just revokes whatever token is in the request.
// Requires authMW to have already verified the token and stashed it in
// context.
func (h *Handler) Logout(c echo.Context) error {
	token, _ := c.Get("raw_token").(string)

	if err := h.service.Logout(c.Request().Context(), token); err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to logout")
	}

	return response.SuccessResponse(c, http.StatusOK, "logout successful", nil)
}
