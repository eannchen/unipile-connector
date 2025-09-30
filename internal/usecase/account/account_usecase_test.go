package account

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"

	"unipile-connector/internal/domain/entity"
	"unipile-connector/internal/domain/errs"
	"unipile-connector/internal/domain/repository"
	"unipile-connector/internal/domain/service"
)

type mockAccountRepo struct {
	createFunc                       func(ctx context.Context, account *entity.Account) error
	getByUserIDFunc                  func(ctx context.Context, userID uint) ([]*entity.Account, error)
	getByUserIDAndAccountIDForUpdate func(ctx context.Context, userID uint, accountID string) (*entity.Account, error)
	getWithStatusFunc                func(ctx context.Context, userID uint, accountID, checkpoint string) (*entity.AccountWithStatus, error)
	updateFunc                       func(ctx context.Context, account *entity.Account) error
	deleteByUserIDAndAccountIDFunc   func(ctx context.Context, userID uint, accountID string) error
}

func (m *mockAccountRepo) Create(ctx context.Context, account *entity.Account) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, account)
	}
	return nil
}

func (m *mockAccountRepo) GetByUserID(ctx context.Context, userID uint) ([]*entity.Account, error) {
	if m.getByUserIDFunc != nil {
		return m.getByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockAccountRepo) GetByUserIDAndAccountIDForUpdate(ctx context.Context, userID uint, accountID string) (*entity.Account, error) {
	if m.getByUserIDAndAccountIDForUpdate != nil {
		return m.getByUserIDAndAccountIDForUpdate(ctx, userID, accountID)
	}
	return nil, nil
}

func (m *mockAccountRepo) GetWithStatus(ctx context.Context, userID uint, accountID, checkpoint string) (*entity.AccountWithStatus, error) {
	if m.getWithStatusFunc != nil {
		return m.getWithStatusFunc(ctx, userID, accountID, checkpoint)
	}
	return nil, nil
}

func (m *mockAccountRepo) Update(ctx context.Context, account *entity.Account) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, account)
	}
	return nil
}

func (m *mockAccountRepo) DeleteByUserIDAndAccountID(ctx context.Context, userID uint, accountID string) error {
	if m.deleteByUserIDAndAccountIDFunc != nil {
		return m.deleteByUserIDAndAccountIDFunc(ctx, userID, accountID)
	}
	return nil
}

type mockTxRepo struct {
	doFunc func(ctx context.Context, fn func(*repository.Repositories) error) error
}

func (m *mockTxRepo) Do(ctx context.Context, fn func(*repository.Repositories) error) error {
	if m.doFunc != nil {
		return m.doFunc(ctx, fn)
	}
	return fn(&repository.Repositories{})
}

type mockUnipileClient struct {
	listAccountsFunc              func() (*service.AccountListResponse, error)
	testConnectionFunc            func() error
	getAccountFunc                func(accountID string) (*service.Account, error)
	getAccountWithLongPollingFunc func(accountID string, timeout time.Duration) (*service.Account, error)
	deleteAccountFunc             func(accountID string) error
	connectLinkedInFunc           func(req *service.ConnectLinkedInRequest) (*service.ConnectLinkedInResponse, error)
	solveCheckpointFunc           func(req *service.SolveCheckpointRequest) (*service.SolveCheckpointResponse, error)
}

func (m *mockUnipileClient) ListAccounts() (*service.AccountListResponse, error) {
	if m.listAccountsFunc != nil {
		return m.listAccountsFunc()
	}
	return nil, nil
}

func (m *mockUnipileClient) TestConnection() error {
	if m.testConnectionFunc != nil {
		return m.testConnectionFunc()
	}
	return nil
}

func (m *mockUnipileClient) GetAccount(accountID string) (*service.Account, error) {
	if m.getAccountFunc != nil {
		return m.getAccountFunc(accountID)
	}
	return nil, nil
}

func (m *mockUnipileClient) GetAccountWithLongPolling(accountID string, timeout time.Duration) (*service.Account, error) {
	if m.getAccountWithLongPollingFunc != nil {
		return m.getAccountWithLongPollingFunc(accountID, timeout)
	}
	return nil, nil
}

