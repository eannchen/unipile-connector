package service

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

func TestJWTService_GenerateAndValidateToken(t *testing.T) {
	service := NewJWTService("secret", "issuer")

	token, err := service.GenerateToken(42, "alice")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := service.ValidateToken(token)
	require.NoError(t, err)
	require.Equal(t, uint(42), claims.UserID)
	require.Equal(t, "alice", claims.Username)
	require.Equal(t, "issuer", claims.Issuer)
	require.WithinDuration(t, time.Now().Add(24*time.Hour), claims.ExpiresAt.Time, 5*time.Second)
}

func TestJWTService_ValidateToken_Invalid(t *testing.T) {
	service := NewJWTService("secret", "issuer")

	_, err := service.ValidateToken("invalid.token.string")
	require.Error(t, err)
}

func TestJWTService_ValidateToken_UnexpectedMethod(t *testing.T) {
	service := NewJWTService("secret", "issuer")

	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	token := jwt.NewWithClaims(jwt.SigningMethodES256, jwt.MapClaims{"foo": "bar"})
	signed, err := token.SignedString(privKey)
	require.NoError(t, err)

	_, err = service.ValidateToken(signed)
	require.Error(t, err)
	require.Contains(t, err.Error(), "unexpected signing method")
}

func TestJWTService_RefreshToken(t *testing.T) {
	service := NewJWTService("secret", "issuer")

	token, err := service.GenerateToken(7, "bob")
	require.NoError(t, err)

	refreshed, err := service.RefreshToken(token)
	require.NoError(t, err)

	originalClaims, err := service.ValidateToken(token)
	require.NoError(t, err)
	newClaims, err := service.ValidateToken(refreshed)
	require.NoError(t, err)

	require.Equal(t, originalClaims.UserID, newClaims.UserID)
	require.Equal(t, originalClaims.Username, newClaims.Username)
	require.True(t, newClaims.ExpiresAt.Time.After(originalClaims.ExpiresAt.Time) || newClaims.ExpiresAt.Time.Equal(originalClaims.ExpiresAt.Time))
}
