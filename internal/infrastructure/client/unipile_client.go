package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"unipile-connector/internal/domain/service"
)

// UnipileClientImpl handles communication with Unipile API
type UnipileClientImpl struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewUnipileClient creates a new Unipile client
func NewUnipileClient(baseURL, apiKey string) service.UnipileClient {
	return &UnipileClientImpl{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ListAccounts lists all accounts from Unipile API
func (c *UnipileClientImpl) ListAccounts() (*service.AccountListResponse, error) {
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

	var response service.AccountListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// TestConnection tests the connection to Unipile API by calling ListAccounts
func (c *UnipileClientImpl) TestConnection() error {
	_, err := c.ListAccounts()
	if err != nil {
		return fmt.Errorf("unipile API health check failed: %w", err)
	}
	return nil
}

// GetAccount gets the status of a LinkedIn account
func (c *UnipileClientImpl) GetAccount(accountID string) (*service.Account, error) {
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

	switch resp.StatusCode {
	case http.StatusOK:
		var response service.Account
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return &response, nil
	case http.StatusNotFound:
		return nil, service.ErrUnipileAccountNotFound
	default:
		return nil, fmt.Errorf("unipile API error (status %d): %s", resp.StatusCode, string(body))
	}
}

// GetAccountWithLongPolling gets the status of a LinkedIn account with long polling
// This method will wait until the account status changes to "OK" or timeout occurs
func (c *UnipileClientImpl) GetAccountWithLongPolling(accountID string, timeout time.Duration) (*service.Account, error) {
	url := fmt.Sprintf("%s/api/v1/accounts/%s", c.baseURL, accountID)

	// Create a client with longer timeout for long polling
	longPollClient := &http.Client{
		Timeout: timeout,
	}

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("X-API-KEY", c.apiKey)
	httpReq.Header.Set("accept", "application/json")

	resp, err := longPollClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		var response service.Account
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return &response, nil
	case http.StatusNotFound:
		return nil, service.ErrUnipileAccountNotFound
	default:
		return nil, fmt.Errorf("unipile API error (status %d): %s", resp.StatusCode, string(body))
	}
}

// DeleteAccount deletes an account from Unipile API
func (c *UnipileClientImpl) DeleteAccount(accountID string) error {
	url := fmt.Sprintf("%s/api/v1/accounts/%s", c.baseURL, accountID)

	httpReq, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("X-API-KEY", c.apiKey)
	httpReq.Header.Set("accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	switch resp.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusNotFound:
		return service.ErrUnipileAccountNotFound
	default:
		return fmt.Errorf("unipile API error (status %d): %s", resp.StatusCode, string(body))
	}
}

// ConnectLinkedIn connects a LinkedIn account using Unipile
func (c *UnipileClientImpl) ConnectLinkedIn(req *service.ConnectLinkedInRequest) (*service.ConnectLinkedInResponse, error) {
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

	var response service.ConnectLinkedInResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	response.Status = resp.StatusCode
	response.RowBody = string(body)

	// Handle different response status codes
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		return &response, nil
	default:
		return nil, fmt.Errorf("unipile API error (status %d): %s", resp.StatusCode, string(body))
	}
}

// SolveCheckpoint solves a LinkedIn authentication checkpoint
func (c *UnipileClientImpl) SolveCheckpoint(req *service.SolveCheckpointRequest) (*service.SolveCheckpointResponse, error) {
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

	var response service.SolveCheckpointResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// Handle different response status codes
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		return &response, nil
	case http.StatusUnauthorized:
		return nil, service.ErrUnipileInvalidCodeOrExpiredCheckpoint
	default:
		if response.Type == "errors/authentication_intent_error" {
			return nil, service.ErrUnipileInvalidCodeOrExpiredCheckpoint
		}
		return nil, fmt.Errorf("unipile API error (status %d): %s", resp.StatusCode, string(body))
	}
}
