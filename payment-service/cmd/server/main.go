package main

import (
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
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/logger"
	paymentv1 "github.com/Djarottosca/finalproject-ftgo-1/proto/payment/v1"
)

func main() {
	logger.Init()

	// 1. config
	cfg, err := config.Load()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to load config")
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
		logger.Log.Fatal().Err(err).Msg("failed to initialize payment provider")
	}
	logger.Log.Info().Str("provider", cfg.Provider.Name).Msg("payment provider initialized")

	// 3. gRPC server — satu-satunya pintu masuk payment-service
	grpcServer := grpc.NewServer()
	paymentv1.RegisterPaymentServiceServer(grpcServer, grpcserver.NewPaymentServer(prov, logger.Log))
	reflection.Register(grpcServer)

	addr := net.JoinHostPort(cfg.App.Host, strconv.Itoa(cfg.App.GRPCPort))
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Log.Fatal().Err(err).Str("addr", addr).Msg("failed to listen on gRPC port")
	}
	go func() {
		logger.Log.Info().Str("addr", addr).Str("env", cfg.App.Env).Msg("gRPC server listening")
		if err := grpcServer.Serve(lis); err != nil {
			logger.Log.Error().Err(err).Msg("gRPC server stopped unexpectedly")
		}
	}()

	// 4. graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Log.Info().Msg("shutdown signal received, shutting down gracefully")

	grpcServer.GracefulStop()
	logger.Log.Info().Msg("payment-service stopped gracefully")
}
