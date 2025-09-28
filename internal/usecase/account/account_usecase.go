package account

import (
	"context"
	"errors"
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
func NewAccountUsecase(accountRepo repository.AccountRepository, txRepo repository.TxRepository, unipileClient *client.UnipileClient, logger *logrus.Logger) *AccountUsecase {
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
func (a *AccountUsecase) DisconnectLinkedIn(ctx context.Context, userID uint) error {
	return a.txRepo.Do(ctx, func(repos *repository.Repositories) error {
		account, err := repos.Account.GetByUserIDAndProviderForUpdate(ctx, userID, "LINKEDIN")
		if err != nil {
			if errors.Is(err, repository.ErrAccountNotFound) {
				return nil
			}
			return err
		}
		if err := a.unipileClient.DeleteAccount(account.AccountID); err != nil && err != client.ErrAccountNotFound {
			return err
		}
		return repos.Account.DeleteByUserIDAndProvider(ctx, userID, "LINKEDIN")
	})
}

// ConnectLinkedInRequest represents request to connect LinkedIn account via Unipile
type ConnectLinkedInRequest struct {
	Username    string
	Password    string
	AccessToken string
	UserAgent   string
}

// ConnectLinkedInResponse represents response from Unipile connection
type ConnectLinkedInResponse struct {
	Success     bool
	Account     *entity.Account
	Checkpoint  *client.Checkpoint
	ExpiresAt   time.Time
	RowResponse string
}

// ConnectLinkedInAccount connects a LinkedIn account for a user
func (a *AccountUsecase) ConnectLinkedInAccount(ctx context.Context, userID uint, req *ConnectLinkedInRequest) (*ConnectLinkedInResponse, error) {

	var connResp *ConnectLinkedInResponse

	err := a.txRepo.Do(ctx, func(repos *repository.Repositories) error {
		account, err := repos.Account.GetByUserIDAndProviderForUpdate(ctx, userID, "LINKEDIN")
		if err != nil && !errors.Is(err, repository.ErrAccountNotFound) {
			return err
		}
		if account != nil { // Account already exists and connected
			return nil
		}

		resp, err := a.unipileClient.ConnectLinkedIn(&client.ConnectLinkedInRequest{
			Provider:    "LINKEDIN",
			Username:    req.Username,
			Password:    req.Password,
			AccessToken: req.AccessToken,
			UserAgent:   req.UserAgent,
		})
		if err != nil {
			return err
		}

		account = &entity.Account{
			UserID:    userID,
			Provider:  "LINKEDIN",
			AccountID: resp.AccountID,
		}

		if resp.Checkpoint == nil {
			if err := a.accountRepo.Create(ctx, account); err != nil {
				a.logger.Errorf("Failed to create account: %v with accountID created on Unipile: %s", err, account.AccountID)
				return err
			}
			connResp = &ConnectLinkedInResponse{
				Success: true,
				Account: account,
			}
			return nil
		}

		connResp = &ConnectLinkedInResponse{
			Success:     false,
			Account:     account,
			Checkpoint:  resp.Checkpoint,
			ExpiresAt:   time.Now().Add(4 * time.Minute),
			RowResponse: resp.RowBody,
		}

		return nil
	})

	return connResp, err
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
	var err error

	err = a.txRepo.Do(ctx, func(repos *repository.Repositories) error {
		account, err = repos.Account.GetByUserIDAndProviderForUpdate(ctx, userID, "LINKEDIN")
		if err != nil && !errors.Is(err, repository.ErrAccountNotFound) {
			return err
		}
		if account != nil { // Account already exists and connected
			return nil
		}

		resp, err := a.unipileClient.SolveCheckpoint(&client.SolveCheckpointRequest{
			Provider:  "LINKEDIN",
			AccountID: req.AccountID,
			Code:      req.Code,
		})
		if err != nil {
			if errors.Is(err, client.ErrInvalidCodeOrExpiredCheckpoint) {
				return ErrInvalidCodeOrExpiredCheckpoint
			}
			return err
		}

		account = &entity.Account{
			UserID:    userID,
			Provider:  "LINKEDIN",
			AccountID: resp.AccountID,
		}
		if err := a.accountRepo.Create(ctx, account); err != nil {
			a.logger.Errorf("Failed to create account: %v with accountID created on Unipile: %s", err, account.AccountID)
			return err
		}
		return nil
	})

	return account, err
}
