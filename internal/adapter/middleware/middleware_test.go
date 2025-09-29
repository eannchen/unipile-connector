package middleware

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestNewMiddlewares(t *testing.T) {
	cors := gin.HandlerFunc(func(c *gin.Context) {})
	jwt := gin.HandlerFunc(func(c *gin.Context) {})
	rate := gin.HandlerFunc(func(c *gin.Context) {})

	m := NewMiddlewares(cors, jwt, rate)
	require.NotNil(t, m)
	require.IsType(t, cors, m.CORSMiddleware)
	require.IsType(t, jwt, m.JWTMiddleware)
	require.IsType(t, rate, m.RateLimitMiddleware)
}
