package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/memory"

	"unipile-connector/internal/adapter/handler"
	"unipile-connector/internal/adapter/middleware"
	"unipile-connector/internal/adapter/repository/postgres"
	"unipile-connector/internal/domain/service"
	"unipile-connector/internal/infrastructure/client"
	"unipile-connector/internal/infrastructure/config"
	"unipile-connector/internal/infrastructure/database"
	"unipile-connector/internal/infrastructure/server"
	"unipile-connector/internal/usecase/account"
	"unipile-connector/internal/usecase/user"
	"unipile-connector/pkg/logger"
)

func main() {
	// Initialize logger
	log := logger.NewLogger()
	log.SetLevel(logrus.InfoLevel)

	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Parse log level
	logLevelAfterLoad, err := logrus.ParseLevel(cfg.Log.Level)
	if err != nil {
		log.Fatalf("Failed to parse log level: %v", err)
	}

	// Connect to database
	db, err := database.Connect(database.Config{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		DBName:   cfg.Database.DBName,
		SSLMode:  cfg.Database.SSLMode,
	})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	// Close database connection
	defer func() {
		if sqlDB, err := db.DB(); err == nil {
			if err := sqlDB.Close(); err != nil {
				log.Errorf("Database close error: %v", err)
			}
		}
	}()

	// Run migrations
	if err := database.RunMigrations(db, log); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Unipile client
	unipileClient := client.NewUnipileClient(cfg.Unipile.BaseURL, cfg.Unipile.APIKey)

	// Test Unipile connection
	if err := unipileClient.TestConnection(); err != nil {
		log.Warnf("Failed to connect to Unipile API: %v", err)
	}

	// Initialize JWT service and middleware
	jwtService := service.NewJWTService(cfg.JWT.SecretKey, cfg.JWT.Issuer)
	jwtMiddleware := middleware.NewJWTMiddleware(jwtService).AuthMiddleware()
	corsMiddleware := middleware.CORSMiddleware(cfg.Server.Host)
	rate, err := limiter.NewRateFromFormatted("5-S")
	if err != nil {
		log.Fatalf("Failed to create rate: %v", err)
	}
	rateLimiter := limiter.New(memory.NewStore(), rate, limiter.WithTrustForwardHeader(true))
	rateLimitMiddleware := mgin.NewMiddleware(rateLimiter)
	middlewares := middleware.NewMiddlewares(corsMiddleware, jwtMiddleware, rateLimitMiddleware)

	// Initialize repositories
	repos := postgres.GetRepositories(db)

	// Initialize use cases
	userUsecase := user.NewUserUsecase(repos.User, jwtService, log)
	accountUsecase := account.NewAccountUsecase(repos.Tx, repos.Account, unipileClient, log)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(userUsecase)
	accountHandler := handler.NewAccountHandler(accountUsecase)
	handlers := handler.NewHandlers(authHandler, accountHandler)

	// Initialize server
	srv := server.NewServer(middlewares, handlers)

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Infof("Starting server on %s", addr)

	// Set log level before running server
	log.SetLevel(logLevelAfterLoad)

	// Start server in a goroutine
	go func() {
		if err := srv.Run(addr); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Gracefully shutdown server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Warn("Gracefully Shutting down application...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("Server shutdown error: %v", err)
	}

	log.Warn("Application exited")
}
