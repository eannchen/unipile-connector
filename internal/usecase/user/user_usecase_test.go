package user

import (
	"context"
	"errors"
	"testing"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"unipile-connector/internal/domain/entity"
	"unipile-connector/internal/domain/errs"
	"unipile-connector/internal/domain/repository"
	"unipile-connector/internal/domain/service"
)

type mockUserRepo struct {
	createFunc        func(ctx context.Context, user *entity.User) error
	getByIDFunc       func(ctx context.Context, id uint) (*entity.User, error)
	getByUsernameFunc func(ctx context.Context, username string) (*entity.User, error)
}

func (m *mockUserRepo) Create(ctx context.Context, user *entity.User) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, user)
	}
	return nil
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uint) (*entity.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	if m.getByUsernameFunc != nil {
		return m.getByUsernameFunc(ctx, username)
	}
	return nil, nil
}

type mockJWTService struct {
	generateTokenFunc func(userID uint, username string) (string, error)
	validateTokenFunc func(token string) (*service.Claims, error)
	refreshTokenFunc  func(token string) (string, error)
}

func (m *mockJWTService) GenerateToken(userID uint, username string) (string, error) {
	if m.generateTokenFunc != nil {
		return m.generateTokenFunc(userID, username)
	}
	return "", nil
}

func (m *mockJWTService) ValidateToken(token string) (*service.Claims, error) {
	if m.validateTokenFunc != nil {
		return m.validateTokenFunc(token)
	}
	return nil, nil
}

func (m *mockJWTService) RefreshToken(token string) (string, error) {
	if m.refreshTokenFunc != nil {
		return m.refreshTokenFunc(token)
	}
	return "", nil
}

func (m *mockJWTService) BlacklistToken(token string) {
	// Mock implementation - do nothing
}

func TestGetUserByID_Success(t *testing.T) {
	ctx := context.Background()
	expected := &entity.User{ID: 1, Username: "alice"}

	userRepo := &mockUserRepo{
		getByIDFunc: func(_ context.Context, id uint) (*entity.User, error) {
			if id != 1 {
				t.Fatalf("unexpected id: %d", id)
			}
			return expected, nil
		},
	}

	uc := NewUserUsecase(userRepo, &mockJWTService{}, logrus.New())

	user, err := uc.GetUserByID(ctx, 1)
	if err != nil {
		t.Fatalf("GetUserByID returned error: %v", err)
	}

	if user != expected {
		t.Fatalf("expected %+v, got %+v", expected, user)
	}
}

