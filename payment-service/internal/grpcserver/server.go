package grpcserver

import (
	"context"
	"errors"
	"log/slog"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/Djarottosca/finalproject-ftgo-1/payment-service/internal/provider"
	paymentv1 "github.com/Djarottosca/finalproject-ftgo-1/proto/payment/v1"
)

// PaymentServer mengimplementasi paymentv1.PaymentServiceServer.
// Perannya anti-corruption layer: nerjemahin tipe proto <-> tipe provider,
// jadi dunia gRPC dan dunia integrasi gak saling bocor.
type PaymentServer struct {
	paymentv1.UnimplementedPaymentServiceServer
	provider provider.PaymentProvider
	logger   *slog.Logger
}

func NewPaymentServer(p provider.PaymentProvider, logger *slog.Logger) *PaymentServer {
	return &PaymentServer{provider: p, logger: logger}
}

func (s *PaymentServer) CreatePayment(ctx context.Context, req *paymentv1.CreatePaymentRequest) (*paymentv1.CreatePaymentResponse, error) {
	// proto -> provider
	params := provider.CreateInvoiceParams{
		OrderID:        req.GetOrderId(),
		UserID:         req.GetUserId(),
		CustomerName:   req.GetCustomerName(),
		CustomerEmail:  req.GetCustomerEmail(),
		Amount:         req.GetAmount(),
		Description:    req.GetDescription(),
		IdempotencyKey: req.GetIdempotencyKey(),
		Items:          itemsToProvider(req.GetItems()),
	}

	inv, err := s.provider.CreateInvoice(ctx, params)
	if err != nil {
		s.logger.Error("gagal bikin invoice", "order_id", req.GetOrderId(), "err", err)
		return nil, status.Error(codes.Internal, "gagal membuat pembayaran")
	}

	// provider -> proto
	return &paymentv1.CreatePaymentResponse{
		PaymentReference: inv.Reference,
		PaymentLink:      inv.PaymentLink,
		Status:           statusToProto(inv.Status),
		ExpiresAt:        timestamppb.New(inv.ExpiresAt),
	}, nil
}

func (s *PaymentServer) GetPaymentStatus(ctx context.Context, req *paymentv1.GetPaymentStatusRequest) (*paymentv1.GetPaymentStatusResponse, error) {
	inv, err := s.provider.GetInvoice(ctx, req.GetPaymentReference())
	if err != nil {
		if errors.Is(err, provider.ErrInvoiceNotFound) {
			return nil, status.Error(codes.NotFound, "invoice tidak ditemukan")
		}
		s.logger.Error("gagal ambil status invoice", "reference", req.GetPaymentReference(), "err", err)
		return nil, status.Error(codes.Internal, "gagal mengambil status pembayaran")
	}

	return &paymentv1.GetPaymentStatusResponse{
		PaymentReference: inv.Reference,
		OrderId:          inv.OrderID,
		Status:           statusToProto(inv.Status),
	}, nil
}

// statusToProto: provider.Status -> paymentv1.PaymentStatus.
func statusToProto(s provider.Status) paymentv1.PaymentStatus {
	switch s {
	case provider.StatusPending:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_PENDING
	case provider.StatusPaid:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_PAID
	case provider.StatusExpired:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_EXPIRED
	case provider.StatusFailed:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_FAILED
	default:
		return paymentv1.PaymentStatus_PAYMENT_STATUS_UNSPECIFIED
	}
}

func itemsToProvider(items []*paymentv1.PaymentItem) []provider.InvoiceItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]provider.InvoiceItem, 0, len(items))
	for _, it := range items {
		out = append(out, provider.InvoiceItem{
			Name:     it.GetName(),
			Quantity: it.GetQuantity(),
			Price:    it.GetPrice(),
		})
	}
	return out
}
