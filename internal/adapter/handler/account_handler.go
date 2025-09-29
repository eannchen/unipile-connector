package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"unipile-connector/internal/domain/errs"
	"unipile-connector/internal/usecase/account"
)

// AccountHandler handles account-related requests
type AccountHandler interface {
	ListUserAccounts(c *gin.Context)
	DisconnectLinkedIn(c *gin.Context)
	ConnectLinkedIn(c *gin.Context)
	SolveCheckpoint(c *gin.Context)
}

// AccountHandlerImpl handles account-related requests
type AccountHandlerImpl struct {
	accountUsecase account.AccountUsecase
}

// NewAccountHandler creates a new account handler
func NewAccountHandler(accountUsecase account.AccountUsecase) AccountHandler {
	return &AccountHandlerImpl{
		accountUsecase: accountUsecase,
	}
}

func (h *AccountHandlerImpl) userIDFromContext(c *gin.Context) (uint, error) {
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

// ListUserAccounts retrieves all accounts for the current user
func (h *AccountHandlerImpl) ListUserAccounts(c *gin.Context) {
	userID, err := h.userIDFromContext(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	accounts, err := h.accountUsecase.ListUserAccounts(c.Request.Context(), userID)
	if err != nil {
		RespondError(c, err)
		return
	}

	RespondSuccess(c, http.StatusOK, "Accounts retrieved successfully", gin.H{
		"accounts": accounts,
	})
}

// DisconnectLinkedInRequest represents request to disconnect a LinkedIn account
type DisconnectLinkedInRequest struct {
	AccountID string `json:"account_id" binding:"required"`
}

// DisconnectLinkedIn disconnects LinkedIn account for the current user
func (h *AccountHandlerImpl) DisconnectLinkedIn(c *gin.Context) {
	userID, err := h.userIDFromContext(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	var req DisconnectLinkedInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, errs.WrapValidationError(err, "Invalid request data"))
		return
	}

	err = h.accountUsecase.DisconnectLinkedIn(c.Request.Context(), userID, req.AccountID)
	if err != nil {
		RespondError(c, err)
		return
	}

	RespondSuccess(c, http.StatusOK, "LinkedIn account disconnected successfully", nil)
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
func (h *AccountHandlerImpl) ConnectLinkedIn(c *gin.Context) {
	userID, err := h.userIDFromContext(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	var req ConnectLinkedInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, errs.WrapValidationError(err, "Invalid request data"))
		return
	}

	// Validate request based on type
	if req.Type == "credentials" {
		if req.Username == "" || req.Password == "" {
			RespondError(c, errs.WrapValidationError(errors.New("username and password required for credentials type"), "Username and password required for credentials type"))
			return
		}
	} else if req.Type == "cookie" {
		if req.AccessToken == "" {
			RespondError(c, errs.WrapValidationError(errors.New("access token required for cookie type"), "Access token required for cookie type"))
			return
		}
	} else {
		RespondError(c, errs.WrapValidationError(errors.New("type must be 'credentials' or 'cookie"), "Type must be 'credentials' or 'cookie'"))
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
		RespondError(c, err)
		return
	}

	RespondSuccess(c, http.StatusOK, "LinkedIn account connected successfully", gin.H{
		"account": entityAccount,
	})
}

// SolveCheckpointRequest represents request to solve a checkpoint
type SolveCheckpointRequest struct {
	AccountID string `json:"account_id" binding:"required"`
	Code      string `json:"code" binding:"required"`
}

// SolveCheckpoint handles LinkedIn checkpoint solving
func (h *AccountHandlerImpl) SolveCheckpoint(c *gin.Context) {
	userID, err := h.userIDFromContext(c)
	if err != nil {
		RespondError(c, err)
		return
	}

	var req SolveCheckpointRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondError(c, errs.WrapValidationError(err, "Invalid request data"))
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
				"type":  "ErrInvalidCodeOrExpiredCheckpoint",
				"error": err.Error(),
			})
			return
		}
		RespondError(c, err)
		return
	}

	RespondSuccess(c, http.StatusOK, "LinkedIn checkpoint solved successfully", gin.H{
		"account": entityAccount,
	})
}
