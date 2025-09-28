package handler

import (
	"net/http"
	"unipile-connector/internal/infrastructure/client"
	"unipile-connector/internal/usecase/account"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// AccountHandler handles account-related requests
type AccountHandler struct {
	accountUsecase *account.AccountUsecase
	unipileClient  *client.UnipileClient
	logger         *logrus.Logger
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(accountUsecase *account.AccountUsecase, unipileClient *client.UnipileClient, logger *logrus.Logger) *AccountHandler {
	return &AccountHandler{
		accountUsecase: accountUsecase,
		unipileClient:  unipileClient,
		logger:         logger,
	}
}

// ConnectLinkedInRequest represents LinkedIn connection request
type ConnectLinkedInRequest struct {
	Type     string `json:"type" binding:"required"` // "credentials" or "cookie"
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Cookie   string `json:"cookie,omitempty"`
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
		h.logger.WithError(err).Error("Invalid LinkedIn connection request")
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
		if req.Cookie == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Cookie required for cookie type"})
			return
		}
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Type must be 'credentials' or 'cookie'"})
		return
	}

	// Call Unipile API
	unipileReq := &client.ConnectLinkedInRequest{
		Type:     req.Type,
		Username: req.Username,
		Password: req.Password,
		Cookie:   req.Cookie,
	}

	unipileResp, err := h.unipileClient.ConnectLinkedIn(unipileReq)
	if err != nil {
		h.logger.WithError(err).Error("Failed to connect LinkedIn via Unipile")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect LinkedIn account"})
		return
	}

	if !unipileResp.Success {
		c.JSON(http.StatusBadRequest, gin.H{"error": unipileResp.Message})
		return
	}

	// Store account in database
	account, err := h.accountUsecase.ConnectLinkedInAccount(c.Request.Context(), userID, unipileResp.AccountID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to store LinkedIn account")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "LinkedIn account connected successfully",
		"account_id": account.AccountID,
		"account": gin.H{
			"id":         account.ID,
			"provider":   account.Provider,
			"account_id": account.AccountID,
			"created_at": account.CreatedAt,
		},
	})
}

// GetUserAccounts retrieves all accounts for the current user
func (h *AccountHandler) GetUserAccounts(c *gin.Context) {
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

	accounts, err := h.accountUsecase.GetUserAccounts(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to get user accounts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve accounts"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"accounts": accounts,
	})
}

// GetLinkedInAccount retrieves LinkedIn account for the current user
func (h *AccountHandler) GetLinkedInAccount(c *gin.Context) {
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

	account, err := h.accountUsecase.GetLinkedInAccount(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "LinkedIn account not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"account": gin.H{
			"id":         account.ID,
			"provider":   account.Provider,
			"account_id": account.AccountID,
			"created_at": account.CreatedAt,
		},
	})
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

	err := h.accountUsecase.DisconnectLinkedInAccount(c.Request.Context(), userID)
	if err != nil {
		h.logger.WithError(err).Error("Failed to disconnect LinkedIn account")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disconnect account"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "LinkedIn account disconnected successfully",
	})
}
