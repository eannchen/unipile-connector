package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"unipile-connector/internal/domain/errs"
	"unipile-connector/internal/domain/service"
)

type jwtServiceMock struct {
	validateFn func(tokenString string) (*service.Claims, error)
}

func (m *jwtServiceMock) GenerateToken(userID uint, username string) (string, error) {
	return "", nil
}

func (m *jwtServiceMock) ValidateToken(tokenString string) (*service.Claims, error) {
	if m.validateFn == nil {
		return nil, nil
	}
	return m.validateFn(tokenString)
}

func (m *jwtServiceMock) RefreshToken(tokenString string) (string, error) {
	return "", nil
}

var _ service.JWTService = (*jwtServiceMock)(nil)

func TestJWTMiddleware_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := NewJWTMiddleware(&jwtServiceMock{}).AuthMiddleware()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	c.Request = req

	middleware(c)

	require.Equal(t, http.StatusUnauthorized, w.Code)

	var resp errs.CodedError
	require.NoError(t, errsJSONUnmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, errs.ValidationErrorKind, resp.Kind)
	require.Equal(t, "Authorization header required", resp.Message)
}

func TestJWTMiddleware_InvalidHeaderFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := NewJWTMiddleware(&jwtServiceMock{}).AuthMiddleware()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token abc")
	c.Request = req

	middleware(c)

	require.Equal(t, http.StatusUnauthorized, w.Code)

	var resp errs.CodedError
	require.NoError(t, errsJSONUnmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "Invalid authorization header format", resp.Message)
}

func TestJWTMiddleware_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := NewJWTMiddleware(&jwtServiceMock{
		validateFn: func(tokenString string) (*service.Claims, error) {
			return nil, errors.New("boom")
		},
	}).AuthMiddleware()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer badtoken")
	c.Request = req

	middleware(c)

	require.Equal(t, http.StatusUnauthorized, w.Code)

	var resp errs.CodedError
	require.NoError(t, errsJSONUnmarshal(w.Body.Bytes(), &resp))
	require.Equal(t, "Invalid or expired token", resp.Message)
}

func TestJWTMiddleware_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	middleware := NewJWTMiddleware(&jwtServiceMock{
		validateFn: func(tokenString string) (*service.Claims, error) {
			require.Equal(t, "goodtoken", tokenString)
			return &service.Claims{UserID: 42, Username: "alice"}, nil
		},
	}).AuthMiddleware()

	called := false
	engine := gin.New()
	engine.Use(middleware)
	engine.GET("/protected", func(c *gin.Context) {
		called = true
		require.Equal(t, uint(42), c.GetUint("user_id"))
		require.Equal(t, "alice", c.GetString("username"))
		c.Status(http.StatusOK)
	})

	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer goodtoken")
	w := httptest.NewRecorder()

	engine.ServeHTTP(w, req)

	require.True(t, called)
	require.Equal(t, http.StatusOK, w.Code)
}

func errsJSONUnmarshal(data []byte, target interface{}) error {
	return json.Unmarshal(data, target)
}
