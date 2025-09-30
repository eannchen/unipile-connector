package account

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/sirupsen/logrus"

	"unipile-connector/internal/domain/entity"
	"unipile-connector/internal/domain/errs"
	"unipile-connector/internal/domain/repository"
	"unipile-connector/internal/domain/service"
)

// Usecase handles account business logic
type Usecase interface {
	ListUserAccounts(ctx context.Context, userID uint) ([]*entity.Account, error)
	DisconnectLinkedIn(ctx context.Context, userID uint, accountID string) error
	ConnectLinkedInAccount(ctx context.Context, userID uint, req *ConnectLinkedInRequest) (*entity.Account, error)
	SolveCheckpoint(ctx context.Context, userID uint, req *SolveCheckpointRequest) (*entity.Account, error)
	WaitForAccountValidation(ctx context.Context, userID uint, accountID string, timeout time.Duration) (*entity.Account, error)
}

// UsecaseImpl handles account business logic
type UsecaseImpl struct {
	txRepo        repository.TxRepository
	accountRepo   repository.AccountRepository
	unipileClient service.UnipileClient
	logger        *logrus.Logger
}

// NewAccountUsecase creates a new account usecase
func NewAccountUsecase(txRepo repository.TxRepository, accountRepo repository.AccountRepository, unipileClient service.UnipileClient, logger *logrus.Logger) Usecase {
	return &UsecaseImpl{
		txRepo:        txRepo,
		accountRepo:   accountRepo,
		unipileClient: unipileClient,
		logger:        logger,
	}
}

// ListUserAccounts retrieves all accounts for a user
func (a *UsecaseImpl) ListUserAccounts(ctx context.Context, userID uint) ([]*entity.Account, error) {
	accounts, err := a.accountRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, errs.WrapInternalError(err, "Failed to get accounts by user ID")
	}
	return accounts, nil
}

// DisconnectLinkedIn disconnects LinkedIn account for a user
func (a *UsecaseImpl) DisconnectLinkedIn(ctx context.Context, userID uint, accountID string) error {
	return a.txRepo.Do(ctx, func(repos *repository.Repositories) error {
		if err := repos.Account.DeleteByUserIDAndAccountID(ctx, userID, accountID); err != nil {
			return errs.WrapInternalError(err, "Failed to delete account")
		}
		if err := a.unipileClient.DeleteAccount(accountID); err != nil && err != service.ErrUnipileAccountNotFound {
			return errs.WrapInternalError(err, "Failed to delete account on Unipile")
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
func (a *UsecaseImpl) ConnectLinkedInAccount(ctx context.Context, userID uint, req *ConnectLinkedInRequest) (*entity.Account, error) {

	resp, err := a.unipileClient.ConnectLinkedIn(&service.ConnectLinkedInRequest{
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
			return nil, errs.WrapInternalError(err, "Failed to create account")
		}
		return account, nil
	}

	checkpointBody, err := json.Marshal(resp.Checkpoint)
	if err != nil {
		return nil, errs.WrapInternalError(err, "Failed to marshal checkpoint")
	}

	account.AccountStatusHistories = append(account.AccountStatusHistories, entity.AccountStatusHistory{
		Checkpoint:          resp.Checkpoint.Type,
		CheckpointMetadata:  checkpointBody,
		CheckpointExpiresAt: time.Now().Add(270 * time.Second), // 4.5 minutes
		Status:              "PENDING",
	})

	if err := a.accountRepo.Create(ctx, account); err != nil {
		return nil, errs.WrapInternalError(err, "Failed to create account")
	}

	return account, nil
}

// SolveCheckpointRequest represents request to solve a checkpoint
type SolveCheckpointRequest struct {
	AccountID string
	Code      string
}

// SolveCheckpoint solves a LinkedIn authentication checkpoint
func (a *UsecaseImpl) SolveCheckpoint(ctx context.Context, userID uint, req *SolveCheckpointRequest) (*entity.Account, error) {

	var account *entity.Account

	if err := a.txRepo.Do(ctx, func(repos *repository.Repositories) error {
		var err error

		account, err = repos.Account.GetByUserIDAndAccountIDForUpdate(ctx, userID, req.AccountID)
		if err != nil {
			if errors.Is(err, repository.ErrAccountNotFound) {
				return errs.WrapValidationError(errors.New("account not found"), "Account not found")
			}
			return errs.WrapInternalError(err, "Failed to get account")
		}
		if account.CurrentStatus == "OK" {
			return nil
		}

		account.CurrentStatus = "OK"
		if err := repos.Account.Update(ctx, account); err != nil {
			return errs.WrapInternalError(err, "Failed to update account")
		}

		if _, err := a.unipileClient.SolveCheckpoint(&service.SolveCheckpointRequest{
			Provider:  account.Provider,
			AccountID: req.AccountID,
			Code:      req.Code,
		}); err != nil {
			if errors.Is(err, service.ErrUnipileInvalidCodeOrExpiredCheckpoint) {
				return errs.ErrInvalidCodeOrExpiredCheckpoint
			}
			return errs.WrapInternalError(err, "Failed to solve checkpoint")
		}

		return nil
	}); err != nil {
		return nil, err
	}

	return account, nil
}

// WaitForAccountValidation waits for IN_APP_VALIDATION checkpoint to be resolved using long polling
func (a *UsecaseImpl) WaitForAccountValidation(ctx context.Context, userID uint, accountID string, timeout time.Duration) (*entity.Account, error) {
	accountWithStatus, err := a.accountRepo.GetWithStatus(ctx, userID, accountID, "IN_APP_VALIDATION")
	if err != nil {
		if errors.Is(err, repository.ErrAccountNotFound) {
			return nil, errs.WrapValidationError(errors.New("account not found, expired, or not in IN_APP_VALIDATION state"), "Account not found, expired, or not in IN_APP_VALIDATION state")
		}
		return nil, errs.WrapInternalError(err, "Failed to get account")
	}
	if accountWithStatus.CurrentStatus == "OK" {
		return &accountWithStatus.Account, nil
	}

	// Use timeout provided by frontend (similar to Unipile API timeout)
	pollTimeout := timeout
	if pollTimeout <= 0 {
		pollTimeout = 5 * time.Minute // Default fallback
	}

	logFields := logrus.Fields{
		"userID":    userID,
		"accountID": accountID,
		"timeout":   pollTimeout,
	}
	a.logger.WithFields(logFields).Info("Starting long polling for IN_APP_VALIDATION")

	// Use long polling to wait for account status change
	unipileAccount, err := a.unipileClient.GetAccountWithLongPolling(accountID, pollTimeout)
	if err != nil {
		a.logger.WithError(err).WithFields(logFields).Error("Long polling failed")
		if errors.Is(err, service.ErrUnipileAccountNotFound) {
			return nil, errs.WrapValidationError(errors.New("account not found"), "Account not found")
		}
		return nil, errs.WrapInternalError(err, "Failed to check account status")
	}

	// Check if any source has status "OK"
	accountStatusOK := false
	for _, source := range unipileAccount.Sources {
		if source.Status == "OK" {
			accountStatusOK = true
			break
		}
	}
	if !accountStatusOK {
		return nil, errs.WrapValidationError(errors.New("account validation failed"), "Account validation failed")
	}

	account := accountWithStatus.Account
	account.CurrentStatus = "OK"
	if err := a.accountRepo.Update(ctx, &account); err != nil {
		return nil, errs.WrapInternalError(err, "Failed to update account status")
	}

	return &account, nil
}
