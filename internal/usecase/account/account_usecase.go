package account

import (
	"context"

	"github.com/sirupsen/logrus"

	"unipile-connector/internal/domain/entity"
	"unipile-connector/internal/domain/repository"
	"unipile-connector/internal/infrastructure/client"
)

// AccountUsecase handles account business logic
type AccountUsecase struct {
	accountRepo   repository.AccountRepository
	unipileClient *client.UnipileClient
	logger        *logrus.Logger
}

// NewAccountUsecase creates a new account usecase
func NewAccountUsecase(accountRepo repository.AccountRepository, unipileClient *client.UnipileClient, logger *logrus.Logger) *AccountUsecase {
	return &AccountUsecase{
		accountRepo:   accountRepo,
		unipileClient: unipileClient,
		logger:        logger,
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

// ListUserAccounts retrieves all accounts for a user
func (a *AccountUsecase) ListUserAccounts(ctx context.Context, userID uint) ([]*entity.Account, error) {
	return a.accountRepo.GetByUserID(ctx, userID)
}

// DisconnectLinkedIn disconnects LinkedIn account for a user
func (a *AccountUsecase) DisconnectLinkedIn(ctx context.Context, userID uint) error {
	return a.accountRepo.DeleteByUserIDAndProvider(ctx, userID, "LINKEDIN")
}