func (m *mockUnipileClient) DeleteAccount(accountID string) error {
	if m.deleteAccountFunc != nil {
		return m.deleteAccountFunc(accountID)
	}
	return nil
}

func (m *mockUnipileClient) ConnectLinkedIn(req *service.ConnectLinkedInRequest) (*service.ConnectLinkedInResponse, error) {
	if m.connectLinkedInFunc != nil {
		return m.connectLinkedInFunc(req)
	}
	return nil, nil
}

func (m *mockUnipileClient) SolveCheckpoint(req *service.SolveCheckpointRequest) (*service.SolveCheckpointResponse, error) {
	if m.solveCheckpointFunc != nil {
		return m.solveCheckpointFunc(req)
	}
	return nil, nil
}

func TestConnectLinkedInAccount_SuccessWithoutCheckpoint(t *testing.T) {
	ctx := context.Background()
	var createdAccount *entity.Account

	accountRepo := &mockAccountRepo{
		createFunc: func(_ context.Context, account *entity.Account) error {
			createdAccount = account
			return nil
		},
	}

	unipileClient := &mockUnipileClient{
		connectLinkedInFunc: func(req *service.ConnectLinkedInRequest) (*service.ConnectLinkedInResponse, error) {
			if req.Provider != "LINKEDIN" {
				t.Fatalf("expected provider LINKEDIN, got %s", req.Provider)
			}
			return &service.ConnectLinkedInResponse{AccountID: "acc-123"}, nil
		},
	}

	uc := NewAccountUsecase(&mockTxRepo{}, accountRepo, unipileClient, logrus.New())

	account, err := uc.ConnectLinkedInAccount(ctx, 42, &ConnectLinkedInRequest{Username: "user", Password: "pass"})
	if err != nil {
		t.Fatalf("ConnectLinkedInAccount returned error: %v", err)
	}

	if createdAccount == nil {
		t.Fatalf("expected account to be persisted")
	}

	if account.AccountID != "acc-123" {
		t.Fatalf("expected account ID acc-123, got %s", account.AccountID)
	}

	if account.CurrentStatus != "OK" {
		t.Fatalf("expected current status OK, got %s", account.CurrentStatus)
	}

	if account.UserID != 42 {
		t.Fatalf("expected user ID 42, got %d", account.UserID)
	}

	if len(account.AccountStatusHistories) != 0 {
		t.Fatalf("expected no status histories, got %d", len(account.AccountStatusHistories))
	}
}

func TestConnectLinkedInAccount_SuccessWithCheckpoint(t *testing.T) {
	ctx := context.Background()
	var createdAccount *entity.Account
	start := time.Now()

	accountRepo := &mockAccountRepo{
		createFunc: func(_ context.Context, account *entity.Account) error {
			createdAccount = account
			return nil
		},
	}

	unipileClient := &mockUnipileClient{
		connectLinkedInFunc: func(req *service.ConnectLinkedInRequest) (*service.ConnectLinkedInResponse, error) {
			checkpoint := &service.Checkpoint{Type: "OTP", Source: "APP"}
			return &service.ConnectLinkedInResponse{AccountID: "acc-otp", Checkpoint: checkpoint}, nil
		},
	}

	uc := NewAccountUsecase(&mockTxRepo{}, accountRepo, unipileClient, logrus.New())

	account, err := uc.ConnectLinkedInAccount(ctx, 7, &ConnectLinkedInRequest{AccessToken: "token", UserAgent: "agent"})
	if err != nil {
		t.Fatalf("ConnectLinkedInAccount returned error: %v", err)
	}

	if account.CurrentStatus != "PENDING" {
		t.Fatalf("expected current status PENDING, got %s", account.CurrentStatus)
	}

	if len(account.AccountStatusHistories) != 1 {
		t.Fatalf("expected one status history entry, got %d", len(account.AccountStatusHistories))
	}

	history := account.AccountStatusHistories[0]
	if history.Checkpoint != "OTP" {
		t.Fatalf("expected checkpoint type OTP, got %s", history.Checkpoint)
	}

	var checkpoint service.Checkpoint
	if err := json.Unmarshal(history.CheckpointMetadata, &checkpoint); err != nil {
		t.Fatalf("failed to unmarshal checkpoint metadata: %v", err)
	}
	if checkpoint.Type != "OTP" || checkpoint.Source != "APP" {
		t.Fatalf("unexpected checkpoint metadata: %+v", checkpoint)
	}

	if history.Status != "PENDING" {
		t.Fatalf("expected history status PENDING, got %s", history.Status)
	}

	lowerBound := start.Add(260 * time.Second)
	upperBound := start.Add(280 * time.Second)
	if history.CheckpointExpiresAt.Before(lowerBound) || history.CheckpointExpiresAt.After(upperBound) {
		t.Fatalf("expected expiry within 260-280s window, got %v", history.CheckpointExpiresAt.Sub(start))
	}

	if createdAccount == nil {
		t.Fatalf("expected account to be persisted")
	}
}

