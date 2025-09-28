package main

import (
	"log"
	"unipile-connector/config"
	"unipile-connector/internal/infrastructure/client"
	"unipile-connector/internal/infrastructure/database"
	"unipile-connector/internal/infrastructure/server"
	"unipile-connector/pkg/logger"

	"github.com/sirupsen/logrus"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
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
	if err := srv.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

