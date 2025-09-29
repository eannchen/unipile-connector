package service

import (
	"context"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// TokenBlacklistService handles token blacklisting operations
type TokenBlacklistService interface {
	// AddToBlacklist adds a token to the blacklist
	AddToBlacklist(ctx context.Context, tokenString string)
	// IsBlacklisted checks if a token is blacklisted
	IsBlacklisted(ctx context.Context, tokenString string) (bool, error)
	// StartCleanup starts the cleanup goroutine for expired tokens
	StartCleanup(ctx context.Context)
	// StopCleanup stops the cleanup goroutine
	StopCleanup()
}

// BlacklistedToken represents a blacklisted token with its expiration time
type BlacklistedToken struct {
	Token     string
	ExpiresAt time.Time
}

// TokenBlacklistServiceImpl implements token blacklisting with memory store
type TokenBlacklistServiceImpl struct {
	blacklist map[string]time.Time // token -> expiration time
	mutex     sync.RWMutex
	stopCh    chan struct{}
}

// NewTokenBlacklistService creates a new token blacklist service
func NewTokenBlacklistService() TokenBlacklistService {
	return &TokenBlacklistServiceImpl{
		blacklist: make(map[string]time.Time),
		stopCh:    make(chan struct{}),
	}
}

// AddToBlacklist adds a token to the blacklist
func (t *TokenBlacklistServiceImpl) AddToBlacklist(ctx context.Context, tokenString string) {
	// Parse token to get expiration time
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// No need to verify the signature here, just extract claims
		return []byte(""), nil
	})

	// Default 24 hours
	expiresAt := time.Now().Add(24 * time.Hour)

	if err != nil {
		// It handles cases where the token might be malformed but we still want to blacklist it
		t.mutex.Lock()
		t.blacklist[tokenString] = expiresAt
		t.mutex.Unlock()
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if exp, exists := claims["exp"]; exists {
			if expFloat, ok := exp.(float64); ok {
				expiresAt = time.Unix(int64(expFloat), 0)
			}
		}
	}

	t.mutex.Lock()
	t.blacklist[tokenString] = expiresAt
	t.mutex.Unlock()

	return
}

// IsBlacklisted checks if a token is blacklisted
func (t *TokenBlacklistServiceImpl) IsBlacklisted(ctx context.Context, tokenString string) (bool, error) {
	t.mutex.RLock()
	defer t.mutex.RUnlock()

	expiresAt, exists := t.blacklist[tokenString]
	if !exists {
		return false, nil
	}

	// If token has expired, consider it not blacklisted (cleanup will remove it)
	if time.Now().After(expiresAt) {
		return false, nil
	}

	return true, nil
}

// StartCleanup starts the cleanup goroutine for expired tokens
func (t *TokenBlacklistServiceImpl) StartCleanup(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(5 * time.Minute) // Cleanup every 5 minutes
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				t.cleanupExpiredTokens()
			case <-t.stopCh:
				return
			case <-ctx.Done():
				return
			}
		}
	}()
}

// StopCleanup stops the cleanup goroutine
func (t *TokenBlacklistServiceImpl) StopCleanup() {
	close(t.stopCh)
}

// cleanupExpiredTokens removes expired tokens from the blacklist
func (t *TokenBlacklistServiceImpl) cleanupExpiredTokens() {
	now := time.Now()
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for token, expiresAt := range t.blacklist {
		if now.After(expiresAt) {
			delete(t.blacklist, token)
		}
	}
}
