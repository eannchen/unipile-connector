package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"unipile-connector/internal/domain/errs"
	"unipile-connector/internal/usecase/user"
)

// AuthHandler handles authentication requests
type AuthHandler interface {
	Register(c *gin.Context)
	Login(c *gin.Context)
	Logout(c *gin.Context)
	RefreshToken(c *gin.Context)
	GetCurrentUser(c *gin.Context)
}

// AuthHandlerImpl handles authentication requests
type AuthHandlerImpl struct {
	userUsecase user.Usecase
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(userUsecase user.Usecase) AuthHandler {
	return &AuthHandlerImpl{
		userUsecase: userUsecase,
	}
}

// RegisterRequest represents user registration request
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest represents user login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandlerImpl) userIDFromContext(c *gin.Context) (uint, error) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		return 0, errs.ErrUserNotAuthenticated
	}

	userID, ok := userIDStr.(uint)
	if !ok {
		return 0, errs.ErrInvalidUserID
	}

	return userID, nil
}

// Register handles user registration
func (h *AuthHandlerImpl) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, errs.WrapValidationError(err, "Invalid request data"))
		return
	}

	user, err := h.userUsecase.CreateUser(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		RespondError(c, err)
		return
	}

	RespondSuccess(c, http.StatusCreated, "User created successfully", gin.H{
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
		},
	})
}

// Login handles user login
func (h *AuthHandlerImpl) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, errs.WrapValidationError(err, "Invalid request data"))
		return
	}

	user, token, err := h.userUsecase.AuthenticateUser(c.Request.Context(), req.Username, req.Password)
	if err != nil {
		RespondError(c, err)
		return
	}

	RespondSuccess(c, http.StatusOK, "Login successful", gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
		},
	})
}

// Logout handles user logout
func (h *AuthHandlerImpl) Logout(c *gin.Context) {
	// Get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		RespondError(c, errs.WrapValidationError(errors.New("authorization header required"), "Authorization header required"))
		return
	}

	// Extract token
	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	// Blacklist the token
	h.userUsecase.BlacklistToken(c.Request.Context(), tokenString)

	RespondSuccess(c, http.StatusOK, "Logout successful", nil)
}

// RefreshToken handles token refresh
func (h *AuthHandlerImpl) RefreshToken(c *gin.Context) {
	// Get current token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		RespondError(c, errs.WrapValidationError(errors.New("authorization header required"), "Authorization header required"))
		return
	}

	// Extract token
	tokenString := authHeader
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		tokenString = authHeader[7:]
	}

	// Refresh token
	newToken, err := h.userUsecase.RefreshToken(c.Request.Context(), tokenString)
	if err != nil {
		RespondError(c, err)
		return
	}

	RespondSuccess(c, http.StatusOK, "Token refreshed successfully", gin.H{
		"token": newToken,
	})
}

// GetCurrentUser returns current user info
func (h *AuthHandlerImpl) GetCurrentUser(c *gin.Context) {
	userID, err := h.userIDFromContext(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	user, err := h.userUsecase.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		RespondError(c, err)
		return
	}

	RespondSuccess(c, http.StatusOK, "User retrieved successfully", gin.H{
		"id":       user.ID,
		"username": user.Username,
	})
}
