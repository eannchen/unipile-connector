package postgres

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"unipile-connector/internal/domain/entity"
	"unipile-connector/internal/domain/repository"
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

func (r *accountRepo) GetByUserID(ctx context.Context, userID uint) ([]*entity.Account, error) {
	var accounts []*entity.Account
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&accounts).Error
	if err != nil {
		return nil, err
	}
	return accounts, nil
}

func (r *accountRepo) GetByUserIDAndAccountIDForUpdate(ctx context.Context, userID uint, accountID string) (*entity.Account, error) {
	var account entity.Account
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ? AND account_id = ?", userID, accountID).First(&account).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrAccountNotFound
		}
		return nil, err
	}
	return &account, nil
}

func (r *accountRepo) GetWithStatus(ctx context.Context, userID uint, accountID, checkpoint string) (*entity.AccountWithStatus, error) {
	var accountWithStatus entity.AccountWithStatus
	err := r.db.WithContext(ctx).
		Debug().
		Select("accounts.*, ash.status as current_status, ash.checkpoint, ash.checkpoint_expires_at").
		Table("accounts").
		Joins(`INNER JOIN (
			SELECT account_id, status, checkpoint, checkpoint_expires_at
			FROM account_status_histories ash
			WHERE checkpoint_expires_at > ? AND checkpoint = ? AND deleted_at IS NULL
			ORDER BY ash.created_at DESC
			LIMIT 1
		) ash ON accounts.id = ash.account_id`, time.Now(), checkpoint).
		Where("accounts.user_id = ? AND accounts.account_id = ? AND accounts.deleted_at IS NULL", userID, accountID).
		First(&accountWithStatus).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, repository.ErrAccountNotFound
		}
		return nil, err
	}
	return &accountWithStatus, nil
}

func (r *accountRepo) Update(ctx context.Context, account *entity.Account) error {
	return r.db.WithContext(ctx).Save(account).Error
}

func (r *accountRepo) DeleteByUserIDAndAccountID(ctx context.Context, userID uint, accountID string) error {
	return r.db.WithContext(ctx).Where("user_id = ? AND account_id = ?", userID, accountID).Delete(&entity.Account{}).Error
}
