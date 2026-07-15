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
	sim      *provider.Simulation // non-nil cuma kalau provider aktif = simulasi
	logger   *slog.Logger
}

func NewPaymentServer(p provider.PaymentProvider, logger *slog.Logger) *PaymentServer {
	s := &PaymentServer{provider: p, logger: logger}
	// Type-assert sekali di sini. MarkPaid ada di LUAR interface PaymentProvider,
	// jadi cuma kesimpen kalau provider-nya emang simulasi.
	if sim, ok := p.(*provider.Simulation); ok {
		s.sim = sim
	}
	return s
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
		s.logger.Error("failed to create payment", "order_id", req.GetOrderId(), "err", err)
		return nil, status.Error(codes.Internal, "failed to create payment")
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
			return nil, status.Error(codes.NotFound, "invoice not found")
		}
		s.logger.Error("failed to get payment status", "reference", req.GetPaymentReference(), "err", err)
		return nil, status.Error(codes.Internal, "failed to get payment status")
	}

	return &paymentv1.GetPaymentStatusResponse{
		PaymentReference: inv.Reference,
		OrderId:          inv.OrderID,
		Status:           statusToProto(inv.Status),
	}, nil
}

// SimulatePayment forces an invoice into PAID without going through Xendit.
// It only works when the active provider is the simulation; in production
// (Xendit) it is rejected.
func (s *PaymentServer) SimulatePayment(ctx context.Context, req *paymentv1.SimulatePaymentRequest) (*paymentv1.SimulatePaymentResponse, error) {
	if s.sim == nil {
		return nil, status.Error(codes.FailedPrecondition, "simulation is not active: current provider is not simulation")
	}

	inv, err := s.sim.MarkPaid(req.GetPaymentReference())
	if err != nil {
		if errors.Is(err, provider.ErrInvoiceNotFound) {
			return nil, status.Error(codes.NotFound, "invoice not found")
		}
		s.logger.Error("failed to simulate payment", "reference", req.GetPaymentReference(), "err", err)
		return nil, status.Error(codes.Internal, "failed to simulate payment")
	}

	s.logger.Info("payment simulated",
		"order_id", inv.OrderID,
		"reference", inv.Reference,
	)

	return &paymentv1.SimulatePaymentResponse{
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
