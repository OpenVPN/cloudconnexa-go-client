package cloudconnexa

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/time/rate"
)

// createTestClientWithHostConnectors creates a test client with HostConnectors service
func createTestClientWithHostConnectors(server *httptest.Server) *Client {
	client := &Client{
		client:            server.Client(),
		BaseURL:           server.URL,
		Token:             "test-token",
		ReadRateLimiter:   rate.NewLimiter(rate.Every(1), 5),
		UpdateRateLimiter: rate.NewLimiter(rate.Every(1), 5),
	}
	client.HostConnectors = (*HostConnectorsService)(&service{client: client})
	return client
}

func TestHostConnectorsService_Activate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		// Check the path
		if r.URL.Path != "/api/v1/hosts/connectors/connector-123/activate" {
			t.Errorf("Expected path /api/v1/hosts/connectors/connector-123/activate, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestClientWithHostConnectors(server)

	err := client.HostConnectors.Activate("connector-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestHostConnectorsService_Suspend(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		// Check the path
		if r.URL.Path != "/api/v1/hosts/connectors/connector-456/suspend" {
			t.Errorf("Expected path /api/v1/hosts/connectors/connector-456/suspend, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestClientWithHostConnectors(server)

	err := client.HostConnectors.Suspend("connector-456")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestHostConnectorsService_Activate_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "connector not found"}`))
	}))
	defer server.Close()

	client := createTestClientWithHostConnectors(server)

	err := client.HostConnectors.Activate("nonexistent-connector")
	if err == nil {
		t.Error("Expected error for non-existent connector, got nil")
	}
}

func TestHostConnectorsService_Suspend_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "connector already suspended"}`))
	}))
	defer server.Close()

	client := createTestClientWithHostConnectors(server)

	err := client.HostConnectors.Suspend("already-suspended-connector")
	if err == nil {
		t.Error("Expected error for already suspended connector, got nil")
	}
}

func TestHostConnector_LicensedField(t *testing.T) {
	connector := HostConnector{
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
