package middleware

import (
	"net/http"
	"strconv"
	"unipile-connector/internal/usecase/user"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AuthMiddleware handles authentication
type AuthMiddleware struct {
	userUsecase *user.UserUsecase
	logger      *logrus.Logger
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(userUsecase *user.UserUsecase, logger *logrus.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		userUsecase: userUsecase,
		logger:      logger,
	}
}

// SimpleAuthMiddleware provides simple session-based authentication
// In a real application, you would use JWT tokens or proper session management
func (m *AuthMiddleware) SimpleAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// For demo purposes, we'll use a simple header-based auth
		// In production, use proper JWT or session management
		userIDHeader := c.GetHeader("X-User-ID")
		if userIDHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		userID, err := strconv.ParseUint(userIDHeader, 10, 32)
		if err != nil {
			m.logger.WithError(err).Error("Invalid user ID in header")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		// Verify user exists
		_, err = m.userUsecase.GetUserByID(c.Request.Context(), uint(userID))
		if err != nil {
			m.logger.WithError(err).Error("User not found")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			c.Abort()
			return
		}

		// Set user ID in context
		c.Set("user_id", uint(userID))
		c.Next()
	}
}

// OptionalAuthMiddleware provides optional authentication
func (m *AuthMiddleware) OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDHeader := c.GetHeader("X-User-ID")
		if userIDHeader != "" {
			userID, err := strconv.ParseUint(userIDHeader, 10, 32)
			if err == nil {
				// Verify user exists
				_, err = m.userUsecase.GetUserByID(c.Request.Context(), uint(userID))
				if err == nil {
					c.Set("user_id", uint(userID))
				}
			}
		}
		c.Next()
	}
}

// CORSMiddleware handles CORS
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-User-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
