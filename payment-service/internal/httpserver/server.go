package httpserver

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"

	echo "github.com/labstack/echo/v4"

	"github.com/Djarottosca/finalproject-ftgo-1/payment-service/internal/provider"
)

// Server nampung semua pintu masuk HTTP ke payment-service:
// webhook Xendit (produksi) dan pay-endpoint simulasi (demo).
// Dua-duanya berujung di satu jalur yang sama: handlePaidEvent.
type Server struct {
	provider provider.PaymentProvider
	sim      *provider.Simulation // non-nil cuma kalau provider aktif = simulasi
	logger   *slog.Logger
}

func NewServer(p provider.PaymentProvider, logger *slog.Logger) *Server {
	s := &Server{provider: p, logger: logger}
	// Type-assert SEKALI di sini, bukan tiap request. MarkPaid ada di luar
	// interface PaymentProvider, jadi cuma kesimpen kalau provider-nya emang simulasi.
	if sim, ok := p.(*provider.Simulation); ok {
		s.sim = sim
	}
	return s
}

// Register nempelin route ke Echo instance yang dibikin di cmd/server.
func (s *Server) Register(e *echo.Echo) {
	e.POST("/webhook/xendit", s.handleXenditWebhook)
	e.GET("/simulation/pay", s.handleSimulationPay)
}

// Pintu produksi: Xendit POST ke sini pas user selesai bayar.
func (s *Server) handleXenditWebhook(c echo.Context) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "gagal baca body"})
	}
	signature := c.Request().Header.Get("x-callback-token")

	event, err := s.provider.ParseWebhook(body, signature)
	if err != nil {
		s.logger.Warn("webhook xendit ditolak", "err", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "webhook tidak valid"})
	}

	s.handlePaidEvent(c.Request().Context(), event)
	// Xendit cuma peduli status 2xx = "gue udah nerima". Kalau bukan 2xx, dia retry.
	return c.NoContent(http.StatusOK)
}

// Pintu demo: dibuka manual di browser, GET /simulation/pay?ref=xxx
func (s *Server) handleSimulationPay(c echo.Context) error {
	if s.sim == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "endpoint simulasi tidak aktif"})
	}
	ref := c.QueryParam("ref")
	if ref == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "ref wajib diisi"})
	}

	event, err := s.sim.MarkPaid(ref)
	if err != nil {
		if errors.Is(err, provider.ErrInvoiceNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "invoice tidak ditemukan"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "gagal proses pembayaran"})
	}

	s.handlePaidEvent(c.Request().Context(), event)
	return c.JSON(http.StatusOK, map[string]string{
		"message": "pembayaran simulasi berhasil",
		"ref":     ref,
	})
}

// Titik konvergensi. Dari webhook Xendit MAUPUN simulasi, semuanya nyampe sini
// dalam bentuk provider.WebhookEvent yang identik. Dari sini "paid" dikabarin ke core.
func (s *Server) handlePaidEvent(ctx context.Context, event provider.WebhookEvent) {
	s.logger.Info("paid event diterima",
		"order_id", event.OrderID,
		"reference", event.Reference,
		"status", event.Status,
	)

	// TODO(callback): kabarin core kalau order ini sudah paid.
	// Nunggu keputusan transport dari tim core:
	//   - gRPC-both  : coreClient.ConfirmPayment(ctx, event.OrderID, event.Reference)
	//   - REST intern: POST core /internal/payments/confirm
	// Sampai itu ketok, event cukup di-log dulu biar flow demo tetap keliatan jalan.
}
