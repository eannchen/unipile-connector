package user

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"unipile-connector/internal/domain/entity"
	"unipile-connector/internal/domain/errs"
	"unipile-connector/internal/domain/repository"
	"unipile-connector/internal/domain/service"
)

// UserUsecase handles user business logic
type UserUsecase interface {
	GetUserByID(ctx context.Context, id uint) (*entity.User, error)
	CreateUser(ctx context.Context, username, password string) (*entity.User, error)
	AuthenticateUser(ctx context.Context, username, password string) (*entity.User, string, error)
	RefreshToken(ctx context.Context, token string) (string, error)
}

// UserUsecaseImpl handles user business logic
type UserUsecaseImpl struct {
	userRepo   repository.UserRepository
	jwtService service.JWTService
	logger     *logrus.Logger
}

// NewUserUsecase creates a new user usecase
func NewUserUsecase(userRepo repository.UserRepository, jwtService service.JWTService, logger *logrus.Logger) UserUsecase {
	return &UserUsecaseImpl{
		userRepo:   userRepo,
		jwtService: jwtService,
		logger:     logger,
	}
}

// GetUserByID retrieves a user by ID
func (u *UserUsecaseImpl) GetUserByID(ctx context.Context, id uint) (*entity.User, error) {
	user, err := u.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errs.WrapInternalError(err, "Failed to get user by ID")
	}
	return user, nil
}

// CreateUser creates a new user
func (u *UserUsecaseImpl) CreateUser(ctx context.Context, username, password string) (*entity.User, error) {
	// Check if user already exists
	existingUser, err := u.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, errs.WrapInternalError(err, "Failed to get user by username")
	}
	if existingUser != nil {
		return nil, errs.WrapValidationError(errors.New("username already exists"), "Username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errs.WrapInternalError(err, "Failed to hash password")
	}

	user := &entity.User{
		Username: username,
		Password: string(hashedPassword),
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, errs.WrapInternalError(err, "Failed to create user")
	}

	return user, nil
}

// AuthenticateUser authenticates a user with username and password
func (u *UserUsecaseImpl) AuthenticateUser(ctx context.Context, username, password string) (*entity.User, string, error) {
	user, err := u.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, "", errs.WrapValidationError(errors.New("invalid credentials"), "Invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, "", errs.WrapValidationError(errors.New("invalid credentials"), "Invalid credentials")
	}

	// Generate JWT token
	token, err := u.jwtService.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, "", errs.WrapInternalError(err, "Failed to generate token")
	}

	return user, token, nil
}

// RefreshToken refreshes a token
func (u *UserUsecaseImpl) RefreshToken(ctx context.Context, token string) (string, error) {
	newToken, err := u.jwtService.RefreshToken(token)
	if err != nil {
		return "", errs.WrapInternalError(err, "Failed to refresh token")
	}
	return newToken, nil
}
