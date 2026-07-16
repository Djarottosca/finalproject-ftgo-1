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
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/gorm"

	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/cache"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/config"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/database"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/grpcclient"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/middleware"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/address"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/admin"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/auth"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/cart"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/order"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/payment"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/product"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/productimage"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/supplier"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/user"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/jwt"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/logger"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/validator"
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
	e.Validator = validator.New()

	e.Use(echoMiddleware.RequestID())
	e.Use(echoMiddleware.Recover())
	e.Use(middleware.RequestLoggerMiddleware())

	// gRPC connection to payment-service. grpc.NewClient is lazy: it does not
	// dial until the first RPC, so core can start even if payment-service is
	// down, and reconnects on its own. Closed on shutdown.
	paymentConn, err := grpc.NewClient(
		a.Config.PaymentServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Log.Fatal().Err(err).Msg("failed to init payment-service client")
	}
	defer paymentConn.Close()
	paymentClient := grpcclient.NewPaymentClient(paymentConn)

	authManager := jwt.NewAuthManager(a.Config.JWTSecret)
	authMW := middleware.AuthMiddleware(authManager)
	adminMW := middleware.RequireRole("admin")
	supplierMW := middleware.RequireRole("supplier")

	userRepo := user.NewRepository(a.Database)
	userHandler := user.NewHandler(user.NewService(userRepo))

	authHandler := auth.NewHandler(auth.NewService(userRepo, authManager))

	supplierRepo := supplier.NewRepository(a.Database)
	supplierHandler := supplier.NewHandler(supplier.NewService(supplierRepo))

	productRepo := product.NewRepository(a.Database)
	productCache := product.NewRedisCache(a.Redis)
	productHandler := product.NewHandler(product.NewService(productRepo, productCache), supplierRepo)

	productImageRepo := productimage.NewRepository(a.Database)
	productImageHandler := productimage.NewHandler(productimage.NewService(productImageRepo, productRepo), supplierRepo)

	addressRepo := address.NewRepository(a.Database)
	addressHandler := address.NewHandler(address.NewService(addressRepo))

	cartRepo := cart.NewRepository(a.Database)
	cartHandler := cart.NewHandler(cart.NewService(cartRepo, productRepo))

	orderRepo := order.NewRepository(a.Database)
	orderHandler := order.NewHandler(order.NewService(orderRepo, cartRepo), supplierRepo)

	paymentRepo := payment.NewRepository(a.Database)
	paymentHandler := payment.NewHandler(payment.NewService(paymentRepo, paymentClient))

	adminRepo := admin.NewRepository(a.Database)
	adminHandler := admin.NewHandler(admin.NewService(adminRepo))

	v1 := e.Group("/api/v1")

	// public routes
	v1.POST("/auth/login", authHandler.Login)
	v1.POST("/users", userHandler.Create)
	v1.POST("/suppliers", supplierHandler.Register, authMW)
	v1.GET("/products", productHandler.List)
	v1.GET("/products/:slug", productHandler.Detail)

	// user routes
	v1.GET("/users", userHandler.List, authMW)
	v1.GET("/users/:id", userHandler.Get, authMW)
	v1.PUT("/users/:id", userHandler.Update, authMW)
	v1.DELETE("/users/:id", userHandler.Delete, authMW)
	v1.GET("/suppliers/:id", supplierHandler.Get, authMW)
	v1.POST("/addresses", addressHandler.Create, authMW)
	v1.GET("/addresses", addressHandler.List, authMW)
	v1.PUT("/addresses/:id", addressHandler.Update, authMW)
	v1.DELETE("/addresses/:id", addressHandler.Delete, authMW)
	v1.GET("/cart", cartHandler.GetCart, authMW)
	v1.POST("/cart/items", cartHandler.AddItem, authMW)
	v1.PUT("/cart/items/:product_id", cartHandler.UpdateItem, authMW)
	v1.DELETE("/cart/items/:product_id", cartHandler.RemoveItem, authMW)
	v1.POST("/orders/checkout", orderHandler.Checkout, authMW)
	v1.GET("/orders", orderHandler.ListMine, authMW)
	v1.GET("/orders/:id", orderHandler.Get, authMW)
	v1.POST("/payments", paymentHandler.Create, authMW)
	v1.GET("/payments/:orderId", paymentHandler.GetStatus, authMW)

	// supplier routes
	v1.GET("/supplier/products", productHandler.ListMine, authMW, supplierMW)
	v1.POST("/supplier/products", productHandler.Create, authMW, supplierMW)
	v1.PUT("/supplier/products/:id", productHandler.Update, authMW, supplierMW)
	v1.PATCH("/supplier/products/:id/discount", productHandler.SetDiscount, authMW, supplierMW)
	v1.PATCH("/supplier/products/:id/stock", productHandler.AdjustStock, authMW, supplierMW)
	v1.DELETE("/supplier/products/:id", productHandler.Delete, authMW, supplierMW)
	v1.POST("/supplier/products/:id/images", productImageHandler.Add, authMW, supplierMW)
	v1.GET("/supplier/products/:id/images", productImageHandler.List, authMW, supplierMW)
	v1.DELETE("/supplier/product-images/:imageId", productImageHandler.Delete, authMW, supplierMW)
	v1.GET("/supplier/orders", orderHandler.ListForSupplier, authMW, supplierMW)
	v1.PATCH("/supplier/orders/:id/status", orderHandler.UpdateStatus, authMW, supplierMW)

	// admin routes
	v1.GET("/admin/suppliers", supplierHandler.List, authMW, adminMW)
	v1.PATCH("/admin/suppliers/:id/review", supplierHandler.Review, authMW, adminMW)
	v1.GET("/admin/reports/stock", adminHandler.StockReport, authMW, adminMW)

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
