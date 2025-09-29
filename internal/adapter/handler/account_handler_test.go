package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"unipile-connector/internal/domain/entity"
	"unipile-connector/internal/domain/errs"
	accountusecase "unipile-connector/internal/usecase/account"
)

type accountUsecaseMock struct {
	listUserAccountsFn   func(ctx context.Context, userID uint) ([]*entity.Account, error)
	disconnectLinkedInFn func(ctx context.Context, userID uint, accountID string) error
	connectLinkedInFn    func(ctx context.Context, userID uint, req *accountusecase.ConnectLinkedInRequest) (*entity.Account, error)
	solveCheckpointFn    func(ctx context.Context, userID uint, req *accountusecase.SolveCheckpointRequest) (*entity.Account, error)
}

var _ accountusecase.Usecase = (*accountUsecaseMock)(nil)

func (m *accountUsecaseMock) ListUserAccounts(ctx context.Context, userID uint) ([]*entity.Account, error) {
	if m.listUserAccountsFn == nil {
		return nil, nil
	}
	return m.listUserAccountsFn(ctx, userID)
}

func (m *accountUsecaseMock) DisconnectLinkedIn(ctx context.Context, userID uint, accountID string) error {
	if m.disconnectLinkedInFn == nil {
		return nil
	}
	return m.disconnectLinkedInFn(ctx, userID, accountID)
}

func (m *accountUsecaseMock) ConnectLinkedInAccount(ctx context.Context, userID uint, req *accountusecase.ConnectLinkedInRequest) (*entity.Account, error) {
	if m.connectLinkedInFn == nil {
		return nil, nil
	}
	return m.connectLinkedInFn(ctx, userID, req)
}

func (m *accountUsecaseMock) SolveCheckpoint(ctx context.Context, userID uint, req *accountusecase.SolveCheckpointRequest) (*entity.Account, error) {
	if m.solveCheckpointFn == nil {
		return nil, nil
	}
	return m.solveCheckpointFn(ctx, userID, req)
}

func TestAccountHandler_ListUserAccounts_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &AccountHandlerImpl{
		accountUsecase: &accountUsecaseMock{
			listUserAccountsFn: func(ctx context.Context, userID uint) ([]*entity.Account, error) {
				return []*entity.Account{
					{
						UserID:        userID,
						Provider:      "LINKEDIN",
						AccountID:     "acc-123",
						CurrentStatus: "OK",
					},
				}, nil
			},
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest(http.MethodGet, "/api/v1/accounts", nil)
	c.Set("user_id", uint(42))

	h.ListUserAccounts(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["message"] != "Accounts retrieved successfully" {
		t.Fatalf("unexpected message: %v", resp["message"])
	}

	accounts, ok := resp["accounts"].([]interface{})
	if !ok || len(accounts) != 1 {
		t.Fatalf("expected accounts array with one item, got %v", resp["accounts"])
	}
}

func TestAccountHandler_ConnectLinkedIn_ValidationError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &AccountHandlerImpl{accountUsecase: &accountUsecaseMock{}}

	body := bytes.NewBufferString(`{"type":"credentials","username":"user"}`) // missing password
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/accounts/linkedin/connect", body)
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Set("user_id", uint(1))

	h.ConnectLinkedIn(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp errs.CodedError
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Kind != errs.ValidationErrorKind {
		t.Fatalf("expected validation error kind, got %s", resp.Kind)
	}
}

func TestAccountHandler_SolveCheckpoint_InvalidCode(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &AccountHandlerImpl{
		accountUsecase: &accountUsecaseMock{
			solveCheckpointFn: func(ctx context.Context, userID uint, req *accountusecase.SolveCheckpointRequest) (*entity.Account, error) {
				return nil, errs.ErrInvalidCodeOrExpiredCheckpoint
			},
		},
	}

	payload := map[string]string{
		"account_id": "acc-1",
		"code":       "000000",
	}
	raw, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/accounts/linkedin/checkpoint", bytes.NewBuffer(raw))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Set("user_id", uint(1))

	h.SolveCheckpoint(c)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
	}

	var resp errs.CodedError
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Kind != errs.ValidationErrorKind {
		t.Fatalf("expected validation error kind, got %s", resp.Kind)
	}
	if resp.Message != "Invalid code or expired checkpoint" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
}

func TestAccountHandler_DisconnectLinkedIn_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var called bool
	h := &AccountHandlerImpl{
		accountUsecase: &accountUsecaseMock{
			disconnectLinkedInFn: func(ctx context.Context, userID uint, accountID string) error {
				called = true
				if userID != 5 || accountID != "acc-77" {
					t.Fatalf("unexpected arguments: %d, %s", userID, accountID)
				}
				return nil
			},
		},
	}

	payload := map[string]string{"account_id": "acc-77"}
	raw, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/accounts/linkedin", bytes.NewBuffer(raw))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Set("user_id", uint(5))

	h.DisconnectLinkedIn(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}
	if !called {
		t.Fatal("expected usecase to be called")
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp["message"] != "LinkedIn account disconnected successfully" {
		t.Fatalf("unexpected message: %v", resp["message"])
	}
}
