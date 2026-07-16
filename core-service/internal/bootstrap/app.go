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
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/external/rajaongkir"
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
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/shipping"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/supplier"
	"github.com/Djarottosca/finalproject-ftgo-1/core-service/internal/modules/user"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/jwt"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/logger"
	"github.com/Djarottosca/finalproject-ftgo-1/pkg/validator"
)

type App struct {
	Config             *config.Config
	Database           *gorm.DB
	Redis              *redis.Client
	NotificationConn   *grpc.ClientConn
	NotificationClient *grpcclient.NotificationClient
	PaymentConn        *grpc.ClientConn
	PaymentClient      *grpcclient.PaymentClient
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

	// jalan di jaringan internal/dev, belum ada TLS antar service.
	notifConn, err := grpc.NewClient(
		cfg.NotificationServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Log.Fatal().Err(err).Str("addr", cfg.NotificationServiceAddr).Msg("failed to connect notification-service")
	}
	notifClient := grpcclient.NewNotificationClient(notifConn)

	payConn, err := grpc.NewClient(
		cfg.PaymentServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logger.Log.Fatal().Err(err).Str("addr", cfg.PaymentServiceAddr).Msg("failed to connect payment-service")
	}
	payClient := grpcclient.NewPaymentClient(payConn)

	logger.Log.Info().Msg("app bootstrapped")

	return &App{
		Config:             cfg,
		Database:           db,
		Redis:              rdb,
		NotificationConn:   notifConn,
		NotificationClient: notifClient,
		PaymentConn:        payConn,
		PaymentClient:      payClient,
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

	authManager := jwt.NewAuthManager(a.Config.JWTSecret)
	authMW := middleware.AuthMiddleware(authManager)
	adminMW := middleware.RequireRole("admin")
	supplierMW := middleware.RequireRole("supplier")

	userRepo := user.NewRepository(a.Database)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	authService := auth.NewService(userRepo, authManager)
	authHandler := auth.NewHandler(authService)

	supplierRepo := supplier.NewRepository(a.Database)
	supplierService := supplier.NewService(supplierRepo)
	supplierHandler := supplier.NewHandler(supplierService)

	productRepo := product.NewRepository(a.Database)
	productCache := product.NewRedisCache(a.Redis)
	productService := product.NewService(productRepo, productCache)
	productHandler := product.NewHandler(productService, supplierRepo)

	productImageRepo := productimage.NewRepository(a.Database)
	productImageService := productimage.NewService(productImageRepo, productRepo)
	productImageHandler := productimage.NewHandler(productImageService, supplierRepo)

	addressRepo := address.NewRepository(a.Database)
	addressService := address.NewService(addressRepo)
	addressHandler := address.NewHandler(addressService)

	cartRepo := cart.NewRepository(a.Database)
	cartService := cart.NewService(cartRepo, productRepo)
	cartHandler := cart.NewHandler(cartService)

	shippingRepo := shipping.NewRepository(a.Database)

	orderRepo := order.NewRepository(a.Database)
	orderService := order.NewService(orderRepo, cartRepo, userRepo, shippingRepo, a.NotificationClient)
	orderHandler := order.NewHandler(orderService, supplierRepo)

	paymentRepo := payment.NewRepository(a.Database)
	paymentService := payment.NewService(paymentRepo, a.PaymentClient)
	paymentHandler := payment.NewHandler(paymentService)

	adminRepo := admin.NewRepository(a.Database)
	adminService := admin.NewService(adminRepo)
	adminHandler := admin.NewHandler(adminService)

	rajaOngkirClient := rajaongkir.NewHttpClient(&a.Config.RajaOngkir)
	shippingService := shipping.NewService(rajaOngkirClient)
	shippingHandler := shipping.NewHandler(shippingService)

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
	v1.GET("/shipping/destinations", shippingHandler.SearchDestinations, authMW)
	v1.POST("/shipping/cost", shippingHandler.CalculateCost, authMW)

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
	v1.GET("/admin/reports/sales", adminHandler.SalesReport, authMW, adminMW)
	v1.GET("/admin/transactions", adminHandler.Transactions, authMW, adminMW)

	go func() {
		addr := a.Config.App.Host + ":" + strconv.Itoa(a.Config.App.Port)
		logger.Log.Info().Str("addr", addr).Str("env", a.Config.App.Env).Msg("starting server")
		if err := e.Start(addr); err != nil && err != http.ErrServerClosed {
			logger.Log.Fatal().Err(err).Msg("server error")
		}
	}()

	<-ctx.Done()

	logger.Log.Info().Msg("shutting down")
	if err := e.Shutdown(context.Background()); err != nil {
		logger.Log.Fatal().Err(err).Msg("server shutdown error")
	}
	if a.NotificationConn != nil {
		if err := a.NotificationConn.Close(); err != nil {
			logger.Log.Warn().Err(err).Msg("failed to close notification-service connection")
		}
	}
	if a.PaymentConn != nil {
		if err := a.PaymentConn.Close(); err != nil {
			logger.Log.Warn().Err(err).Msg("failed to close payment-service connection")
		}
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
