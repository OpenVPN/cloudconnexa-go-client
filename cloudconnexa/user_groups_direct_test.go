package cloudconnexa

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// createTestUserGroupsClient creates a test client with the given server for user groups testing
func createTestUserGroupsClient(server *httptest.Server) *Client {
	client := &Client{
		client:      server.Client(),
		BaseURL:     server.URL,
		Token:       "test-token",
		RateLimiter: rate.NewLimiter(rate.Every(1*time.Second), 5),
	}
	client.UserGroups = (*UserGroupsService)(&service{client: client})
	return client
}

func TestUserGroupsService_GetByID(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method and path
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/user-groups/group-123" {
			t.Errorf("Expected path /api/v1/user-groups/group-123, got %s", r.URL.Path)
		}

		// Mock response
		userGroup := UserGroup{
			ID:                 "group-123",
			Name:               "Test Group",
			ConnectAuth:        "LOCAL",
			InternetAccess:     "BLOCKED",
			MaxDevice:          5,
			SystemSubnets:      []string{"10.0.0.0/8", "192.168.0.0/16"},
			VpnRegionIDs:       []string{"region-1", "region-2"},
			AllRegionsIncluded: false,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(userGroup)
	}))
	defer server.Close()

	// Create client with mock server
	client := createTestUserGroupsClient(server)

	// Test the GetByID method
	result, err := client.UserGroups.GetByID("group-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.ID != "group-123" {
		t.Errorf("Expected group ID 'group-123', got %s", result.ID)
	}

	if result.Name != "Test Group" {
		t.Errorf("Expected group name 'Test Group', got %s", result.Name)
	}

	if result.MaxDevice != 5 {
		t.Errorf("Expected max device 5, got %d", result.MaxDevice)
	}

	if len(result.SystemSubnets) != 2 {
		t.Errorf("Expected 2 system subnets, got %d", len(result.SystemSubnets))
	}

	if len(result.VpnRegionIDs) != 2 {
		t.Errorf("Expected 2 VPN region IDs, got %d", len(result.VpnRegionIDs))
	}

	if result.AllRegionsIncluded {
		t.Error("Expected AllRegionsIncluded to be false")
	}
}

func TestUserGroupsService_GetByID_NotFound(t *testing.T) {
	// Create a mock server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "User group not found"}`))
	}))
	defer server.Close()

	// Create client with mock server
	client := createTestUserGroupsClient(server)

	// Test the GetByID method with non-existent group
	_, err := client.UserGroups.GetByID("non-existent")
	if err == nil {
		t.Error("Expected error for non-existent group, got nil")
	}
}

func TestUserGroupsService_GetByID_vs_Get(t *testing.T) {
	// Test that GetByID is more efficient than Get
	// This test demonstrates the difference between direct API call and pagination search

	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++

		switch r.URL.Path {
		case "/api/v1/user-groups/group-123":
			// Direct endpoint - should be called only once
			userGroup := UserGroup{
				ID:   "group-123",
				Name: "Test Group",
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(userGroup)
		case "/api/v1/user-groups":
			// Pagination endpoint - may be called multiple times
			response := UserGroupPageResponse{
				Content: []UserGroup{
					{ID: "group-123", Name: "Test Group"},
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

	client := createTestUserGroupsClient(server)

	// Test GetByID (direct endpoint)
	callCount = 0
	_, err := client.UserGroups.GetByID("group-123")
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	directCalls := callCount

	// Test Get (pagination search)
	callCount = 0
	_, err = client.UserGroups.Get("group-123")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	paginationCalls := callCount

	// GetByID should make fewer API calls
	if directCalls >= paginationCalls {
		t.Errorf("Expected GetByID to make fewer calls than Get. GetByID: %d, Get: %d", directCalls, paginationCalls)
	}

	t.Logf("GetByID made %d API calls, Get made %d API calls", directCalls, paginationCalls)
}

func TestUserGroupsService_GetByID_CompleteFields(t *testing.T) {
	// Test that GetByID returns all expected fields
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		userGroup := UserGroup{
			ID:                 "group-456",
			Name:               "Complete Test Group",
			ConnectAuth:        "EVERY_TIME",
			InternetAccess:     "GLOBAL_INTERNET",
			MaxDevice:          10,
			SystemSubnets:      []string{"172.16.0.0/12"},
			VpnRegionIDs:       []string{"us-east-1", "eu-west-1", "ap-southeast-1"},
			AllRegionsIncluded: true,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(userGroup)
	}))
	defer server.Close()

	client := createTestUserGroupsClient(server)

	result, err := client.UserGroups.GetByID("group-456")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify all fields are properly populated
	if result.ConnectAuth != "EVERY_TIME" {
		t.Errorf("Expected ConnectAuth 'EVERY_TIME', got %s", result.ConnectAuth)
	}

	if result.InternetAccess != "GLOBAL_INTERNET" {
		t.Errorf("Expected InternetAccess 'GLOBAL_INTERNET', got %s", result.InternetAccess)
	}

	if result.MaxDevice != 10 {
		t.Errorf("Expected MaxDevice 10, got %d", result.MaxDevice)
	}

	if !result.AllRegionsIncluded {
		t.Error("Expected AllRegionsIncluded to be true")
	}

	if len(result.VpnRegionIDs) != 3 {
		t.Errorf("Expected 3 VPN regions, got %d", len(result.VpnRegionIDs))
	}
}
