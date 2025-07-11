package cloudconnexa

import (
	"encoding/json"
	"golang.org/x/time/rate"
	"net/http"
	"net/http/httptest"
	"testing"
)

// createTestDNSClient creates a test client with the given server for DNS testing
func createTestDNSClient(server *httptest.Server) *Client {
	client := &Client{
		client:            server.Client(),
		BaseURL:           server.URL,
		Token:             "test-token",
		ReadRateLimiter:   rate.NewLimiter(rate.Every(1), 5),
		UpdateRateLimiter: rate.NewLimiter(rate.Every(1), 5),
	}
	client.DNSRecords = (*DNSRecordsService)(&service{client: client})
	return client
}

func TestDNSRecordsService_GetByID(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method and path
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/dns-records/record-123" {
			t.Errorf("Expected path /api/v1/dns-records/record-123, got %s", r.URL.Path)
		}

		// Mock response
		record := DNSRecord{
			ID:            "record-123",
			Domain:        "example.com",
			Description:   "Test DNS record",
			IPV4Addresses: []string{"192.168.1.1", "192.168.1.2"},
			IPV6Addresses: []string{"2001:db8::1"},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(record)
	}))
	defer server.Close()

	// Create client with mock server
	client := createTestDNSClient(server)

	// Test the GetByID method
	result, err := client.DNSRecords.GetByID("record-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.ID != "record-123" {
		t.Errorf("Expected record ID 'record-123', got %s", result.ID)
	}

	if result.Domain != "example.com" {
		t.Errorf("Expected domain 'example.com', got %s", result.Domain)
	}

	if len(result.IPV4Addresses) != 2 {
		t.Errorf("Expected 2 IPv4 addresses, got %d", len(result.IPV4Addresses))
	}

	if len(result.IPV6Addresses) != 1 {
		t.Errorf("Expected 1 IPv6 address, got %d", len(result.IPV6Addresses))
	}
}

func TestDNSRecordsService_GetByID_NotFound(t *testing.T) {
	// Create a mock server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "DNS record not found"}`))
	}))
	defer server.Close()

	// Create client with mock server
	client := createTestDNSClient(server)

	// Test the GetByID method with non-existent record
	_, err := client.DNSRecords.GetByID("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent record, got nil")
	}
}

func TestDNSRecordsService_GetByID_vs_GetDNSRecord(t *testing.T) {
	// Test that GetByID is more efficient than GetDNSRecord
	// This test demonstrates the difference between direct API call and pagination search

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		switch r.URL.Path {
		case "/api/v1/dns-records/record-123":
			// Direct endpoint - should be called only once
			record := DNSRecord{
				ID:     "record-123",
				Domain: "example.com",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(record)
		case "/api/v1/dns-records":
			// Pagination endpoint - may be called multiple times
			response := DNSRecordPageResponse{
				Content: []DNSRecord{
					{ID: "record-123", Domain: "example.com"},
				},
				Page:       0,
				Size:       10,
				TotalPages: 1,
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	client := createTestDNSClient(server)

	// Test GetByID (direct endpoint)
	callCount = 0
	_, err := client.DNSRecords.GetByID("record-123")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	directCalls := callCount

	// Test GetDNSRecord (pagination search)
	callCount = 0
	_, err = client.DNSRecords.GetDNSRecord("record-123")
	if err != nil {
		t.Fatalf("GetDNSRecord failed: %v", err)
	}
	paginationCalls := callCount

	// GetByID should make fewer or equal API calls (both should be 1 in this simple case)
	if directCalls > paginationCalls {
		t.Errorf("Expected GetByID to make fewer or equal calls than GetDNSRecord. GetByID: %d, GetDNSRecord: %d", directCalls, paginationCalls)
	}

	// Both should make exactly 1 call in this test case
	if directCalls != 1 {
		t.Errorf("Expected GetByID to make exactly 1 call, got %d", directCalls)
	}

	t.Logf("GetByID made %d API calls, GetDNSRecord made %d API calls", directCalls, paginationCalls)
}
