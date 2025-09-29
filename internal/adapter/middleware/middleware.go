package middleware

import "github.com/gin-gonic/gin"

// Middlewares handles all middlewares
type Middlewares struct {
	CORSMiddleware      gin.HandlerFunc
	JWTMiddleware       gin.HandlerFunc
	RateLimitMiddleware gin.HandlerFunc
}

// NewMiddlewares creates a new middleware
func NewMiddlewares(corsMiddleware, jwtMiddleware, rateLimitMiddleware gin.HandlerFunc) *Middlewares {
	return &Middlewares{CORSMiddleware: corsMiddleware, JWTMiddleware: jwtMiddleware, RateLimitMiddleware: rateLimitMiddleware}
}
