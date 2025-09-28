package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"unipile-connector/config"
	"unipile-connector/internal/infrastructure/client"
	"unipile-connector/internal/infrastructure/database"
	"unipile-connector/internal/infrastructure/server"
	"unipile-connector/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	log := logger.NewLogger()
	log.SetLevel(logrus.InfoLevel)

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
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize repositories
	userRepo, accountRepo := database.GetRepositories(db)

	// Initialize Unipile client
	unipileClient := client.NewUnipileClient(cfg.Unipile.BaseURL, cfg.Unipile.APIKey)

	// Test Unipile connection
	if err := unipileClient.TestConnection(); err != nil {
		log.Warnf("Failed to connect to Unipile API: %v", err)
	}

	// Initialize server
	srv := server.NewServer(userRepo, accountRepo, unipileClient, log)

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	log.Infof("Starting server on %s", addr)

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
	log.Info("Shutting down application...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Errorf("Server shutdown error: %v", err)
	}

	log.Info("Application exited")
}
