package handler

import (
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

	err := h.accountUsecase.DisconnectLinkedIn(c.Request.Context(), userID)
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
	Type        string `json:"type" binding:"required"` // "credentials" or ""
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

	unipileReq := &account.ConnectLinkedInRequest{
		Username:    req.Username,
		Password:    req.Password,
		AccessToken: req.AccessToken,
		UserAgent:   req.UserAgent,
	}

	// Store account in database
	resp, err := h.accountUsecase.ConnectLinkedInAccount(c.Request.Context(), userID, unipileReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store account"})
		return
	}

	if resp.Success {
		c.JSON(http.StatusOK, gin.H{
			"message":    "LinkedIn account connected successfully",
			"account_id": resp.Account.AccountID,
			"account": gin.H{
				"id":         resp.Account.ID,
				"provider":   resp.Account.Provider,
				"account_id": resp.Account.AccountID,
				"created_at": resp.Account.CreatedAt,
			},
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "Checkpoint required",
		"account_id": resp.Account.AccountID,
		"checkpoint": gin.H{
			"type": resp.Checkpoint.Type,
		},
		"expires_at":   resp.ExpiresAt,
		"row_response": resp.RowResponse,
	})
}

// SolveCheckpointRequest represents request to solve a checkpoint
type SolveCheckpointRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Code      string `json:"code" binding:"required"`
}

// SolveCheckpoint handles LinkedIn checkpoint solving
func (h *AccountHandler) SolveCheckpoint(c *gin.Context) {
	// userIDStr, exists := c.Get("user_id")
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
	// 	return
	// }

	// userID, ok := userIDStr.(uint)
	// if !ok {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
	// 	return
	// }

	// var req SolveCheckpointRequest
	// if err := c.ShouldBindJSON(&req); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
	// 	return
	// }

	// // Call Unipile API to solve checkpoint
	// unipileReq := &client.SolveCheckpointRequest{
	// 	Provider:  "LINKEDIN",
	// 	AccountID: req.AccountID,
	// 	Code:      req.Code,
	// }

	// unipileResp, err := h.unipileClient.SolveCheckpoint(unipileReq)
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to solve checkpoint"})
	// 	return
	// }

	// // Check if another checkpoint is required
	// if unipileResp.Checkpoint != nil {
	// 	c.JSON(http.StatusAccepted, gin.H{
	// 		"message":      "Another checkpoint required",
	// 		"account_id":   unipileResp.AccountID,
	// 		"checkpoint":   unipileResp.Checkpoint,
	// 		"requires_2fa": unipileResp.Checkpoint.Type == "2FA" || unipileResp.Checkpoint.Type == "OTP",
	// 	})
	// 	return
	// }

	// Store account in database
	// account, err := h.accountUsecase.ConnectLinkedInAccount(c.Request.Context(), userID, "")
	// if err != nil {
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store account"})
	// 	return
	// }

	// c.JSON(http.StatusOK, gin.H{
	// 	"message":    "LinkedIn account connected successfully",
	// 	"account_id": account.AccountID,
	// 	"account": gin.H{
	// 		"id":         account.ID,
	// 		"provider":   account.Provider,
	// 		"account_id": account.AccountID,
	// 		"created_at": account.CreatedAt,
	// 	},
	// })
}
