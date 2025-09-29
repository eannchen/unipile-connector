package repository

import (
	"context"
	"errors"

	"unipile-connector/internal/domain/entity"
)

// AccountRepository defines the interface for account data operations
type AccountRepository interface {
	Create(ctx context.Context, account *entity.Account) error
	GetByUserID(ctx context.Context, userID uint) ([]*entity.Account, error)
	GetByUserIDAndAccountIDForUpdate(ctx context.Context, userID uint, accountID string) (*entity.Account, error)
	Update(ctx context.Context, account *entity.Account) error
	DeleteByUserIDAndAccountID(ctx context.Context, userID uint, accountID string) error
}

// ErrAccountNotFound is returned when an account is not found
var ErrAccountNotFound = errors.New("account not found")
