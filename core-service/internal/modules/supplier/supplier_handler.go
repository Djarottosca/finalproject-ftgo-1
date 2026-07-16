package supplier

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

// Register handles POST /supplier/register — public self-signup. Creates the
// backing user account (role=supplier) and the supplier profile together,
// no prior /admin/users or /user/register call needed.
func (h *Handler) Register(c echo.Context) error {
	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid request body")
	}
	if err := c.Validate(&req); err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, err.Error())
	}

	res, err := h.service.Register(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, ErrUsernameTaken) {
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

	res, err := h.service.Get(c.Request().Context(), id)
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

	res, err := h.service.List(c.Request().Context(), status)
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

	res, err := h.service.Review(c.Request().Context(), id, req)
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
