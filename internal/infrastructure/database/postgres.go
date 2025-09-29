package database

import (
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"unipile-connector/internal/infrastructure/database/migration"
)

// Config holds database configuration
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// Connect establishes a connection to PostgreSQL
func Connect(config Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host, config.Port, config.User, config.Password, config.DBName, config.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// RunMigrations runs database migrations
func RunMigrations(db *gorm.DB, log *logrus.Logger) error {
	db.Logger.LogMode(logger.Info)

	if err := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		migration.InitialSchema,
	}).Migrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	db.Logger.LogMode(logger.Warn)
	log.Info("Database migrations completed successfully")
	return nil
}
