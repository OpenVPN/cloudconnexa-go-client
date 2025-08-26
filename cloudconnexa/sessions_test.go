package cloudconnexa

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// createTestSessionsClient creates a test client with the given server for sessions testing
func createTestSessionsClient(server *httptest.Server) *Client {
	client := &Client{
		client:            server.Client(),
		BaseURL:           server.URL,
		Token:             "test-token",
		ReadRateLimiter:   rate.NewLimiter(rate.Every(1), 5),
		UpdateRateLimiter: rate.NewLimiter(rate.Every(1), 5),
	}
	client.Sessions = (*SessionsService)(&service{client: client})
	return client
}

func TestSessionsService_List(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method and path
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/sessions" {
			t.Errorf("Expected path /api/v1/sessions, got %s", r.URL.Path)
		}

		// Check query parameters
		query := r.URL.Query()
		if query.Get("size") != "100" {
			t.Errorf("Expected size=100, got %s", query.Get("size"))
		}

		// Mock response
		response := SessionsResponse{
			Sessions: []Session{
				{
					ID:     "session-1",
					UserID: "user-1",
					Status: "ACTIVE",
				},
				{
					ID:     "session-2",
					UserID: "user-2",
					Status: "COMPLETED",
				},
			},
			NextCursor: "next-cursor-123",
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with mock server
	client := createTestSessionsClient(server)

	// Test the List method
	options := SessionsListOptions{
		Size: 100,
	}
	result, err := client.Sessions.List(options)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Sessions) != 2 {
		t.Errorf("Expected 2 sessions, got %d", len(result.Sessions))
	}

	if result.Sessions[0].ID != "session-1" {
		t.Errorf("Expected session ID 'session-1', got %s", result.Sessions[0].ID)
	}

	if result.NextCursor != "next-cursor-123" {
		t.Errorf("Expected next cursor 'next-cursor-123', got %s", result.NextCursor)
	}
}

func TestSessionsService_ListActive(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check query parameters
		query := r.URL.Query()
		if query.Get("status") != "ACTIVE" {
			t.Errorf("Expected status=ACTIVE, got %s", query.Get("status"))
		}

		// Mock response
		response := SessionsResponse{
			Sessions: []Session{
				{
					ID:     "session-1",
					UserID: "user-1",
					Status: "ACTIVE",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with mock server
	client := createTestSessionsClient(server)

	// Test the ListActive method
	result, err := client.Sessions.ListActive(10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(result.Sessions))
	}

	if result.Sessions[0].Status != "ACTIVE" {
		t.Errorf("Expected session status 'ACTIVE', got %s", result.Sessions[0].Status)
	}
}

func TestSessionsService_ListByDateRange(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check query parameters
		query := r.URL.Query()
		if query.Get("startDate") == "" {
			t.Error("Expected startDate parameter")
		}
		if query.Get("endDate") == "" {
			t.Error("Expected endDate parameter")
		}

		// Mock response
		response := SessionsResponse{
			Sessions: []Session{
				{
					ID:        "session-1",
					UserID:    "user-1",
					Status:    "COMPLETED",
					StartTime: time.Now(),
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with mock server
	client := createTestSessionsClient(server)

	// Test the ListByDateRange method
	startDate := time.Now().AddDate(0, 0, -7) // 7 days ago
	endDate := time.Now()
	result, err := client.Sessions.ListByDateRange(startDate, endDate, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Sessions) != 1 {
		t.Errorf("Expected 1 session, got %d", len(result.Sessions))
	}
}

func TestSessionsService_List_InvalidSize(t *testing.T) {
	client := &Client{
		ReadRateLimiter:   rate.NewLimiter(rate.Every(1), 5),
		UpdateRateLimiter: rate.NewLimiter(rate.Every(1), 5),
	}
	client.Sessions = (*SessionsService)(&service{client: client})

	// Test with size too small
	options := SessionsListOptions{
		Size: 0,
	}
	_, err := client.Sessions.List(options)
	if err == nil {
		t.Error("Expected error for size 0, got nil")
	}

	// Test with size too large
	options.Size = 101
	_, err = client.Sessions.List(options)
	if err == nil {
		t.Error("Expected error for size 101, got nil")
	}
}
