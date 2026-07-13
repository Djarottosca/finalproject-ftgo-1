package auth

import (
	"errors"
	"net/http"

	echo "github.com/labstack/echo/v4"

	"github.com/Djarottosca/finalproject-ftgo-1/pkg/response"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.POST("/auth/login", h.Login)
}

func (h *Handler) Login(c echo.Context) error {
	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	res, err := h.service.Login(req)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			return response.ErrorResponse(c, http.StatusUnauthorized, err.Error())
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to generate token")
	}

	return response.SuccessResponse(c, http.StatusOK, "login successful", res)
}
