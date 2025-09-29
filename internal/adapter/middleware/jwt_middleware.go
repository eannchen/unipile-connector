package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"unipile-connector/internal/domain/service"
)

// JWTMiddleware handles JWT authentication
type JWTMiddleware interface {
	AuthMiddleware() gin.HandlerFunc
}

// JWTMiddlewareImpl handles JWT authentication
type JWTMiddlewareImpl struct {
	jwtService service.JWTService
}

// NewJWTMiddleware creates a new JWT middleware
func NewJWTMiddleware(jwtService service.JWTService) JWTMiddleware {
	return &JWTMiddlewareImpl{
		jwtService: jwtService,
	}
}

// AuthMiddleware validates JWT tokens
func (m *JWTMiddlewareImpl) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check if header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token required"})
			c.Abort()
			return
		}

		// Validate token
		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("token_claims", claims)

		c.Next()
	}
}
