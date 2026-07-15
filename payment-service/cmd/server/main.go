package main

import (
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/Djarottosca/finalproject-ftgo-1/payment-service/internal/config"
	"github.com/Djarottosca/finalproject-ftgo-1/payment-service/internal/grpcserver"
	"github.com/Djarottosca/finalproject-ftgo-1/payment-service/internal/provider"
	paymentv1 "github.com/Djarottosca/finalproject-ftgo-1/proto/payment/v1"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))

	// 1. config
	cfg, err := config.Load()
	if err != nil {
		logger.Error("failed to load config", "err", err)
		os.Exit(1)
	}

	// 2. provider (simulation / xendit)
	prov, err := provider.New(provider.Config{
		Name:              cfg.Provider.Name,
		BaseURL:           cfg.App.BaseURL,
		InvoiceTTL:        cfg.Provider.InvoiceTTL,
		XenditAPIKey:      cfg.Xendit.APIKey,
		XenditBaseURL:     cfg.Xendit.BaseURL,
		XenditReturnURL:   cfg.Xendit.ReturnURL,
		XenditHTTPTimeout: cfg.Xendit.HTTPTimeout,
		XenditCurrency:    cfg.Xendit.Currency,
		XenditCountry:     cfg.Xendit.Country,
	})
	if err != nil {
		logger.Error("failed to initialize payment provider", "err", err)
		os.Exit(1)
	}
	logger.Info("payment provider initialized", "provider", cfg.Provider.Name)

	// 3. gRPC server — satu-satunya pintu masuk payment-service
	grpcServer := grpc.NewServer()
	paymentv1.RegisterPaymentServiceServer(grpcServer, grpcserver.NewPaymentServer(prov, logger))
	reflection.Register(grpcServer)

	addr := net.JoinHostPort(cfg.App.Host, strconv.Itoa(cfg.App.GRPCPort))
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("failed to listen on gRPC port", "addr", addr, "err", err)
		os.Exit(1)
	}
	go func() {
		logger.Info("gRPC server listening", "addr", addr)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server stopped unexpectedly", "err", err)
		}
	}()

	// 4. graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutdown signal received, shutting down gracefully")

	grpcServer.GracefulStop()
	logger.Info("payment-service stopped gracefully")
}
