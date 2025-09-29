package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService handles JWT token operations
type JWTService interface {
	GenerateToken(userID uint, username string) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
	RefreshToken(tokenString string) (string, error)
	BlacklistToken(tokenString string)
}

// Claims represents JWT claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

// JWTServiceImpl handles JWT token operations
type JWTServiceImpl struct {
	secretKey        []byte
	issuer           string
	blacklistService TokenBlacklistService
}

// NewJWTService creates a new JWT service
func NewJWTService(secretKey string, issuer string, blacklistService TokenBlacklistService) JWTService {
	return &JWTServiceImpl{
		secretKey:        []byte(secretKey),
		issuer:           issuer,
		blacklistService: blacklistService,
	}
}

// GenerateToken generates a new JWT token for a user
func (j *JWTServiceImpl) GenerateToken(userID uint, username string) (string, error) {
	now := time.Now()

	// Generate a unique JWT ID
	jtiBytes := make([]byte, 16)
	if _, err := rand.Read(jtiBytes); err != nil {
		return "", err
	}
	jti := hex.EncodeToString(jtiBytes)

	claims := Claims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        jti, // JWT ID for uniqueness
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
func (j *JWTServiceImpl) ValidateToken(tokenString string) (*Claims, error) {
	// Check if token is blacklisted first
	isBlacklisted, err := j.blacklistService.IsBlacklisted(context.Background(), tokenString)
	if err != nil {
		return nil, err
	}
	if isBlacklisted {
		return nil, errors.New("token is blacklisted")
	}

	return j.validateTokenWithoutBlacklist(tokenString)
}

// RefreshToken generates a new token with extended expiration
func (j *JWTServiceImpl) RefreshToken(tokenString string) (string, error) {
	claims, err := j.validateTokenWithoutBlacklist(tokenString)
	if err != nil {
		return "", err
	}

	// Blacklist the old token
	j.BlacklistToken(tokenString)

	// Generate new token with extended expiration
	return j.GenerateToken(claims.UserID, claims.Username)
}

// validateTokenWithoutBlacklist validates a JWT token without checking blacklist
func (j *JWTServiceImpl) validateTokenWithoutBlacklist(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return j.secretKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// BlacklistToken adds a token to the blacklist
func (j *JWTServiceImpl) BlacklistToken(tokenString string) {
	j.blacklistService.AddToBlacklist(context.Background(), tokenString)
}
