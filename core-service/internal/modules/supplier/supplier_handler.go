package supplier

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

// RegisterRoutes wires supplier routes. authMW authenticates the request;
// review/list are additionally restricted to the admin role.
func (h *Handler) RegisterRoutes(e *echo.Echo, authMW echo.MiddlewareFunc) {
	g := e.Group("/suppliers", authMW)
	g.POST("", h.Register)
	g.GET("/:id", h.Get)
	g.GET("", h.List, middleware.RequireRole("admin"))
	g.PATCH("/:id/review", h.Review, middleware.RequireRole("admin"))
}

func (h *Handler) Register(c echo.Context) error {
	userID, _ := c.Get("user_id").(int)

	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	res, err := h.service.Register(userID, req)
	if err != nil {
		if errors.Is(err, ErrAlreadyExists) {
			return response.ErrorResponse(c, http.StatusConflict, err.Error())
		}
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to register supplier")
	}

	return response.SuccessResponse(c, http.StatusCreated, "supplier registration submitted", res)
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
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to get supplier")
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}

func (h *Handler) List(c echo.Context) error {
	status := c.QueryParam("status")

	res, err := h.service.List(status)
	if err != nil {
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to list suppliers")
	}

	return response.SuccessResponse(c, http.StatusOK, "ok", res)
}

func (h *Handler) Review(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid id")
	}

	var req ReviewRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}

	res, err := h.service.Review(id, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrNotFound):
			return response.ErrorResponse(c, http.StatusNotFound, err.Error())
		case errors.Is(err, ErrInvalidStatus):
			return response.ErrorResponse(c, http.StatusBadRequest, err.Error())
		default:
			return response.ErrorResponse(c, http.StatusInternalServerError, "failed to review supplier")
		}
	}

	return response.SuccessResponse(c, http.StatusOK, "supplier reviewed", res)
}
