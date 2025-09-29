package postgres

import (
	"context"
	"unipile-connector/internal/domain/repository"

	"gorm.io/gorm"
)

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
