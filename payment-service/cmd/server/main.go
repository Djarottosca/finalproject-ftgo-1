package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	echo "github.com/labstack/echo/v4"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/Djarottosca/finalproject-ftgo-1/payment-service/internal/config"
	"github.com/Djarottosca/finalproject-ftgo-1/payment-service/internal/grpcserver"
	"github.com/Djarottosca/finalproject-ftgo-1/payment-service/internal/httpserver"
	"github.com/Djarottosca/finalproject-ftgo-1/payment-service/internal/provider"
	paymentv1 "github.com/Djarottosca/finalproject-ftgo-1/proto/payment/v1"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: true}))

	// 1. config
	cfg, err := config.Load()
	if err != nil {
		logger.Error("gagal load config", "err", err)
		os.Exit(1)
	}

	// 2. provider: sekali, instance ini yang di-share ke gRPC & HTTP
	prov, err := provider.New(provider.Config{
		Name:                cfg.Provider.Name,
		BaseURL:             cfg.App.BaseURL,
		XenditAPIKey:        cfg.Xendit.APIKey,
		XenditCallbackToken: cfg.Xendit.CallbackToken,
	})
	if err != nil {
		logger.Error("gagal inisialisasi provider", "err", err)
		os.Exit(1)
	}
	logger.Info("provider aktif", "name", cfg.Provider.Name)

	// 3. gRPC server (dipanggil core)
	grpcServer := grpc.NewServer()
	paymentv1.RegisterPaymentServiceServer(grpcServer, grpcserver.NewPaymentServer(prov, logger))
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", net.JoinHostPort(cfg.App.Host, strconv.Itoa(cfg.App.GRPCPort)))
	if err != nil {
		logger.Error("gagal listen gRPC", "port", cfg.App.GRPCPort, "err", err)
		os.Exit(1)
	}
	go func() {
		logger.Info("gRPC server jalan", "port", cfg.App.GRPCPort)
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("gRPC server berhenti", "err", err)
		}
	}()

	// 4. HTTP server (webhook + simulasi), prov yang SAMA
	e := echo.New()
	e.HideBanner = true
	httpserver.NewServer(prov, logger).Register(e)

	addr := net.JoinHostPort(cfg.App.Host, strconv.Itoa(cfg.App.HTTPPort))
	go func() {
		logger.Info("HTTP server jalan", "addr", addr)
		if err := e.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server berhenti", "err", err)
		}
	}()

	// 5. graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("shutdown diminta, membersihkan")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	grpcServer.GracefulStop()
	if err := e.Shutdown(ctx); err != nil {
		logger.Error("gagal shutdown HTTP", "err", err)
	}
	logger.Info("payment-service berhenti dengan bersih")
}
