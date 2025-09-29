package service

import "github.com/golang-jwt/jwt/v5"

// JWTService handles JWT token operations
type JWTService interface {
	GenerateToken(userID uint, username string) (string, error)
	ValidateToken(tokenString string) (*Claims, error)
	RefreshToken(tokenString string) (string, error)
}

// Claims represents JWT claims
type Claims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}
