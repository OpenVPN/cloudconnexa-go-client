package cloudconnexa

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// createTestClient creates a test client with the given server
func createTestClient(server *httptest.Server) *Client {
	client := &Client{
		client:            server.Client(),
		BaseURL:           server.URL,
		Token:             "test-token",
		ReadRateLimiter:   rate.NewLimiter(rate.Every(1), 5),
		UpdateRateLimiter: rate.NewLimiter(rate.Every(1), 5),
	}
	client.Devices = (*DevicesService)(&service{client: client})
	client.Sessions = (*SessionsService)(&service{client: client})
	client.NetworkConnectors = (*NetworkConnectorsService)(&service{client: client})
	return client
}

func TestDevicesService_List(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method and path
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/devices" {
			t.Errorf("Expected path /api/v1/devices, got %s", r.URL.Path)
		}

		// Mock response
		response := DevicePageResponse{
			Content: []DeviceDetail{
				{
					ID:     "device-1",
					Name:   "Test Device 1",
					UserID: "user-1",
					Status: "ACTIVE",
				},
				{
					ID:     "device-2",
					Name:   "Test Device 2",
					UserID: "user-2",
					Status: "INACTIVE",
				},
			},
			NumberOfElements: 2,
			Page:             0,
			Size:             10,
			Success:          true,
			TotalElements:    2,
			TotalPages:       1,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with mock server
	client := createTestClient(server)

	// Test the List method
	options := DeviceListOptions{
		Page: 0,
		Size: 10,
	}
	result, err := client.Devices.List(options)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Content) != 2 {
		t.Errorf("Expected 2 devices, got %d", len(result.Content))
	}

	if result.Content[0].ID != "device-1" {
		t.Errorf("Expected device ID 'device-1', got %s", result.Content[0].ID)
	}

	if result.TotalElements != 2 {
		t.Errorf("Expected total elements 2, got %d", result.TotalElements)
	}
}

func TestDevicesService_GetByID(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method and path
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/devices/device-123" {
			t.Errorf("Expected path /api/v1/devices/device-123, got %s", r.URL.Path)
		}

		// Mock response
		device := DeviceDetail{
			ID:        "device-123",
			Name:      "Test Device",
			UserID:    "user-123",
			Status:    "ACTIVE",
			Type:      "CLIENT",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(device)
	}))
	defer server.Close()

	// Create client with mock server
	client := createTestClient(server)

	// Test the GetByID method
	result, err := client.Devices.GetByID("device-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.ID != "device-123" {
		t.Errorf("Expected device ID 'device-123', got %s", result.ID)
	}

	if result.Name != "Test Device" {
		t.Errorf("Expected device name 'Test Device', got %s", result.Name)
	}

	if result.Status != "ACTIVE" {
		t.Errorf("Expected device status 'ACTIVE', got %s", result.Status)
	}
}

func TestDevicesService_Update(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method and path
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/devices/device-123" {
			t.Errorf("Expected path /api/v1/devices/device-123, got %s", r.URL.Path)
		}

		// Mock response
		device := DeviceDetail{
			ID:          "device-123",
			Name:        "Updated Device Name",
			Description: "Updated description",
			UserID:      "user-123",
			Status:      "ACTIVE",
			Type:        "CLIENT",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(device)
	}))
	defer server.Close()

	// Create client with mock server
	client := createTestClient(server)

	// Test the Update method
	updateRequest := DeviceUpdateRequest{
		Name:        "Updated Device Name",
		Description: "Updated description",
	}
	result, err := client.Devices.Update("device-123", updateRequest)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Name != "Updated Device Name" {
		t.Errorf("Expected device name 'Updated Device Name', got %s", result.Name)
	}

	if result.Description != "Updated description" {
		t.Errorf("Expected device description 'Updated description', got %s", result.Description)
	}
}

func TestDevicesService_Block(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Mock response
		device := DeviceDetail{
			ID:     "device-123",
			Name:   "Test Device",
			UserID: "user-123",
			Status: "BLOCKED",
			Type:   "CLIENT",
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(device)
	}))
	defer server.Close()

	// Create client with mock server
	client := createTestClient(server)

	// Test the Block method
	result, err := client.Devices.Block("device-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.Status != "BLOCKED" {
		t.Errorf("Expected device status 'BLOCKED', got %s", result.Status)
	}
}

func TestDevicesService_ListByUserID(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check query parameters
		query := r.URL.Query()
		if query.Get("userId") != "user-123" {
			t.Errorf("Expected userId=user-123, got %s", query.Get("userId"))
		}

		// Mock response
		response := DevicePageResponse{
			Content: []DeviceDetail{
				{
					ID:     "device-1",
					Name:   "User Device 1",
					UserID: "user-123",
					Status: "ACTIVE",
				},
				{
					ID:     "device-2",
					Name:   "User Device 2",
					UserID: "user-123",
					Status: "INACTIVE",
				},
			},
			NumberOfElements: 2,
			Page:             0,
			Size:             100,
			Success:          true,
			TotalElements:    2,
			TotalPages:       1,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Create client with mock server
	client := createTestClient(server)

	// Test the ListByUserID method
	result, err := client.Devices.ListByUserID("user-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result) != 2 {
		t.Errorf("Expected 2 devices, got %d", len(result))
	}

	for _, device := range result {
		if device.UserID != "user-123" {
			t.Errorf("Expected device userID 'user-123', got %s", device.UserID)
		}
	}
}

func TestDevicesService_List_InvalidSize(t *testing.T) {
	client := &Client{
		ReadRateLimiter:   rate.NewLimiter(rate.Every(1), 5),
		UpdateRateLimiter: rate.NewLimiter(rate.Every(1), 5),
	}
	client.Devices = (*DevicesService)(&service{client: client})

	// Test with size too large
	options := DeviceListOptions{
		Size: 1001,
	}
	_, err := client.Devices.List(options)
	if err == nil {
		t.Error("Expected error for size 1001, got nil")
	}
}
