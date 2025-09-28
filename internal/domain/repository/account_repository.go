package repository

import (
	"context"
	"unipile-connector/internal/domain/entity"
)

// AccountRepository defines the interface for account data operations
type AccountRepository interface {
	Create(ctx context.Context, account *entity.Account) error
	GetByID(ctx context.Context, id uint) (*entity.Account, error)
	GetByUserID(ctx context.Context, userID uint) ([]*entity.Account, error)
	GetByUserIDAndProvider(ctx context.Context, userID uint, provider string) (*entity.Account, error)
	Update(ctx context.Context, account *entity.Account) error
	Delete(ctx context.Context, id uint) error
	DeleteByUserIDAndProvider(ctx context.Context, userID uint, provider string) error
}

