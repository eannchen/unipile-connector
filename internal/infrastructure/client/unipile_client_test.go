package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"unipile-connector/internal/domain/service"
)

func TestUnipileClient_ListAccounts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "GET", r.Method)
		require.Equal(t, "/api/v1/accounts", r.URL.Path)
		require.Equal(t, "test-key", r.Header.Get("X-API-KEY"))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(service.AccountListResponse{Object: "list"})
	}))
	t.Cleanup(server.Close)

	c := &UnipileClientImpl{baseURL: server.URL, apiKey: "test-key", httpClient: server.Client()}
	resp, err := c.ListAccounts()
	require.NoError(t, err)
	require.Equal(t, "list", resp.Object)
}

func TestUnipileClient_ListAccounts_ErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusForbidden)
	}))
	t.Cleanup(server.Close)

	c := &UnipileClientImpl{baseURL: server.URL, apiKey: "key", httpClient: server.Client()}
	_, err := c.ListAccounts()
	require.Error(t, err)
}

func TestUnipileClient_GetAccount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/accounts/123", r.URL.Path)
		_ = json.NewEncoder(w).Encode(service.Account{ID: "123"})
	}))
	t.Cleanup(server.Close)

	c := &UnipileClientImpl{baseURL: server.URL, apiKey: "key", httpClient: server.Client()}
	resp, err := c.GetAccount("123")
	require.NoError(t, err)
	require.Equal(t, "123", resp.ID)
}

func TestUnipileClient_DeleteAccount_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	c := &UnipileClientImpl{baseURL: server.URL, apiKey: "key", httpClient: server.Client()}
	err := c.DeleteAccount("123")
	require.ErrorIs(t, err, service.ErrUnipileAccountNotFound)
}

func TestUnipileClient_DeleteAccount_OtherError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "fail", http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	c := &UnipileClientImpl{baseURL: server.URL, apiKey: "key", httpClient: server.Client()}
	err := c.DeleteAccount("123")
	require.Error(t, err)
}

func TestUnipileClient_ConnectLinkedIn(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		body, _ := io.ReadAll(r.Body)
		require.Contains(t, string(body), "\"provider\":\"LINKEDIN\"")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(body)
	}))
	t.Cleanup(server.Close)

	c := &UnipileClientImpl{baseURL: server.URL, apiKey: "key", httpClient: server.Client()}
	resp, err := c.ConnectLinkedIn(&service.ConnectLinkedInRequest{Provider: "LINKEDIN"})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.Status)
}

func TestUnipileClient_SolveCheckpoint_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(service.SolveCheckpointResponse{Type: "errors/authentication_intent_error"})
	}))
	t.Cleanup(server.Close)

	c := &UnipileClientImpl{baseURL: server.URL, apiKey: "key", httpClient: server.Client()}
	_, err := c.SolveCheckpoint(&service.SolveCheckpointRequest{AccountID: "123"})
	require.ErrorIs(t, err, service.ErrUnipileInvalidCodeOrExpiredCheckpoint)
}
