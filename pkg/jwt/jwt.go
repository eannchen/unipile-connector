package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"unipile-connector/internal/domain/service"
)

// Service handles JWT token operations
type Service struct {
	secretKey []byte
	issuer    string
}

// NewService creates a new JWT service
func NewService(secretKey string, issuer string) service.JWTService {
	return &Service{
		secretKey: []byte(secretKey),
		issuer:    issuer,
	}
}

// GenerateToken generates a new JWT token for a user
func (j *Service) GenerateToken(userID uint, username string) (string, error) {
	now := time.Now()
	claims := service.Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    j.issuer,
			Subject:   username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(24 * time.Hour)), // 24 hours
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secretKey)
}

// ValidateToken validates a JWT token and returns the claims
func (j *Service) ValidateToken(tokenString string) (*service.Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &service.Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*service.Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshToken generates a new token with extended expiration
func (j *Service) RefreshToken(tokenString string) (string, error) {
	claims, err := j.ValidateToken(tokenString)
	if err != nil {
		return "", err
	}

	// Generate new token with extended expiration
	return j.GenerateToken(claims.UserID, claims.Username)
}