func TestConnectLinkedInAccount_ClientError(t *testing.T) {
	ctx := context.Background()

	wantErr := errors.New("connect failed")
	unipileClient := &mockUnipileClient{
		connectLinkedInFunc: func(req *service.ConnectLinkedInRequest) (*service.ConnectLinkedInResponse, error) {
			return nil, wantErr
		},
	}

	uc := NewAccountUsecase(&mockTxRepo{}, &mockAccountRepo{}, unipileClient, logrus.New())

	_, err := uc.ConnectLinkedInAccount(ctx, 1, &ConnectLinkedInRequest{})
	if err != wantErr {
		t.Fatalf("expected error %v, got %v", wantErr, err)
	}
}

func TestSolveCheckpoint_Success(t *testing.T) {
	ctx := context.Background()
	accountForUpdate := &entity.Account{UserID: 4, AccountID: "acc-1", Provider: "LINKEDIN", CurrentStatus: "PENDING"}
	var updatedAccount *entity.Account
	var solveCalled bool

	accountRepo := &mockAccountRepo{
		getByUserIDAndAccountIDForUpdate: func(_ context.Context, userID uint, accountID string) (*entity.Account, error) {
			if userID != 4 || accountID != "acc-1" {
				t.Fatalf("unexpected lookup params userID=%d accountID=%s", userID, accountID)
			}
			return accountForUpdate, nil
		},
		updateFunc: func(_ context.Context, account *entity.Account) error {
			updatedAccount = account
			return nil
		},
	}

	txRepo := &mockTxRepo{
		doFunc: func(ctx context.Context, fn func(*repository.Repositories) error) error {
			return fn(&repository.Repositories{Account: accountRepo})
		},
	}

	unipileClient := &mockUnipileClient{
		solveCheckpointFunc: func(req *service.SolveCheckpointRequest) (*service.SolveCheckpointResponse, error) {
			solveCalled = true
			if req.AccountID != "acc-1" || req.Code != "123456" {
				t.Fatalf("unexpected solve request: %+v", req)
			}
			return &service.SolveCheckpointResponse{}, nil
		},
	}

	uc := NewAccountUsecase(txRepo, &mockAccountRepo{}, unipileClient, logrus.New())

	account, err := uc.SolveCheckpoint(ctx, 4, &SolveCheckpointRequest{AccountID: "acc-1", Code: "123456"})
	if err != nil {
		t.Fatalf("SolveCheckpoint returned error: %v", err)
	}

	if !solveCalled {
		t.Fatalf("expected SolveCheckpoint to call Unipile client")
	}

	if updatedAccount == nil {
		t.Fatalf("expected account update to be called")
	}

	if account.CurrentStatus != "OK" {
		t.Fatalf("expected account status OK, got %s", account.CurrentStatus)
	}
}