func TestGetUserByID_NotFound(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepo{
		getByIDFunc: func(_ context.Context, id uint) (*entity.User, error) {
			return nil, repository.ErrRecordNotFound
		},
	}

	uc := NewUserUsecase(userRepo, &mockJWTService{}, logrus.New())

	_, err := uc.GetUserByID(ctx, 99)
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

func TestCreateUser_Success(t *testing.T) {
	ctx := context.Background()
	var persistedUser *entity.User

	userRepo := &mockUserRepo{
		createFunc: func(_ context.Context, user *entity.User) error {
			persistedUser = user
			return nil
		},
	}

	uc := NewUserUsecase(userRepo, &mockJWTService{}, logrus.New())

	user, err := uc.CreateUser(ctx, "bob", "secret")
	if err != nil {
		t.Fatalf("CreateUser returned error: %v", err)
	}

	if user.Username != "bob" {
		t.Fatalf("expected username bob, got %s", user.Username)
	}

	if persistedUser == nil {
		t.Fatalf("expected user to be persisted")
	}

	if persistedUser.Password == "secret" {
		t.Fatalf("expected password to be hashed")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(persistedUser.Password), []byte("secret")); err != nil {
		t.Fatalf("stored password not hash of secret: %v", err)
	}
}

func TestCreateUser_DuplicateUsername(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepo{
		createFunc: func(_ context.Context, user *entity.User) error {
			return repository.ErrDuplicateKey
		},
	}

	uc := NewUserUsecase(userRepo, &mockJWTService{}, logrus.New())

	_, err := uc.CreateUser(ctx, "duplicate", "pw")
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

func TestCreateUser_HashError(t *testing.T) {
	ctx := context.Background()

	oldGenerateFromPassword := bcryptGenerateFromPassword
	bcryptGenerateFromPassword = func(password []byte, cost int) ([]byte, error) {
		return nil, errors.New("hash fail")
	}
	defer func() { bcryptGenerateFromPassword = oldGenerateFromPassword }()

	uc := NewUserUsecase(&mockUserRepo{}, &mockJWTService{}, logrus.New())

	_, err := uc.CreateUser(ctx, "charlie", "pw")
	if err == nil {
		t.Fatalf("expected error but got nil")
	}

	var codedErr *errs.CodedError
	if !errors.As(err, &codedErr) {
		t.Fatalf("expected coded error, got %v", err)
	}

	if codedErr.Kind != errs.SystemErrorKind {
		t.Fatalf("expected system error kind, got %s", codedErr.Kind)
	}
}

func TestAuthenticateUser_Success(t *testing.T) {
	ctx := context.Background()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.DefaultCost)
	userRepo := &mockUserRepo{
		getByUsernameFunc: func(_ context.Context, username string) (*entity.User, error) {
			if username != "dana" {
				t.Fatalf("unexpected username %s", username)
			}
			return &entity.User{ID: 5, Username: username, Password: string(hashed)}, nil
		},
	}

	jwtService := &mockJWTService{
		generateTokenFunc: func(userID uint, username string) (string, error) {
			if userID != 5 || username != "dana" {
				t.Fatalf("unexpected token params userID=%d username=%s", userID, username)
			}
			return "token", nil
		},
	}

	uc := NewUserUsecase(userRepo, jwtService, logrus.New())

	user, token, err := uc.AuthenticateUser(ctx, "dana", "pw")
	if err != nil {
		t.Fatalf("AuthenticateUser returned error: %v", err)
	}

	if user.Username != "dana" {
		t.Fatalf("unexpected user: %+v", user)
	}

	if token != "token" {
		t.Fatalf("expected token 'token', got %s", token)
	}
}

func TestAuthenticateUser_NotFound(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepo{
		getByUsernameFunc: func(_ context.Context, username string) (*entity.User, error) {
			return nil, repository.ErrRecordNotFound
		},
	}

	uc := NewUserUsecase(userRepo, &mockJWTService{}, logrus.New())

	_, _, err := uc.AuthenticateUser(ctx, "nobody", "pw")
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

func TestAuthenticateUser_InvalidPassword(t *testing.T) {
	ctx := context.Background()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.DefaultCost)
	userRepo := &mockUserRepo{
		getByUsernameFunc: func(_ context.Context, username string) (*entity.User, error) {
			return &entity.User{Username: username, Password: string(hashed)}, nil
		},
	}

	uc := NewUserUsecase(userRepo, &mockJWTService{}, logrus.New())

	_, _, err := uc.AuthenticateUser(ctx, "user", "wrong")
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

func TestAuthenticateUser_TokenError(t *testing.T) {
	ctx := context.Background()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.DefaultCost)
	userRepo := &mockUserRepo{
		getByUsernameFunc: func(_ context.Context, username string) (*entity.User, error) {
			return &entity.User{ID: 5, Username: username, Password: string(hashed)}, nil
		},
	}

	jwtService := &mockJWTService{
		generateTokenFunc: func(userID uint, username string) (string, error) {
			return "", errors.New("token fail")
		},
	}

	uc := NewUserUsecase(userRepo, jwtService, logrus.New())

	_, _, err := uc.AuthenticateUser(ctx, "dana", "pw")
	if err == nil {
		t.Fatalf("expected error but got nil")
	}

	var codedErr *errs.CodedError
	if !errors.As(err, &codedErr) {
		t.Fatalf("expected coded error, got %v", err)
	}

	if codedErr.Kind != errs.SystemErrorKind {
		t.Fatalf("expected system error kind, got %s", codedErr.Kind)
	}
}

func TestRefreshToken_Success(t *testing.T) {
	ctx := context.Background()

	jwtService := &mockJWTService{
		refreshTokenFunc: func(token string) (string, error) {
			if token != "old" {
				t.Fatalf("unexpected token %s", token)
			}
			return "new", nil
		},
	}

	uc := NewUserUsecase(&mockUserRepo{}, jwtService, logrus.New())

	token, err := uc.RefreshToken(ctx, "old")
	if err != nil {
		t.Fatalf("RefreshToken returned error: %v", err)
	}

	if token != "new" {
		t.Fatalf("expected new token, got %s", token)
	}
}

func TestRefreshToken_Error(t *testing.T) {
	ctx := context.Background()

	jwtService := &mockJWTService{
		refreshTokenFunc: func(token string) (string, error) {
			return "", errors.New("refresh fail")
		},
	}

	uc := NewUserUsecase(&mockUserRepo{}, jwtService, logrus.New())

	_, err := uc.RefreshToken(ctx, "bad")
	if err == nil {
		t.Fatalf("expected error but got nil")
	}

	var codedErr *errs.CodedError
	if !errors.As(err, &codedErr) {
		t.Fatalf("expected coded error, got %v", err)
	}

	if codedErr.Kind != errs.SystemErrorKind {
		t.Fatalf("expected system error kind, got %s", codedErr.Kind)
	}
}
