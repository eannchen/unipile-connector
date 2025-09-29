package account

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"unipile-connector/internal/domain/entity"
	"unipile-connector/internal/domain/repository"
	"unipile-connector/internal/infrastructure/client"
)

// AccountUsecase handles account business logic
type AccountUsecase struct {
	txRepo        repository.TxRepository
	accountRepo   repository.AccountRepository
	unipileClient *client.UnipileClient
	logger        *logrus.Logger
}

// NewAccountUsecase creates a new account usecase
func NewAccountUsecase(txRepo repository.TxRepository, accountRepo repository.AccountRepository, unipileClient *client.UnipileClient, logger *logrus.Logger) *AccountUsecase {
	return &AccountUsecase{
		txRepo:        txRepo,
		accountRepo:   accountRepo,
		unipileClient: unipileClient,
		logger:        logger,
	}
}

// ListUserAccounts retrieves all accounts for a user
func (a *AccountUsecase) ListUserAccounts(ctx context.Context, userID uint) ([]*entity.Account, error) {
	return a.accountRepo.GetByUserID(ctx, userID)
}

// DisconnectLinkedIn disconnects LinkedIn account for a user
func (a *AccountUsecase) DisconnectLinkedIn(ctx context.Context, userID uint, accountID string) error {
	return a.txRepo.Do(ctx, func(repos *repository.Repositories) error {
		if err := repos.Account.DeleteByUserIDAndAccountID(ctx, userID, accountID); err != nil {
			return fmt.Errorf("failed to delete account: %w", err)
		}
		if err := a.unipileClient.DeleteAccount(accountID); err != nil && err != client.ErrAccountNotFound {
			return fmt.Errorf("failed to delete account on Unipile: %w", err)
		}
		return nil
	})
}

// ConnectLinkedInRequest represents request to connect LinkedIn account via Unipile
type ConnectLinkedInRequest struct {
	Username    string
	Password    string
	AccessToken string
	UserAgent   string
}

// ConnectLinkedInAccount connects a LinkedIn account for a user
func (a *AccountUsecase) ConnectLinkedInAccount(ctx context.Context, userID uint, req *ConnectLinkedInRequest) (*entity.Account, error) {

	resp, err := a.unipileClient.ConnectLinkedIn(&client.ConnectLinkedInRequest{
		Provider:    "LINKEDIN",
		Username:    req.Username,
		Password:    req.Password,
		AccessToken: req.AccessToken,
		UserAgent:   req.UserAgent,
	})
	if err != nil {
		return nil, err
	}

	account := &entity.Account{
		UserID:        userID,
		Provider:      "LINKEDIN",
		AccountID:     resp.AccountID,
		CurrentStatus: "PENDING",
	}

	if resp.Checkpoint == nil {
		account.CurrentStatus = "OK"
		if err := a.accountRepo.Create(ctx, account); err != nil {
			return nil, fmt.Errorf("failed to create account: %w", err)
		}
		return account, nil
	}

	checkpointBody, err := json.Marshal(resp.Checkpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal checkpoint: %w", err)
	}

	account.AccountStatusHistories = append(account.AccountStatusHistories, entity.AccountStatusHistory{
		Checkpoint:          resp.Checkpoint.Type,
		CheckpointMetadata:  checkpointBody,
		CheckpointExpiresAt: time.Now().Add(270 * time.Second), // 4.5 minutes
		Status:              "PENDING",
	})

	if err := a.accountRepo.Create(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to create account: %w", err)
	}

	return account, err
}

// SolveCheckpointRequest represents request to solve a checkpoint
type SolveCheckpointRequest struct {
	AccountID string
	Code      string
}

// ErrInvalidCodeOrExpiredCheckpoint is returned when the code is invalid or the checkpoint expired
var ErrInvalidCodeOrExpiredCheckpoint = errors.New("invalid code or expired checkpoint")

// SolveCheckpoint solves a LinkedIn authentication checkpoint
func (a *AccountUsecase) SolveCheckpoint(ctx context.Context, userID uint, req *SolveCheckpointRequest) (*entity.Account, error) {

	var account *entity.Account

	if err := a.txRepo.Do(ctx, func(repos *repository.Repositories) error {
		var err error

		account, err = repos.Account.GetByUserIDAndAccountIDForUpdate(ctx, userID, req.AccountID)
		if err != nil {
			return fmt.Errorf("failed to get account: %w", err)
		}
		if account.CurrentStatus == "OK" {
			return nil
		}

		account.CurrentStatus = "OK"
		if err := repos.Account.Update(ctx, account); err != nil {
			return fmt.Errorf("failed to update account: %w", err)
		}

		if _, err := a.unipileClient.SolveCheckpoint(&client.SolveCheckpointRequest{
			Provider:  account.Provider,
			AccountID: req.AccountID,
			Code:      req.Code,
		}); err != nil {
			if errors.Is(err, client.ErrInvalidCodeOrExpiredCheckpoint) {
				return ErrInvalidCodeOrExpiredCheckpoint
			}
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return account, nil
}