func TestSolveCheckpoint_AlreadyOK(t *testing.T) {
	ctx := context.Background()
	accountForUpdate := &entity.Account{UserID: 4, AccountID: "acc-2", Provider: "LINKEDIN", CurrentStatus: "OK"}
	var updateCalled bool
	var solveCalled bool

	accountRepo := &mockAccountRepo{
		getByUserIDAndAccountIDForUpdate: func(_ context.Context, userID uint, accountID string) (*entity.Account, error) {
			return accountForUpdate, nil
		},
		updateFunc: func(_ context.Context, account *entity.Account) error {
			updateCalled = true
			return nil
		},
	}

	txRepo := &mockTxRepo{
		doFunc: func(ctx context.Context, fn func(*repository.Repositories) error) error {
			return fn(&repository.Repositories{Account: accountRepo})
		},
	}

	unipileClient := &mockUnipileClient{
		solveCheckpointFunc: func(req *service.SolveCheckpointRequest) (*service.SolveCheckpointResponse, error) {
			solveCalled = true
			return &service.SolveCheckpointResponse{}, nil
		},
	}

	uc := NewAccountUsecase(txRepo, &mockAccountRepo{}, unipileClient, logrus.New())

	account, err := uc.SolveCheckpoint(ctx, 4, &SolveCheckpointRequest{AccountID: "acc-2", Code: "000000"})
	if err != nil {
		t.Fatalf("SolveCheckpoint returned error: %v", err)
	}

	if updateCalled {
		t.Fatalf("expected update not to be called when account already OK")
	}

	if solveCalled {
		t.Fatalf("expected unipile solve not to be called when account already OK")
	}

	if account.CurrentStatus != "OK" {
		t.Fatalf("expected account status OK, got %s", account.CurrentStatus)
	}
}

func TestSolveCheckpoint_InvalidCode(t *testing.T) {
	ctx := context.Background()

	accountRepo := &mockAccountRepo{
		getByUserIDAndAccountIDForUpdate: func(_ context.Context, userID uint, accountID string) (*entity.Account, error) {
			return &entity.Account{UserID: userID, AccountID: accountID, Provider: "LINKEDIN", CurrentStatus: "PENDING"}, nil
		},
		updateFunc: func(_ context.Context, account *entity.Account) error { return nil },
	}

	txRepo := &mockTxRepo{
		doFunc: func(ctx context.Context, fn func(*repository.Repositories) error) error {
			return fn(&repository.Repositories{Account: accountRepo})
		},
	}

	unipileClient := &mockUnipileClient{
		solveCheckpointFunc: func(req *service.SolveCheckpointRequest) (*service.SolveCheckpointResponse, error) {
			return nil, service.ErrUnipileInvalidCodeOrExpiredCheckpoint
		},
	}

	uc := NewAccountUsecase(txRepo, &mockAccountRepo{}, unipileClient, logrus.New())

	_, err := uc.SolveCheckpoint(ctx, 1, &SolveCheckpointRequest{AccountID: "acc-invalid", Code: "bad"})
	if err == nil {
		t.Fatalf("expected error but got nil")
	}

	if !errors.Is(err, errs.ErrInvalidCodeOrExpiredCheckpoint) {
		t.Fatalf("expected invalid code error, got %v", err)
	}
}

func TestSolveCheckpoint_AccountNotFound(t *testing.T) {
	ctx := context.Background()

	accountRepo := &mockAccountRepo{
		getByUserIDAndAccountIDForUpdate: func(_ context.Context, userID uint, accountID string) (*entity.Account, error) {
			return nil, repository.ErrAccountNotFound
		},
	}

	txRepo := &mockTxRepo{
		doFunc: func(ctx context.Context, fn func(*repository.Repositories) error) error {
			return fn(&repository.Repositories{Account: accountRepo})
		},
	}

	uc := NewAccountUsecase(txRepo, &mockAccountRepo{}, &mockUnipileClient{}, logrus.New())

	_, err := uc.SolveCheckpoint(ctx, 10, &SolveCheckpointRequest{AccountID: "missing", Code: "000"})
	if err == nil {
		t.Fatalf("expected error but got nil")
	}

	var codedErr *errs.CodedError
	if !errors.As(err, &codedErr) {
		t.Fatalf("expected coded error, got %v", err)
	}

	if codedErr.Kind != errs.ValidationErrorKind {
		t.Fatalf("expected validation error kind, got %s", codedErr.Kind)
	}
}

