package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// UnipileClient handles communication with Unipile API
type UnipileClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewUnipileClient creates a new Unipile client
func NewUnipileClient(baseURL, apiKey string) *UnipileClient {
	return &UnipileClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
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
	Status     interface{} `json:"status,omitempty"` // Can be string or number
}

// Checkpoint represents a LinkedIn authentication checkpoint
type Checkpoint struct {
	Type string `json:"type"` // "2FA", "OTP", "IN_APP_VALIDATION", "CAPTCHA", "PHONE_REGISTER"
}

// SolveCheckpointRequest represents request to solve a checkpoint
type SolveCheckpointRequest struct {
	Provider  string `json:"provider"` // "LINKEDIN"
	AccountID string `json:"account_id"`
	Code      string `json:"code"`
}

// AccountStatusResponse represents account status response
type AccountStatusResponse struct {
	Object    string `json:"object"`
	AccountID string `json:"account_id"`
	Status    string `json:"status"` // "OK", "CHECKPOINT", "ERROR"
}

// AccountListResponse represents the response from listing accounts
type AccountListResponse struct {
	Object string    `json:"object"`
	Items  []Account `json:"items"`
	Cursor *string   `json:"cursor"`
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

// ListAccounts lists all accounts from Unipile API
func (c *UnipileClient) ListAccounts() (*AccountListResponse, error) {
	url := fmt.Sprintf("%s/api/v1/accounts", c.baseURL)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("X-API-KEY", c.apiKey)
	httpReq.Header.Set("accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unipile API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response AccountListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// TestConnection tests the connection to Unipile API by calling ListAccounts
func (c *UnipileClient) TestConnection() error {
	_, err := c.ListAccounts()
	if err != nil {
		return fmt.Errorf("unipile API health check failed: %w", err)
	}
	return nil
}

// ConnectLinkedIn connects a LinkedIn account using Unipile
func (c *UnipileClient) ConnectLinkedIn(req *ConnectLinkedInRequest) (*ConnectLinkedInResponse, error) {
	url := fmt.Sprintf("%s/api/v1/accounts", c.baseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-KEY", c.apiKey)
	httpReq.Header.Set("accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response ConnectLinkedInResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Handle different response status codes
	switch resp.StatusCode {
	case http.StatusOK:
		// Account connected successfully
		return &response, nil
	case http.StatusAccepted:
		// Checkpoint required
		return &response, nil
	default:
		return nil, fmt.Errorf("unipile API error (status %d): %s", resp.StatusCode, string(body))
	}
}

// SolveCheckpoint solves a LinkedIn authentication checkpoint
func (c *UnipileClient) SolveCheckpoint(req *SolveCheckpointRequest) (*ConnectLinkedInResponse, error) {
	url := fmt.Sprintf("%s/api/v1/accounts/checkpoint", c.baseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-KEY", c.apiKey)
	httpReq.Header.Set("accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var response ConnectLinkedInResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Handle different response status codes
	switch resp.StatusCode {
	case http.StatusOK:
		// Checkpoint solved successfully
		return &response, nil
	case http.StatusAccepted:
		// Another checkpoint required
		return &response, nil
	case http.StatusRequestTimeout:
		return nil, fmt.Errorf("checkpoint timeout: authentication intent expired")
	case http.StatusBadRequest:
		return nil, fmt.Errorf("invalid checkpoint: authentication intent expired")
	default:
		return nil, fmt.Errorf("unipile API error (status %d): %s", resp.StatusCode, string(body))
	}
}

// GetAccountStatus gets the status of a LinkedIn account
func (c *UnipileClient) GetAccountStatus(accountID string) (*AccountStatusResponse, error) {
	url := fmt.Sprintf("%s/api/v1/accounts/%s", c.baseURL, accountID)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("X-API-KEY", c.apiKey)
	httpReq.Header.Set("accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unipile API error (status %d): %s", resp.StatusCode, string(body))
	}

	var response AccountStatusResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}
