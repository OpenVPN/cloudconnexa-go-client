package cloudconnexa

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/time/rate"
)

// createTestClientWithNetworkConnectors creates a test client with NetworkConnectors service
func createTestClientWithNetworkConnectors(server *httptest.Server) *Client {
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

func TestNetworkConnectorsService_Activate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		// Check the path
		if r.URL.Path != "/api/v1/networks/connectors/connector-123/activate" {
			t.Errorf("Expected path /api/v1/networks/connectors/connector-123/activate, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestClientWithNetworkConnectors(server)

	err := client.NetworkConnectors.Activate("connector-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestNetworkConnectorsService_Suspend(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		// Check the path
		if r.URL.Path != "/api/v1/networks/connectors/connector-456/suspend" {
			t.Errorf("Expected path /api/v1/networks/connectors/connector-456/suspend, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestClientWithNetworkConnectors(server)

	err := client.NetworkConnectors.Suspend("connector-456")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestNetworkConnectorsService_Activate_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "connector not found"}`))
	}))
	defer server.Close()

	client := createTestClientWithNetworkConnectors(server)

	err := client.NetworkConnectors.Activate("nonexistent-connector")
	if err == nil {
		t.Error("Expected error for non-existent connector, got nil")
	}
}

func TestNetworkConnectorsService_Suspend_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "connector already suspended"}`))
	}))
	defer server.Close()

	client := createTestClientWithNetworkConnectors(server)

	err := client.NetworkConnectors.Suspend("already-suspended-connector")
	if err == nil {
		t.Error("Expected error for already suspended connector, got nil")
	}
}

func TestNetworkConnector_LicensedField(t *testing.T) {
	connector := NetworkConnector{
		ID:       "connector-123",
		Name:     "test-connector",
		Licensed: true,
	}

	if !connector.Licensed {
		t.Error("Expected Licensed to be true")
	}

	connector.Licensed = false
	if connector.Licensed {
		t.Error("Expected Licensed to be false")
	}
}

func TestIPSecConfig_NewFields(t *testing.T) {
	config := IPSecConfig{
		Platform:       "AWS",
		ConnectorState: "ACTIVE",
		ServerID:       "server-123",
		ServerIP:       "203.0.113.10",
	}

	if config.ConnectorState != "ACTIVE" {
		t.Errorf("Expected ConnectorState 'ACTIVE', got %s", config.ConnectorState)
	}

	if config.ServerID != "server-123" {
		t.Errorf("Expected ServerID 'server-123', got %s", config.ServerID)
	}

	if config.ServerIP != "203.0.113.10" {
		t.Errorf("Expected ServerIP '203.0.113.10', got %s", config.ServerIP)
	}
}
