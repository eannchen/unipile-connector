package service

import (
	"context"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestTokenBlacklistService(t *testing.T) {
	service := NewTokenBlacklistService()
	ctx := context.Background()

	// Test adding a token to blacklist
	token := "test-token-123"
	service.AddToBlacklist(ctx, token)

	// Test checking if token is blacklisted
	isBlacklisted, err := service.IsBlacklisted(ctx, token)
	if err != nil {
		t.Fatalf("Failed to check if token is blacklisted: %v", err)
	}
	if !isBlacklisted {
		t.Error("Token should be blacklisted")
	}

	// Test checking non-blacklisted token
	isBlacklisted, err = service.IsBlacklisted(ctx, "non-blacklisted-token")
	if err != nil {
		t.Fatalf("Failed to check non-blacklisted token: %v", err)
	}
	if isBlacklisted {
		t.Error("Non-blacklisted token should not be blacklisted")
	}
}

func TestTokenBlacklistServiceWithJWT(t *testing.T) {
	service := NewTokenBlacklistService()
	ctx := context.Background()

	// Create a test JWT token
	claims := jwt.MapClaims{
		"user_id":  uint(1),
		"username": "testuser",
		"exp":      time.Now().Add(time.Hour).Unix(), // Expires in 1 hour
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte("test-secret"))
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}

	// Add token to blacklist
	service.AddToBlacklist(ctx, tokenString)

	// Check if token is blacklisted
	isBlacklisted, err := service.IsBlacklisted(ctx, tokenString)
	if err != nil {
		t.Fatalf("Failed to check if token is blacklisted: %v", err)
	}
	if !isBlacklisted {
		t.Error("Token should be blacklisted")
	}
}

func TestTokenBlacklistCleanup(t *testing.T) {
	service := &TokenBlacklistServiceImpl{
		blacklist: make(map[string]time.Time),
		stopCh:    make(chan struct{}),
	}

	// Add an expired token
	expiredToken := "expired-token"
	service.blacklist[expiredToken] = time.Now().Add(-time.Hour) // Expired 1 hour ago

	// Add a valid token
	validToken := "valid-token"
	service.blacklist[validToken] = time.Now().Add(time.Hour) // Expires in 1 hour

	// Run cleanup
	service.cleanupExpiredTokens()

	// Check that expired token is removed
	if _, exists := service.blacklist[expiredToken]; exists {
		t.Error("Expired token should be removed from blacklist")
	}

	// Check that valid token is still there
	if _, exists := service.blacklist[validToken]; !exists {
		t.Error("Valid token should still be in blacklist")
	}
}
