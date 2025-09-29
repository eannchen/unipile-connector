package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"unipile-connector/internal/domain/entity"
	"unipile-connector/internal/domain/errs"
	userusecase "unipile-connector/internal/usecase/user"
)

type userUsecaseMock struct {
	createUserFn       func(ctx context.Context, username, password string) (*entity.User, error)
	authenticateUserFn func(ctx context.Context, username, password string) (*entity.User, string, error)
	refreshTokenFn     func(ctx context.Context, token string) (string, error)
	getUserByIDFn      func(ctx context.Context, id uint) (*entity.User, error)
	blacklistTokenFn   func(ctx context.Context, token string)
}

func (m *userUsecaseMock) CreateUser(ctx context.Context, username, password string) (*entity.User, error) {
	if m.createUserFn == nil {
		return nil, nil
	}
	return m.createUserFn(ctx, username, password)
}

func (m *userUsecaseMock) AuthenticateUser(ctx context.Context, username, password string) (*entity.User, string, error) {
	if m.authenticateUserFn == nil {
		return nil, "", nil
	}
	return m.authenticateUserFn(ctx, username, password)
}

func (m *userUsecaseMock) RefreshToken(ctx context.Context, token string) (string, error) {
	if m.refreshTokenFn == nil {
		return "", nil
	}
	return m.refreshTokenFn(ctx, token)
}

func (m *userUsecaseMock) GetUserByID(ctx context.Context, id uint) (*entity.User, error) {
	if m.getUserByIDFn == nil {
		return nil, nil
	}
	return m.getUserByIDFn(ctx, id)
}

func (m *userUsecaseMock) BlacklistToken(ctx context.Context, token string) {
	if m.blacklistTokenFn != nil {
		m.blacklistTokenFn(ctx, token)
	}
}

var _ userusecase.Usecase = (*userUsecaseMock)(nil)

func TestAuthHandler_Register_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &AuthHandlerImpl{
		userUsecase: &userUsecaseMock{
			createUserFn: func(ctx context.Context, username, password string) (*entity.User, error) {
				if username != "alice" || password != "securepass" {
					t.Fatalf("unexpected payload: %s %s", username, password)
				}
				return &entity.User{ID: 1, Username: username}, nil
			},
		},
	}

	payload := map[string]string{"username": "alice", "password": "securepass"}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Register(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["message"] != "User created successfully" {
		t.Fatalf("unexpected message: %v", resp["message"])
	}

	user, ok := resp["user"].(map[string]interface{})
	if !ok || user["username"] != "alice" {
		t.Fatalf("unexpected user payload: %v", resp["user"])
	}
}

func TestAuthHandler_Login_InvalidCredentials(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &AuthHandlerImpl{
		userUsecase: &userUsecaseMock{
			authenticateUserFn: func(ctx context.Context, username, password string) (*entity.User, string, error) {
				return nil, "", errs.WrapValidationError(errors.New("invalid credentials"), "Invalid credentials")
			},
		},
	}

	payload := map[string]string{"username": "bob", "password": "wrong"}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Login(c)

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

func TestAuthHandler_RefreshToken_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &AuthHandlerImpl{userUsecase: &userUsecaseMock{}}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	c.Request = req

	h.RefreshToken(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}

	var resp errs.CodedError
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Message != "Authorization header required" {
		t.Fatalf("unexpected message: %s", resp.Message)
	}
}

func TestAuthHandler_GetCurrentUser_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &AuthHandlerImpl{
		userUsecase: &userUsecaseMock{
			getUserByIDFn: func(ctx context.Context, id uint) (*entity.User, error) {
				if id != 10 {
					t.Fatalf("unexpected id: %d", id)
				}
				return &entity.User{ID: id, Username: "charlie"}, nil
			},
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", uint(10))
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/auth/me", nil)
	c.Request = req

	h.GetCurrentUser(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["message"] != "User retrieved successfully" {
		t.Fatalf("unexpected message: %v", resp["message"])
	}

	id, ok := resp["id"].(float64)
	if !ok || id != 10 {
		t.Fatalf("unexpected id: %v", resp["id"])
	}

	username, ok := resp["username"].(string)
	if !ok || username != "charlie" {
		t.Fatalf("unexpected username: %v", resp["username"])
	}
}

func TestAuthHandler_Login_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &AuthHandlerImpl{
		userUsecase: &userUsecaseMock{
			authenticateUserFn: func(ctx context.Context, username, password string) (*entity.User, string, error) {
				return &entity.User{ID: 2, Username: username}, "token123", nil
			},
		},
	}

	payload := map[string]string{"username": "bob", "password": "correct"}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req

	h.Login(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["token"] != "token123" {
		t.Fatalf("unexpected token: %v", resp["token"])
	}

	user, ok := resp["user"].(map[string]interface{})
	if !ok || user["username"] != "bob" {
		t.Fatalf("unexpected user: %v", resp["user"])
	}
}
