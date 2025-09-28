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
	Type     string `json:"type"` // "credentials" or "cookie"
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Cookie   string `json:"cookie,omitempty"`
}

// ConnectLinkedInResponse represents the response from LinkedIn connection
type ConnectLinkedInResponse struct {
	AccountID string `json:"account_id"`
	Success   bool   `json:"success"`
	Message   string `json:"message,omitempty"`
}

// ConnectLinkedIn connects a LinkedIn account using Unipile
func (c *UnipileClient) ConnectLinkedIn(req *ConnectLinkedInRequest) (*ConnectLinkedInResponse, error) {
	url := fmt.Sprintf("%s/api/v1/linkedin/connect", c.baseURL)

	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

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
		return nil, fmt.Errorf("unipile API error: %s", string(body))
	}

	var response ConnectLinkedInResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response, nil
}

// TestConnection tests the connection to Unipile API
func (c *UnipileClient) TestConnection() error {
	url := fmt.Sprintf("%s/api/v1/health", c.baseURL)

	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unipile API health check failed with status: %d", resp.StatusCode)
	}

	return nil
}

