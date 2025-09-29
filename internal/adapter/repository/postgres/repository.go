package postgres

import (
	"gorm.io/gorm"

	"unipile-connector/internal/domain/repository"
)

// GetRepositories returns initialized repositories
func GetRepositories(db *gorm.DB) repository.Repositories {
	userRepo := NewUserRepository(db)
	accountRepo := NewAccountRepository(db)
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
