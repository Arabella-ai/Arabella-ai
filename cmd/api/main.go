package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/arabella/ai-studio-backend/config"
	"github.com/arabella/ai-studio-backend/internal/infrastructure/auth"
	"github.com/arabella/ai-studio-backend/internal/infrastructure/cache"
	"github.com/arabella/ai-studio-backend/internal/infrastructure/database"
	"github.com/arabella/ai-studio-backend/internal/infrastructure/provider"
	"github.com/arabella/ai-studio-backend/internal/infrastructure/queue"
	infraRepo "github.com/arabella/ai-studio-backend/internal/infrastructure/repository"
	"github.com/arabella/ai-studio-backend/internal/interface/http/handler"
	"github.com/arabella/ai-studio-backend/internal/interface/http/middleware"
	"github.com/arabella/ai-studio-backend/internal/interface/websocket"
	"github.com/arabella/ai-studio-backend/internal/usecase"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	_ "github.com/arabella/ai-studio-backend/docs"
)

// Version and BuildTime are set during build
var (
	Version   = "dev"
	BuildTime = "unknown"
)

// @title Arabella API
// @version 1.0
// @description AI Video Generation Platform API
// @termsOfService https://arabella.app/terms

// @contact.name API Support
// @contact.url https://arabella.app/support
// @contact.email support@arabella.app

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:8080
// @BasePath /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter the token with the `Bearer ` prefix
func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := initLogger(cfg)
	defer logger.Sync()

	logger.Info("Starting Arabella API",
		zap.String("version", Version),
		zap.String("build_time", BuildTime),
		zap.String("environment", string(cfg.App.Environment)),
	)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize database
	dbConfig := database.PostgresConfig{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		Database:        cfg.Database.Database,
		SSLMode:         cfg.Database.SSLMode,
		MaxConnections:  cfg.Database.MaxConnections,
		MinConnections:  cfg.Database.MinConnections,
		MaxConnLifetime: cfg.Database.MaxConnLifetime,
		MaxConnIdleTime: cfg.Database.MaxConnIdleTime,
	}

	db, err := database.NewPostgresDB(ctx, dbConfig, logger)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	// Initialize Redis
	redisConfig := cache.RedisConfig{
		Host:         cfg.Redis.Host,
		Port:         cfg.Redis.Port,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	}

	redisCache, err := cache.NewRedisCache(ctx, redisConfig, logger)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisCache.Close()

	// Initialize repositories
	userRepo := infraRepo.NewUserRepositoryPostgres(db.Pool())
	templateRepo := infraRepo.NewTemplateRepositoryPostgres(db.Pool())
	videoJobRepo := infraRepo.NewVideoJobRepositoryPostgres(db.Pool())

	// Initialize rate limiter
	rateLimiter := cache.NewRateLimiter(redisCache.Client())

	// Initialize job queue
	jobQueue := queue.NewRedisQueue(redisCache.Client(), logger)

	// Initialize AI providers
	providerRegistry := provider.NewProviderRegistry(logger)

	if cfg.AI.UseMockProvider {
		mockProvider := provider.NewMockProvider(logger, false)
		providerRegistry.Register(mockProvider)
	}

	if cfg.AI.GeminiAPIKey != "" {
		geminiProvider := provider.NewGeminiProvider(cfg.AI.GeminiAPIKey, logger)
		providerRegistry.Register(geminiProvider)
	}

	providerSelector := provider.NewProviderSelector(providerRegistry, logger)

	// Initialize WebSocket hub
	wsHub := websocket.NewHub(logger)
	go wsHub.Run()

	// Initialize auth components
	jwtConfig := auth.JWTConfig{
		SecretKey:            cfg.JWT.SecretKey,
		AccessTokenDuration:  cfg.JWT.AccessTokenDuration,
		RefreshTokenDuration: cfg.JWT.RefreshTokenDuration,
		Issuer:               cfg.JWT.Issuer,
	}
	tokenGenerator := auth.NewJWTTokenGenerator(jwtConfig)

	googleConfig := auth.GoogleAuthConfig{
		ClientID:     cfg.Google.ClientID,
		ClientSecret: cfg.Google.ClientSecret,
	}
	googleVerifier := auth.NewGoogleAuthVerifier(googleConfig, logger)

	// Initialize use cases
	authUseCase := usecase.NewAuthUseCase(userRepo, tokenGenerator, googleVerifier)
	templateUseCase := usecase.NewTemplateUseCase(templateRepo, redisCache)
	userUseCase := usecase.NewUserUseCase(userRepo, videoJobRepo)
	videoUseCase := usecase.NewVideoUseCase(
		videoJobRepo,
		templateRepo,
		userRepo,
		providerSelector,
		jobQueue,
		wsHub,
	)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authUseCase)
	templateHandler := handler.NewTemplateHandler(templateUseCase)
	userHandler := handler.NewUserHandler(userUseCase)
	videoHandler := handler.NewVideoHandler(videoUseCase)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(authUseCase)
	rateLimitMiddleware := middleware.NewRateLimitMiddleware(rateLimiter)
	loggingMiddleware := middleware.NewLoggingMiddleware(logger)

	// Initialize WebSocket handler
	wsHandler := websocket.NewHandler(wsHub, authUseCase, logger)

	// Setup router
	router := setupRouter(cfg, logger, authHandler, templateHandler, userHandler, videoHandler,
		authMiddleware, rateLimitMiddleware, loggingMiddleware, wsHandler)

	// Create HTTP server
	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Server starting",
			zap.String("address", server.Addr),
		)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed to start", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped gracefully")
}

