package database

import (
	"context"
	"fmt"
	"log"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	postgresRepo "unipile-connector/internal/adapter/repository/postgres"
	"unipile-connector/internal/domain/repository"
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
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// RunMigrations runs database migrations
func RunMigrations(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "001_initial_schema",
			Migrate: func(tx *gorm.DB) error {
				// Create users table
				if err := tx.Exec(`
					CREATE TABLE IF NOT EXISTS users (
						id SERIAL PRIMARY KEY,
						username VARCHAR(255) UNIQUE NOT NULL,
						email VARCHAR(255) UNIQUE NOT NULL,
						password VARCHAR(255) NOT NULL,
						created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						deleted_at TIMESTAMP NULL
					);
				`).Error; err != nil {
					return err
				}

				// Create accounts table
				if err := tx.Exec(`
					CREATE TABLE IF NOT EXISTS accounts (
						id SERIAL PRIMARY KEY,
						user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
						provider VARCHAR(100) NOT NULL,
						account_id VARCHAR(255) NOT NULL,
						current_status VARCHAR(100) NOT NULL,
						created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						deleted_at TIMESTAMP NULL
					);
				`).Error; err != nil {
					return err
				}

				// Create accounts_status_history table
				if err := tx.Exec(`
					CREATE TABLE IF NOT EXISTS account_status_histories (
						id SERIAL PRIMARY KEY,
						account_id INTEGER NOT NULL REFERENCES accounts(id) ON DELETE CASCADE,
						checkpoint VARCHAR(100),
						checkpoint_metadata JSONB,
						checkpoint_expires_at TIMESTAMP,
						status VARCHAR(100) NOT NULL,
						created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
						deleted_at TIMESTAMP NULL
					);
				`).Error; err != nil {
					return err
				}

				// Create indexes
				if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);`).Error; err != nil {
					return err
				}
				if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);`).Error; err != nil {
					return err
				}
				if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);`).Error; err != nil {
					return err
				}
				if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_accounts_provider ON accounts(provider);`).Error; err != nil {
					return err
				}
				if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_accounts_status_history_account_id ON account_status_histories(account_id);`).Error; err != nil {
					return err
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Exec(`DROP TABLE IF EXISTS accounts, users CASCADE;`).Error
			},
		},
	})

	if err := m.Migrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// GetRepositories returns initialized repositories
func GetRepositories(db *gorm.DB) repository.Repositories {
	userRepo := postgresRepo.NewUserRepository(db)
	accountRepo := postgresRepo.NewAccountRepository(db)
	txRepo := NewTxRepository(db, &repository.Repositories{
		User:    userRepo,
		Account: accountRepo,
	})
	return repository.Repositories{
		Tx:      txRepo,
		User:    userRepo,
		Account: accountRepo,
	}
}

type txRepository struct {
	db    *gorm.DB
	repos *repository.Repositories
}

// NewTxRepository creates a new transaction repository
func NewTxRepository(db *gorm.DB, repos *repository.Repositories) repository.TxRepository {
	return &txRepository{db: db, repos: repos}
}

func (r *txRepository) Do(ctx context.Context, fn func(repos *repository.Repositories) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(r.repos)
	})
}
