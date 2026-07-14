package bootstrap

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	echo "github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	redis "github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/cache"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/config"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/database"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/middleware"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/address"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/auth"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/product"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/productimage"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/supplier"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/user"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/jwt"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/logger"
)

type App struct {
	Config   *config.Config
	Database *gorm.DB
	Redis    *redis.Client
}

func NewApp() *App {
	logger.Init()

	cfg, err := config.Load()
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to load config")
	}

	db, err := database.NewPostgres(cfg.Database)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to connect postgres")
	}

	rdb, err := cache.NewRedis(cfg.Redis)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to connect redis")
	}

	logger.Log.Info().Msg("app bootstrapped")

	return &App{
		Config:   cfg,
		Database: db,
		Redis:    rdb,
	}
}

// RunServer blocks serving the HTTP API.
func (a *App) RunServer() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(echoMiddleware.RequestID())
	e.Use(echoMiddleware.Recover())
	e.Use(middleware.RequestLoggerMiddleware())

	authManager := jwt.NewAuthManager(a.Config.JWTSecret)
	authService := auth.NewService(authManager)
	authHandler := auth.NewHandler(authService)
	authHandler.RegisterRoutes(e)

	userRepo := user.NewRepository(a.Database)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)
	userHandler.RegisterRoutes(e)

	supplierRepo := supplier.NewRepository(a.Database)
	supplierService := supplier.NewService(supplierRepo)
	supplierHandler := supplier.NewHandler(supplierService)
	supplierHandler.RegisterRoutes(e, middleware.AuthMiddleware(authManager))

	productRepo := product.NewRepository(a.Database)
	productService := product.NewService(productRepo)
	productHandler := product.NewHandler(productService, supplierRepo)
	productHandler.RegisterRoutes(e, middleware.AuthMiddleware(authManager))

	productImageRepo := productimage.NewRepository(a.Database)
	productImageService := productimage.NewService(productImageRepo, productRepo)
	productImageHandler := productimage.NewHandler(productImageService, supplierRepo)
	productImageHandler.RegisterRoutes(e, middleware.AuthMiddleware(authManager))

	addressRepo := address.NewRepository(a.Database)
	addressService := address.NewService(addressRepo)
	addressHandler := address.NewHandler(addressService)
	addressHandler.RegisterRoutes(e, middleware.AuthMiddleware(authManager))

	go func() {
		addr := a.Config.App.Host + ":" + strconv.Itoa(a.Config.App.Port)
		logger.Log.Info().Str("addr", addr).Msg("starting server")
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal().Err(err).Msg("server error")
		}
	}()

	<-ctx.Done()

	logger.Log.Info().Msg("shutting down")
	if err := e.Shutdown(context.Background()); err != nil {
		logger.Log.Fatal().Err(err).Msg("server shutdown error")
	}
}

// RunWorker blocks running the Asynq worker.
func (a *App) RunWorker() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logger.Log.Info().Msg("starting worker")

	<-ctx.Done()

	logger.Log.Info().Msg("shutting down")
}
