package webhook

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	echo "github.com/labstack/echo/v4"

	"github.com/Djarottosca/finalproject-ftgo-1/pkg/logger"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/response"
)

type Handler struct {
	service       Service
	callbackToken string
}

// NewHandler: callbackToken is the verification token from the Xendit
// dashboard (XENDIT_WEBHOOK_TOKEN), not the XENDIT_API_KEY.
func NewHandler(service Service, callbackToken string) *Handler {
	return &Handler{service: service, callbackToken: callbackToken}
}

// xenditWebhookPayload is Xendit's outer envelope; Data varies per event
// type, hence json.RawMessage kept raw until Event confirms the shape.
type xenditWebhookPayload struct {
	Event string          `json:"event"`
	Data  json.RawMessage `json:"data"`
}

type paymentSessionPayload struct {
	PaymentSessionID string `json:"payment_session_id"`
	Status           string `json:"status"`
	ReferenceID      string `json:"reference_id"`
}

// XenditWebhook handles POST /webhooks/xendit. Called by Xendit directly, so
// there's no JWT — the x-callback-token header is the only auth, compared
// against the token configured on the Xendit dashboard's webhook page.
func (h *Handler) XenditWebhook(c echo.Context) error {
	if h.callbackToken == "" || c.Request().Header.Get("x-callback-token") != h.callbackToken {
		logger.Log.Warn().Str("ip", c.RealIP()).Msg("xendit webhook: invalid callback token")
		return response.ErrorResponse(c, http.StatusUnauthorized, "invalid callback token")
	}

	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return response.ErrorResponse(c, http.StatusBadRequest, "failed to read body")
	}

	var payload xenditWebhookPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		logger.Log.Error().Err(err).Str("body", string(body)).Msg("xendit webhook: failed to parse")
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid payload")
	}

	// Only payment_session.* is wired up (that's the API our provider uses).
	// Ack other event types with 200 anyway so Xendit stops retrying them.
	if !strings.HasPrefix(payload.Event, "payment_session.") {
		return response.SuccessResponse(c, http.StatusOK, "ignored", nil)
	}

	var data paymentSessionPayload
	if err := json.Unmarshal(payload.Data, &data); err != nil {
		logger.Log.Error().Err(err).Str("raw_data", string(payload.Data)).Msg("xendit webhook: failed to unmarshal data")
		return response.ErrorResponse(c, http.StatusBadRequest, "invalid data")
	}

	logger.Log.Info().
		Str("event", payload.Event).
		Str("session_id", data.PaymentSessionID).
		Str("status", data.Status).
		Msg("xendit webhook received")

	if err := h.service.HandlePaymentSession(c.Request().Context(), data.PaymentSessionID, data.Status); err != nil {
		if errors.Is(err, ErrPaymentNotFound) {
			logger.Log.Warn().Str("session_id", data.PaymentSessionID).Msg("xendit webhook: payment not found")
			return response.ErrorResponse(c, http.StatusNotFound, "payment not found")
		}
		logger.Log.Error().Err(err).Str("session_id", data.PaymentSessionID).Msg("xendit webhook: failed to process")
		return response.ErrorResponse(c, http.StatusInternalServerError, "failed to process webhook")
	}

	return response.SuccessResponse(c, http.StatusOK, "webhook processed", nil)
}