// initLogger initializes the zap logger
func initLogger(cfg *config.Config) *zap.Logger {
	var zapConfig zap.Config

	if cfg.IsDevelopment() {
		zapConfig = zap.NewDevelopmentConfig()
		zapConfig.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		zapConfig = zap.NewProductionConfig()
		zapConfig.EncoderConfig.TimeKey = "timestamp"
		zapConfig.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	}

	logger, err := zapConfig.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}

	return logger
}

// setupRouter configures the Gin router with all routes and middleware
func setupRouter(
	cfg *config.Config,
	logger *zap.Logger,
	authHandler *handler.AuthHandler,
	templateHandler *handler.TemplateHandler,
	userHandler *handler.UserHandler,
	videoHandler *handler.VideoHandler,
	authMiddleware *middleware.AuthMiddleware,
	rateLimitMiddleware *middleware.RateLimitMiddleware,
	loggingMiddleware *middleware.LoggingMiddleware,
	wsHandler *websocket.Handler,
) *gin.Engine {
	// Set Gin mode
	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.RequestID())
	router.Use(loggingMiddleware.Logger())
	router.Use(loggingMiddleware.Recovery())
	router.Use(middleware.CORS(middleware.DefaultCORSConfig()))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"version": Version,
			"time":    time.Now().UTC().Format(time.RFC3339),
		})
	})

	// Swagger documentation (development only)
	if cfg.IsDevelopment() {
		router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	}

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Rate limiting for all API routes
		v1.Use(rateLimitMiddleware.Limit(100, time.Minute))

		// Auth routes (public)
		authRoutes := v1.Group("/auth")
		{
			authRoutes.POST("/google", authHandler.GoogleAuth)
			authRoutes.POST("/refresh", authHandler.RefreshToken)
			authRoutes.POST("/logout", authMiddleware.RequireAuth(), authHandler.Logout)
		}

		// Template routes (public, optional auth)
		templateRoutes := v1.Group("/templates")
		{
			templateRoutes.GET("", templateHandler.ListTemplates)
			templateRoutes.GET("/popular", templateHandler.GetPopularTemplates)
			templateRoutes.GET("/categories", templateHandler.GetCategories)
			templateRoutes.GET("/category/:category", templateHandler.GetTemplatesByCategory)
			templateRoutes.GET("/:id", templateHandler.GetTemplate)
		}

		// Video routes (authenticated)
		videoRoutes := v1.Group("/videos")
		videoRoutes.Use(authMiddleware.RequireAuth())
		{
			videoRoutes.POST("/generate", rateLimitMiddleware.LimitGeneration(), videoHandler.GenerateVideo)
			videoRoutes.GET("", videoHandler.ListUserVideos)
			videoRoutes.GET("/recent", videoHandler.GetRecentVideos)
			videoRoutes.GET("/:id", videoHandler.GetVideo)
			videoRoutes.GET("/:id/status", videoHandler.GetJobStatus)
			videoRoutes.POST("/:id/cancel", videoHandler.CancelJob)
		}

		// User routes (authenticated)
		userRoutes := v1.Group("/user")
		userRoutes.Use(authMiddleware.RequireAuth())
		{
			userRoutes.GET("/profile", userHandler.GetProfile)
			userRoutes.PUT("/profile", userHandler.UpdateProfile)
			userRoutes.GET("/credits", userHandler.GetCredits)
			userRoutes.DELETE("/account", userHandler.DeleteAccount)
		}

		// Subscription routes (authenticated)
		subscriptionRoutes := v1.Group("/subscriptions")
		subscriptionRoutes.Use(authMiddleware.RequireAuth())
		{
			subscriptionRoutes.POST("", userHandler.UpgradeSubscription)
		}

		// WebSocket routes
		wsRoutes := v1.Group("/ws")
		{
			wsRoutes.GET("/videos/:id", wsHandler.HandleJobConnection)
		}
	}

	return router
}