func TestDisconnectLinkedIn_Success(t *testing.T) {
	ctx := context.Background()
	var deleteCalled bool
	var deleteAccountCalled bool

	accountRepo := &mockAccountRepo{
		deleteByUserIDAndAccountIDFunc: func(_ context.Context, userID uint, accountID string) error {
			deleteCalled = true
			if userID != 9 || accountID != "acc-9" {
				t.Fatalf("unexpected delete params userID=%d accountID=%s", userID, accountID)
			}
			return nil
		},
	}

	unipileClient := &mockUnipileClient{
		deleteAccountFunc: func(accountID string) error {
			deleteAccountCalled = true
			if accountID != "acc-9" {
				t.Fatalf("unexpected account ID %s", accountID)
			}
			return nil
		},
	}

	txRepo := &mockTxRepo{
		doFunc: func(ctx context.Context, fn func(*repository.Repositories) error) error {
			return fn(&repository.Repositories{Account: accountRepo})
		},
	}

	uc := NewAccountUsecase(txRepo, &mockAccountRepo{}, unipileClient, logrus.New())

	if err := uc.DisconnectLinkedIn(ctx, 9, "acc-9"); err != nil {
		t.Fatalf("DisconnectLinkedIn returned error: %v", err)
	}

	if !deleteCalled {
		t.Fatalf("expected account repository delete to be called")
	}

	if !deleteAccountCalled {
		t.Fatalf("expected unipile delete account to be called")
	}
}

