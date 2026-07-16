package product

import (
	"errors"
	"net/http"

	echo "github.com/labstack/echo/v4"

	"github.com/Djarottosca/finalproject-ftgo-1/pkg/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes endpoint publik
func (h *Handler) RegisterRoutes(e *echo.Echo) {
	e.GET("/products", h.List)
	e.GET("/products/:slug", h.Detail)
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
