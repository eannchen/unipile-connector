package postgres

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"unipile-connector/internal/domain/entity"
	"unipile-connector/internal/domain/repository"
)

// userRepo implements UserRepository interface
type userRepo struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *entity.User) error {
	err := r.db.WithContext(ctx).Create(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return repository.ErrDuplicateKey
		}
		return err
	}
	return nil
}

func (r *userRepo) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user entity.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