func TestDisconnectLinkedIn_UnipileAccountNotFound(t *testing.T) {
	ctx := context.Background()

	accountRepo := &mockAccountRepo{
		deleteByUserIDAndAccountIDFunc: func(_ context.Context, userID uint, accountID string) error { return nil },
	}

	unipileClient := &mockUnipileClient{
		deleteAccountFunc: func(accountID string) error {
			return service.ErrUnipileAccountNotFound
		},
	}

	txRepo := &mockTxRepo{
		doFunc: func(ctx context.Context, fn func(*repository.Repositories) error) error {
			return fn(&repository.Repositories{Account: accountRepo})
		},
	}

	uc := NewAccountUsecase(txRepo, &mockAccountRepo{}, unipileClient, logrus.New())

	if err := uc.DisconnectLinkedIn(ctx, 1, "unknown"); err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestListUserAccounts(t *testing.T) {
	ctx := context.Background()
	expected := []*entity.Account{{AccountID: "a"}, {AccountID: "b"}}
	var called bool

	accountRepo := &mockAccountRepo{
		getByUserIDFunc: func(_ context.Context, userID uint) ([]*entity.Account, error) {
			called = true
			if userID != 77 {
				t.Fatalf("unexpected userID %d", userID)
			}
			return expected, nil
		},
	}

	uc := NewAccountUsecase(&mockTxRepo{}, accountRepo, &mockUnipileClient{}, logrus.New())

	accounts, err := uc.ListUserAccounts(ctx, 77)
	if err != nil {
		t.Fatalf("ListUserAccounts returned error: %v", err)
	}

	if !called {
		t.Fatalf("expected repository to be called")
	}

	if len(accounts) != len(expected) {
		t.Fatalf("expected %d accounts, got %d", len(expected), len(accounts))
	}
}

func TestWaitForAccountValidation_Success(t *testing.T) {
	ctx := context.Background()
	accountWithStatus := &entity.AccountWithStatus{
		Account: entity.Account{
			UserID:        1,
			AccountID:     "acc-123",
			Provider:      "LINKEDIN",
			CurrentStatus: "PENDING",
		},
		CurrentStatus:       "PENDING",
		Checkpoint:          "IN_APP_VALIDATION",
		CheckpointExpiresAt: time.Now().Add(5 * time.Minute),
	}
	var updatedAccount *entity.Account

	accountRepo := &mockAccountRepo{
		getWithStatusFunc: func(_ context.Context, userID uint, accountID, checkpoint string) (*entity.AccountWithStatus, error) {
			if userID != 1 || accountID != "acc-123" || checkpoint != "IN_APP_VALIDATION" {
				t.Fatalf("unexpected lookup params userID=%d accountID=%s checkpoint=%s", userID, accountID, checkpoint)
			}
			return accountWithStatus, nil
		},
		updateFunc: func(_ context.Context, account *entity.Account) error {
			updatedAccount = account
			return nil
		},
	}

	unipileClient := &mockUnipileClient{
		getAccountWithLongPollingFunc: func(accountID string, timeout time.Duration) (*service.Account, error) {
			if accountID != "acc-123" {
				t.Fatalf("unexpected account ID %s", accountID)
			}
			if timeout != 300*time.Second {
				t.Fatalf("unexpected timeout %v", timeout)
			}
			return &service.Account{
				Sources: []service.AccountSource{
					{Status: "OK"},
				},
			}, nil
		},
	}

	uc := NewAccountUsecase(&mockTxRepo{}, accountRepo, unipileClient, logrus.New())

	account, err := uc.WaitForAccountValidation(ctx, 1, "acc-123", 300*time.Second)
	if err != nil {
		t.Fatalf("WaitForAccountValidation returned error: %v", err)
	}

	if updatedAccount == nil {
		t.Fatalf("expected account update to be called")
	}

	if account.CurrentStatus != "OK" {
		t.Fatalf("expected account status OK, got %s", account.CurrentStatus)
	}
}

func TestWaitForAccountValidation_AccountNotFound(t *testing.T) {
	ctx := context.Background()

	accountRepo := &mockAccountRepo{
		getWithStatusFunc: func(_ context.Context, userID uint, accountID, checkpoint string) (*entity.AccountWithStatus, error) {
			return nil, repository.ErrAccountNotFound
		},
	}

	uc := NewAccountUsecase(&mockTxRepo{}, accountRepo, &mockUnipileClient{}, logrus.New())

	_, err := uc.WaitForAccountValidation(ctx, 1, "missing", 300*time.Second)
	if err == nil {
		t.Fatalf("expected error but got nil")
	}

	var codedErr *errs.CodedError
	if !errors.As(err, &codedErr) {
		t.Fatalf("expected coded error, got %v", err)
	}

	if codedErr.Kind != errs.ValidationErrorKind {
		t.Fatalf("expected validation error kind, got %s", codedErr.Kind)
	}
}

func TestWaitForAccountValidation_AlreadyOK(t *testing.T) {
	ctx := context.Background()
	accountWithStatus := &entity.AccountWithStatus{
		Account: entity.Account{
			UserID:        1,
			AccountID:     "acc-123",
			Provider:      "LINKEDIN",
			CurrentStatus: "OK",
		},
		CurrentStatus: "OK",
	}

	accountRepo := &mockAccountRepo{
		getWithStatusFunc: func(_ context.Context, userID uint, accountID, checkpoint string) (*entity.AccountWithStatus, error) {
			return accountWithStatus, nil
		},
	}

	uc := NewAccountUsecase(&mockTxRepo{}, accountRepo, &mockUnipileClient{}, logrus.New())

	account, err := uc.WaitForAccountValidation(ctx, 1, "acc-123", 300*time.Second)
	if err != nil {
		t.Fatalf("WaitForAccountValidation returned error: %v", err)
	}

	if account.CurrentStatus != "OK" {
		t.Fatalf("expected account status OK, got %s", account.CurrentStatus)
	}
}

func TestWaitForAccountValidation_UnipileAccountNotFound(t *testing.T) {
	ctx := context.Background()
	accountWithStatus := &entity.AccountWithStatus{
		Account: entity.Account{
			UserID:        1,
			AccountID:     "acc-123",
			Provider:      "LINKEDIN",
			CurrentStatus: "PENDING",
		},
		CurrentStatus:       "PENDING",
		Checkpoint:          "IN_APP_VALIDATION",
		CheckpointExpiresAt: time.Now().Add(5 * time.Minute),
	}

	accountRepo := &mockAccountRepo{
		getWithStatusFunc: func(_ context.Context, userID uint, accountID, checkpoint string) (*entity.AccountWithStatus, error) {
			return accountWithStatus, nil
		},
	}

	unipileClient := &mockUnipileClient{
		getAccountWithLongPollingFunc: func(accountID string, timeout time.Duration) (*service.Account, error) {
			return nil, service.ErrUnipileAccountNotFound
		},
	}

	uc := NewAccountUsecase(&mockTxRepo{}, accountRepo, unipileClient, logrus.New())

	_, err := uc.WaitForAccountValidation(ctx, 1, "acc-123", 300*time.Second)
	if err == nil {
		t.Fatalf("expected error but got nil")
	}

	var codedErr *errs.CodedError
	if !errors.As(err, &codedErr) {
		t.Fatalf("expected coded error, got %v", err)
	}

	if codedErr.Kind != errs.ValidationErrorKind {
		t.Fatalf("expected validation error kind, got %s", codedErr.Kind)
	}
}
