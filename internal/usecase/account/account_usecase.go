package account

import (
	"context"
	"unipile-connector/internal/domain/entity"
	"unipile-connector/internal/domain/repository"
)

// AccountUsecase handles account business logic
type AccountUsecase struct {
	accountRepo repository.AccountRepository
}

// NewAccountUsecase creates a new account usecase
func NewAccountUsecase(accountRepo repository.AccountRepository) *AccountUsecase {
	return &AccountUsecase{
		accountRepo: accountRepo,
	}
}

// ConnectLinkedInAccount connects a LinkedIn account for a user
func (a *AccountUsecase) ConnectLinkedInAccount(ctx context.Context, userID uint, accountID string) (*entity.Account, error) {
	// Check if user already has a LinkedIn account
	existingAccount, err := a.accountRepo.GetByUserIDAndProvider(ctx, userID, "linkedin")
	if err == nil && existingAccount != nil {
		// Update existing account
		existingAccount.AccountID = accountID
		if err := a.accountRepo.Update(ctx, existingAccount); err != nil {
			return nil, err
		}
		return existingAccount, nil
	}

	// Create new account
	account := &entity.Account{
		UserID:    userID,
		Provider:  "linkedin",
		AccountID: accountID,
	}

	if err := a.accountRepo.Create(ctx, account); err != nil {
		return nil, err
	}

	return account, nil
}

// GetUserAccounts retrieves all accounts for a user
func (a *AccountUsecase) GetUserAccounts(ctx context.Context, userID uint) ([]*entity.Account, error) {
	return a.accountRepo.GetByUserID(ctx, userID)
}

// GetLinkedInAccount retrieves LinkedIn account for a user
func (a *AccountUsecase) GetLinkedInAccount(ctx context.Context, userID uint) (*entity.Account, error) {
	return a.accountRepo.GetByUserIDAndProvider(ctx, userID, "linkedin")
}

// DisconnectLinkedInAccount disconnects LinkedIn account for a user
func (a *AccountUsecase) DisconnectLinkedInAccount(ctx context.Context, userID uint) error {
	return a.accountRepo.DeleteByUserIDAndProvider(ctx, userID, "linkedin")
}
