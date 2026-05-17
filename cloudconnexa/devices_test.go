package cloudconnexa

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
				},
				{
					ID:     "device-2",
					Name:   "Test Device 2",
					UserID: "user-2",
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

	// Test the List method
	options := DeviceListOptions{
		Page: 0,
		Size: 100,
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
		if got := r.URL.Query().Get("userId"); got != "user-123" {
			t.Errorf("Expected userId=user-123, got %s", got)
		}

		// Mock response
		device := DeviceDetail{
			ID:     "device-123",
			Name:   "Test Device",
			UserID: "user-123",
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(device)
	}))
	defer server.Close()

	// Create client with mock server
	client := createTestClient(server)

	// Test the GetByID method
	result, err := client.Devices.GetByID("user-123", "device-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if result.ID != "device-123" {
		t.Errorf("Expected device ID 'device-123', got %s", result.ID)
	}

	if result.Name != "Test Device" {
		t.Errorf("Expected device name 'Test Device', got %s", result.Name)
	}
}

func TestDevicesService_Update(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method and path
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/devices/device-123" {
			t.Errorf("Expected path /api/v1/devices/device-123, got %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("userId"); got != "user-123" {
			t.Errorf("Expected userId=user-123, got %s", got)
		}

		// Mock response
		device := DeviceDetail{
			ID:          "device-123",
			Name:        "Updated Device Name",
			Description: "Updated description",
			UserID:      "user-123",
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
	result, err := client.Devices.Update("user-123", "device-123", updateRequest)
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
				},
				{
					ID:     "device-2",
					Name:   "User Device 2",
					UserID: "user-123",
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

func TestDevicesService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/devices" {
			t.Errorf("Expected path /api/v1/devices, got %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("userId"); got != "user-123" {
			t.Errorf("Expected userId=user-123, got %s", got)
		}

		var req DeviceCreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("Failed to decode request body: %v", err)
		}
		if req.Name != "New Device" || req.ClientUUID != "uuid-1" {
			t.Errorf("Unexpected request body: %+v", req)
		}

		device := DeviceDetail{
			ID:         "device-new",
			Name:       req.Name,
			UserID:     "user-123",
			ClientUUID: req.ClientUUID,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(device)
	}))
	defer server.Close()

	client := createTestClient(server)

	device, err := client.Devices.Create("user-123", DeviceCreateRequest{
		Name:        "New Device",
		Description: "for testing",
		ClientUUID:  "uuid-1",
	})
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if device.ID != "device-new" {
		t.Errorf("Expected device ID 'device-new', got %s", device.ID)
	}
	if device.ClientUUID != "uuid-1" {
		t.Errorf("Expected ClientUUID 'uuid-1', got %s", device.ClientUUID)
	}
}

func TestDevicesService_Create_EmptyUserID(t *testing.T) {
	client := createTestClient(httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("Expected no HTTP call when userID is empty")
	})))

	_, err := client.Devices.Create("", DeviceCreateRequest{Name: "x"})
	if err == nil {
		t.Error("Expected error for empty userID, got nil")
	}
}

func TestDevicesService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/devices/device-123" {
			t.Errorf("Expected path /api/v1/devices/device-123, got %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("userId"); got != "user-123" {
			t.Errorf("Expected userId=user-123, got %s", got)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := createTestClient(server)

	if err := client.Devices.Delete("user-123", "device-123"); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestDevicesService_Delete_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error":"device not found"}`))
	}))
	defer server.Close()

	client := createTestClient(server)

	if err := client.Devices.Delete("user-123", "nonexistent"); err == nil {
		t.Error("Expected error for missing device, got nil")
	}
}

func TestDevicesService_GenerateProfile(t *testing.T) {
	const ovpnProfile = "client\nproto udp\nremote example 1194\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/devices/device-123/profile" {
			t.Errorf("Expected path /api/v1/devices/device-123/profile, got %s", r.URL.Path)
		}
		q := r.URL.Query()
		if q.Get("userId") != "user-123" {
			t.Errorf("Expected userId=user-123, got %s", q.Get("userId"))
		}
		if q.Get("regionId") != "region-eu" {
			t.Errorf("Expected regionId=region-eu, got %s", q.Get("regionId"))
		}

		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte(ovpnProfile))
	}))
	defer server.Close()

	client := createTestClient(server)

	profile, err := client.Devices.GenerateProfile("user-123", "device-123", "region-eu")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if profile != ovpnProfile {
		t.Errorf("Expected profile body %q, got %q", ovpnProfile, profile)
	}
}

func TestDevicesService_GenerateProfile_EmptyRegionID(t *testing.T) {
	client := createTestClient(httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {
		t.Error("Expected no HTTP call when regionID is empty")
	})))

	_, err := client.Devices.GenerateProfile("user-123", "device-123", "")
	if err == nil {
		t.Error("Expected error for empty regionID, got nil")
	}
}

func TestDevicesService_RevokeProfile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/devices/device-123/profile" {
			t.Errorf("Expected path /api/v1/devices/device-123/profile, got %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("userId"); got != "user-123" {
			t.Errorf("Expected userId=user-123, got %s", got)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := createTestClient(server)

	if err := client.Devices.RevokeProfile("user-123", "device-123"); err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestDeviceDetail_NewFieldsRoundTrip(t *testing.T) {
	original := DeviceDetail{
		ID:               "device-1",
		Name:             "Device 1",
		UserID:           "user-1",
		ClientUUID:       "uuid-1",
		IPV4Address:      "10.0.0.5",
		IPV6Address:      "fd00::5",
		ConnectionStatus: "ONLINE",
	}

	encoded, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded DeviceDetail
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded != original {
		t.Errorf("Round-trip mismatch.\nWant: %+v\nGot:  %+v", original, decoded)
	}

	// Ensure the JSON tag spellings match the schema (clientUUID, ipV4Address, etc.).
	var raw map[string]any
	if err := json.Unmarshal(encoded, &raw); err != nil {
		t.Fatalf("Unmarshal to map failed: %v", err)
	}
	for _, key := range []string{"clientUUID", "ipV4Address", "ipV6Address", "connectionStatus"} {
		if _, ok := raw[key]; !ok {
			t.Errorf("Expected JSON key %q to be present", key)
		}
	}
}
