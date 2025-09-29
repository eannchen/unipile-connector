package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"unipile-connector/internal/domain/errs"
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
			err := errs.WrapValidationError(errors.New("authorization header required"), "Authorization header required")
			c.AbortWithStatusJSON(http.StatusUnauthorized, err.(*errs.CodedError))
			return
		}

		// Check if header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			err := errs.WrapValidationError(errors.New("invalid authorization header format"), "Invalid authorization header format")
			c.AbortWithStatusJSON(http.StatusUnauthorized, err.(*errs.CodedError))
			return
		}

		// Extract token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			err := errs.WrapValidationError(errors.New("token required"), "Token required")
			c.AbortWithStatusJSON(http.StatusUnauthorized, err.(*errs.CodedError))
			return
		}

		// Validate token
		claims, err := m.jwtService.ValidateToken(tokenString)
		if err != nil {
			err := errs.WrapValidationError(errors.New("invalid or expired token"), "Invalid or expired token")
			c.AbortWithStatusJSON(http.StatusUnauthorized, err.(*errs.CodedError))
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("token_claims", claims)

		c.Next()
	}
}
