package cloudconnexa

import (
	"encoding/json"
	"golang.org/x/time/rate"
	"net/http"
	"net/http/httptest"
	"testing"
)

// createTestIPsecClient creates a test client with the given server for IPsec testing
func createTestIPsecClient(server *httptest.Server) *Client {
	client := &Client{
		client:            server.Client(),
		BaseURL:           server.URL,
		Token:             "test-token",
		ReadRateLimiter:   rate.NewLimiter(rate.Every(1), 5),
		UpdateRateLimiter: rate.NewLimiter(rate.Every(1), 5),
	}
	client.NetworkConnectors = (*NetworkConnectorsService)(&service{client: client})
	return client
}

func TestNetworkConnectorsService_StartIPsec(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method and path
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/networks/connectors/connector-123/ipsec/start" {
			t.Errorf("Expected path /api/v1/networks/connectors/connector-123/ipsec/start, got %s", r.URL.Path)
		}

		// Mock response
		response := IPsecStartResponse{
			Success: true,
			Message: "IPsec tunnel started successfully",
			Status:  "STARTING",
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with mock server
	client := createTestIPsecClient(server)

	// Test the StartIPsec method
	result, err := client.NetworkConnectors.StartIPsec("connector-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Message != "IPsec tunnel started successfully" {
		t.Errorf("Expected message 'IPsec tunnel started successfully', got %s", result.Message)
	}

	if result.Status != "STARTING" {
		t.Errorf("Expected status 'STARTING', got %s", result.Status)
	}
}

func TestNetworkConnectorsService_StopIPsec(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method and path
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/networks/connectors/connector-123/ipsec/stop" {
			t.Errorf("Expected path /api/v1/networks/connectors/connector-123/ipsec/stop, got %s", r.URL.Path)
		}

		// Mock response
		response := IPsecStopResponse{
			Success: true,
			Message: "IPsec tunnel stopped successfully",
			Status:  "STOPPED",
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with mock server
	client := createTestIPsecClient(server)

	// Test the StopIPsec method
	result, err := client.NetworkConnectors.StopIPsec("connector-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if !result.Success {
		t.Error("Expected success to be true")
	}

	if result.Message != "IPsec tunnel stopped successfully" {
		t.Errorf("Expected message 'IPsec tunnel stopped successfully', got %s", result.Message)
	}

	if result.Status != "STOPPED" {
		t.Errorf("Expected status 'STOPPED', got %s", result.Status)
	}
}

func TestNetworkConnectorsService_StartIPsec_Error(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Connector not found"}`))
	}))
	defer server.Close()

	// Create client with mock server
	client := createTestIPsecClient(server)

	// Test the StartIPsec method with error
	_, err := client.NetworkConnectors.StartIPsec("invalid-connector")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestNetworkConnectorsService_StopIPsec_Error(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "Connector not found"}`))
	}))
	defer server.Close()

	// Create client with mock server
	client := createTestIPsecClient(server)

	// Test the StopIPsec method with error
	_, err := client.NetworkConnectors.StopIPsec("invalid-connector")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
