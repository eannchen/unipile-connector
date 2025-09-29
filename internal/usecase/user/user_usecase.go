package user

import (
	"context"
	"errors"

	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"

	"unipile-connector/internal/domain/entity"
	"unipile-connector/internal/domain/repository"
	"unipile-connector/pkg/jwt"
)

// UserUsecase handles user business logic
type UserUsecase struct {
	userRepo   repository.UserRepository
	jwtService *jwt.JWTService
	logger     *logrus.Logger
}

// NewUserUsecase creates a new user usecase
func NewUserUsecase(userRepo repository.UserRepository, jwtService *jwt.JWTService, logger *logrus.Logger) *UserUsecase {
	return &UserUsecase{
		userRepo:   userRepo,
		jwtService: jwtService,
		logger:     logger,
	}
}

// GetUserByID retrieves a user by ID
func (u *UserUsecase) GetUserByID(ctx context.Context, id uint) (*entity.User, error) {
	return u.userRepo.GetByID(ctx, id)
}

// CreateUser creates a new user
func (u *UserUsecase) CreateUser(ctx context.Context, username, password string) (*entity.User, error) {
	// Check if user already exists
	existingUser, _ := u.userRepo.GetByUsername(ctx, username)
	if existingUser != nil {
		return nil, errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &entity.User{
		Username: username,
		Password: string(hashedPassword),
	}

	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// AuthenticateUser authenticates a user with username and password
func (u *UserUsecase) AuthenticateUser(ctx context.Context, username, password string) (*entity.User, string, error) {
	user, err := u.userRepo.GetByUsername(ctx, username)
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, "", errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := u.jwtService.GenerateToken(user.ID, user.Username)
	if err != nil {
		return nil, "", errors.New("failed to generate token")
	}

	return user, token, nil
}

// RefreshToken refreshes a token
func (u *UserUsecase) RefreshToken(ctx context.Context, token string) (string, error) {
	return u.jwtService.RefreshToken(token)
}
