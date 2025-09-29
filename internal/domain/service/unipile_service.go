package service

import "errors"

// UnipileClient handles communication with Unipile API
type UnipileClient interface {
	ListAccounts() (*AccountListResponse, error)
	TestConnection() error
	GetAccount(accountID string) (*Account, error)
	DeleteAccount(accountID string) error
	ConnectLinkedIn(req *ConnectLinkedInRequest) (*ConnectLinkedInResponse, error)
	SolveCheckpoint(req *SolveCheckpointRequest) (*SolveCheckpointResponse, error)
}

// Account represents a single account in the list
type Account struct {
	Object           string                 `json:"object"`
	ConnectionParams map[string]interface{} `json:"connection_params"`
	Name             string                 `json:"name"`
	Type             string                 `json:"type"`
	CreatedAt        string                 `json:"created_at"`
	Sources          []AccountSource        `json:"sources"`
	ID               string                 `json:"id"`
	Groups           []string               `json:"groups"`
}

// AccountSource represents a source within an account
type AccountSource struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// AccountListResponse represents the response from listing accounts
type AccountListResponse struct {
	Object string    `json:"object"`
	Items  []Account `json:"items"`
	Cursor *string   `json:"cursor"`
}

// ConnectLinkedInRequest represents the request to connect LinkedIn account
type ConnectLinkedInRequest struct {
	Provider    string `json:"provider"` // "LINKEDIN"
	Username    string `json:"username,omitempty"`
	Password    string `json:"password,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
	UserAgent   string `json:"user_agent,omitempty"`
}

// ConnectLinkedInResponse represents the response from LinkedIn connection
type ConnectLinkedInResponse struct {
	Object     string      `json:"object"`
	AccountID  string      `json:"account_id"`
	Checkpoint *Checkpoint `json:"checkpoint,omitempty"`
	Status     int         `json:"status"`
	RowBody    string      `json:"row_body,omitempty"`
}

// Checkpoint represents a LinkedIn authentication checkpoint
type Checkpoint struct {
	Type   string `json:"type"`   // "2FA", "OTP", "IN_APP_VALIDATION", "CAPTCHA", "PHONE_REGISTER"
	Source string `json:"source"` // "APP"
}

// SolveCheckpointRequest represents request to solve a checkpoint
type SolveCheckpointRequest struct {
	Provider  string `json:"provider"` // "LINKEDIN"
	AccountID string `json:"account_id"`
	Code      string `json:"code"`
}

// SolveCheckpointResponse represents response from checkpoint solving
type SolveCheckpointResponse struct {
	Object    string `json:"object"`
	AccountID string `json:"account_id"`
	// error response
	Status int    `json:"status"`
	Type   string `json:"type"`
	Title  string `json:"title"`
	Detail string `json:"detail"`
}

// ErrUnipileInvalidCodeOrExpiredCheckpoint is returned when the code is invalid or the checkpoint expired
var ErrUnipileInvalidCodeOrExpiredCheckpoint = errors.New("invalid code or expired checkpoint")

// ErrUnipileAccountNotFound is returned when an account is not found
var ErrUnipileAccountNotFound = errors.New("account not found")
