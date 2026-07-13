package main

import (
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/Djarottosca/finalproject-ftgo-1/notification-service/internal/config"
	"github.com/Djarottosca/finalproject-ftgo-1/notification-service/internal/grpcserver"
	"github.com/Djarottosca/finalproject-ftgo-1/notification-service/internal/provider"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/logger"
	notificationv1 "github.com/Djarottosca/finalproject-ftgo-1/proto/notification/v1"
)

func main() {
	logger.Init()

	cfg, err := config.Load()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to load config")
	}

	mailjetProvider := provider.NewMailjetProvider(cfg.Mailjet)

	notifServer := grpcserver.NewServer(mailjetProvider, logger.Log)

	grpcServer := grpc.NewServer()
	notificationv1.RegisterNotificationServiceServer(grpcServer, notifServer)

	reflection.Register(grpcServer)

	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Log.Fatal().Err(err).Str("addr", addr).Msg("failed to listen")
	}

	logger.Log.Info().Str("addr", addr).Str("env", cfg.App.Env).Msg("notification-service listening (grpc + reflection)")

	if err := grpcServer.Serve(lis); err != nil {
		logger.Log.Fatal().Err(err).Msg("grpc server error")
	}
}
