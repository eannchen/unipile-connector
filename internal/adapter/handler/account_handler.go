package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"unipile-connector/internal/usecase/account"
)

// AccountHandler handles account-related requests
type AccountHandler struct {
	accountUsecase *account.AccountUsecase
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(accountUsecase *account.AccountUsecase) *AccountHandler {
	return &AccountHandler{
		accountUsecase: accountUsecase,
	}
}

// ListUserAccounts retrieves all accounts for the current user
func (h *AccountHandler) ListUserAccounts(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, ok := userIDStr.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	accounts, err := h.accountUsecase.ListUserAccounts(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve accounts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts": accounts,
	})
}

// DisconnectLinkedInRequest represents request to disconnect a LinkedIn account
type DisconnectLinkedInRequest struct {
	AccountID string `json:"account_id" binding:"required"`
}

// DisconnectLinkedIn disconnects LinkedIn account for the current user
func (h *AccountHandler) DisconnectLinkedIn(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, ok := userIDStr.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	var req DisconnectLinkedInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	err := h.accountUsecase.DisconnectLinkedIn(c.Request.Context(), userID, req.AccountID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disconnect account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "LinkedIn account disconnected successfully",
	})
}

// ConnectLinkedInRequest represents LinkedIn connection request
type ConnectLinkedInRequest struct {
	Type        string `json:"type" binding:"required"` // "credentials" or "cookie"
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
}

// ConnectLinkedIn handles LinkedIn account connection
func (h *AccountHandler) ConnectLinkedIn(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, ok := userIDStr.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	var req ConnectLinkedInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	// Validate request based on type
	if req.Type == "credentials" {
		if req.Username == "" || req.Password == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password required for credentials type"})
			return
		}
	} else if req.Type == "cookie" {
		if req.AccessToken == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Access token required for cookie type"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type must be 'credentials' or 'cookie'"})
		return
	}

	connectReq := &account.ConnectLinkedInRequest{
		Username:    req.Username,
		Password:    req.Password,
		AccessToken: req.AccessToken,
		UserAgent:   req.UserAgent,
	}

	// Store account in database
	entityAccount, err := h.accountUsecase.ConnectLinkedInAccount(c.Request.Context(), userID, connectReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store account"})
		return
	}

	c.JSON(http.StatusOK, entityAccount)
	return
}

// SolveCheckpointRequest represents request to solve a checkpoint
type SolveCheckpointRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Code      string `json:"code" binding:"required"`
}

// SolveCheckpoint handles LinkedIn checkpoint solving
func (h *AccountHandler) SolveCheckpoint(c *gin.Context) {
	userIDStr, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userID, ok := userIDStr.(uint)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	var req SolveCheckpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	solveReq := &account.SolveCheckpointRequest{
		AccountID: req.AccountID,
		Code:      req.Code,
	}

	entityAccount, err := h.accountUsecase.SolveCheckpoint(c.Request.Context(), userID, solveReq)
	if err != nil {
		if errors.Is(err, account.ErrInvalidCodeOrExpiredCheckpoint) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"type":   "ErrInvalidCodeOrExpiredCheckpoint",
				"detail": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to solve checkpoint"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "LinkedIn account connected successfully",
		"account_id": entityAccount.AccountID,
		"account": gin.H{
			"id":         entityAccount.ID,
			"provider":   entityAccount.Provider,
			"account_id": entityAccount.AccountID,
			"created_at": entityAccount.CreatedAt,
		},
	})
}
