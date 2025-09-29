package database

import (
	"fmt"
	"log"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
				if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_accounts_user_id ON accounts(user_id);`).Error; err != nil {
					return err
				}
				if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_accounts_provider ON accounts(provider);`).Error; err != nil {
					return err
				}
				if err := tx.Exec(`CREATE INDEX IF NOT EXISTS idx_accounts_status_history_account_id ON account_status_histories(account_id);`).Error; err != nil {
					return err
				}

				// Create trigger function for soft deleting related records
				if err := tx.Exec(`
					CREATE OR REPLACE FUNCTION soft_delete_account_status_histories()
					RETURNS TRIGGER AS $$
					BEGIN
						-- Soft delete related account_status_histories when account is soft deleted
						UPDATE account_status_histories
						SET deleted_at = NOW()
						WHERE account_id = OLD.id
						AND deleted_at IS NULL;
						RETURN OLD;
					END;
					$$ LANGUAGE plpgsql;
				`).Error; err != nil {
					return err
				}

				// Create trigger on accounts table
				if err := tx.Exec(`
					CREATE TRIGGER trigger_soft_delete_account_status_histories
					AFTER UPDATE OF deleted_at ON accounts
					FOR EACH ROW
					WHEN (OLD.deleted_at IS NULL AND NEW.deleted_at IS NOT NULL)
					EXECUTE FUNCTION soft_delete_account_status_histories();
				`).Error; err != nil {
					return err
				}

				return nil
			},
			Rollback: func(tx *gorm.DB) error {
				// Drop trigger and function first
				if err := tx.Exec(`DROP TRIGGER IF EXISTS trigger_soft_delete_account_status_histories ON accounts;`).Error; err != nil {
					return err
				}
				if err := tx.Exec(`DROP FUNCTION IF EXISTS soft_delete_account_status_histories();`).Error; err != nil {
					return err
				}
				// Drop tables
				return tx.Exec(`DROP TABLE IF EXISTS account_status_histories, accounts, users CASCADE;`).Error
			},
		},
	})

	if err := m.Migrate(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations completed successfully")
	return nil
}
