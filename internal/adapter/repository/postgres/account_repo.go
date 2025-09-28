package postgres

import (
	"context"
	"unipile-connector/internal/domain/entity"
	"unipile-connector/internal/domain/repository"

	"gorm.io/gorm"
)

// accountRepo implements AccountRepository interface
type accountRepo struct {
	db *gorm.DB
}

// NewAccountRepository creates a new account repository
func NewAccountRepository(db *gorm.DB) repository.AccountRepository {
	return &accountRepo{db: db}
}

func (r *accountRepo) Create(ctx context.Context, account *entity.Account) error {
	return r.db.WithContext(ctx).Create(account).Error
}

func (r *accountRepo) GetByID(ctx context.Context, id uint) (*entity.Account, error) {
	var account entity.Account
	err := r.db.WithContext(ctx).Preload("User").First(&account, id).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *accountRepo) GetByUserID(ctx context.Context, userID uint) ([]*entity.Account, error) {
	var accounts []*entity.Account
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&accounts).Error
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *accountRepo) GetByUserIDAndProvider(ctx context.Context, userID uint, provider string) (*entity.Account, error) {
	var account entity.Account
	err := r.db.WithContext(ctx).Where("user_id = ? AND provider = ?", userID, provider).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *accountRepo) Update(ctx context.Context, account *entity.Account) error {
	return r.db.WithContext(ctx).Save(account).Error
}

func (r *accountRepo) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.Account{}, id).Error
}

func (r *accountRepo) DeleteByUserIDAndProvider(ctx context.Context, userID uint, provider string) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND provider = ?", userID, provider).Delete(&entity.Account{}).Error
}

