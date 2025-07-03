package cloudconnexa

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

// createTestHostIPServicesClient creates a test client with the given server for host IP services testing
func createTestHostIPServicesClient(server *httptest.Server) *Client {
	client := &Client{
		client:      server.Client(),
		BaseURL:     server.URL,
		Token:       "test-token",
		RateLimiter: rate.NewLimiter(rate.Every(1*time.Second), 5),
	}
	client.HostIPServices = (*HostIPServicesService)(&service{client: client})
	return client
}

func TestHostIPServicesService_UpdatedDTO(t *testing.T) {
	// Test that the updated DTO structure works correctly without duplicate routing information
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method and path
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/hosts/ip-services/service-123" {
			t.Errorf("Expected path /api/v1/hosts/ip-services/service-123, got %s", r.URL.Path)
		}

		// Mock response with the updated DTO structure (no duplicate routes)
		service := HostIPServiceResponse{
			ID:              "service-123",
			Name:            "Test IP Service",
			Description:     "Test service description",
			NetworkItemType: "HOST",
			NetworkItemID:   "host-456",
			Type:            "CUSTOM",
			Config: &IPServiceConfig{
				ServiceTypes: []string{"HTTP", "HTTPS"},
				CustomServiceTypes: []*CustomIPServiceType{
					{
						Protocol: "TCP",
						Port: []Range{
							{LowerValue: 80, UpperValue: 80},
							{LowerValue: 443, UpperValue: 443},
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(service)
	}))
	defer server.Close()

	// Create client with mock server
	client := createTestHostIPServicesClient(server)

	// Test the Get method with updated DTO
	result, err := client.HostIPServices.Get("service-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the response structure
	if result.ID != "service-123" {
		t.Errorf("Expected service ID 'service-123', got %s", result.ID)
	}

	if result.Name != "Test IP Service" {
		t.Errorf("Expected service name 'Test IP Service', got %s", result.Name)
	}

	if result.NetworkItemType != "HOST" {
		t.Errorf("Expected NetworkItemType 'HOST', got %s", result.NetworkItemType)
	}

	if result.NetworkItemID != "host-456" {
		t.Errorf("Expected NetworkItemID 'host-456', got %s", result.NetworkItemID)
	}

	if result.Type != "CUSTOM" {
		t.Errorf("Expected Type 'CUSTOM', got %s", result.Type)
	}

	// Verify Config is properly populated
	if result.Config == nil {
		t.Fatal("Expected Config to be populated, got nil")
	}

	if len(result.Config.ServiceTypes) != 2 {
		t.Errorf("Expected 2 service types, got %d", len(result.Config.ServiceTypes))
	}

	if len(result.Config.CustomServiceTypes) != 1 {
		t.Errorf("Expected 1 custom service type, got %d", len(result.Config.CustomServiceTypes))
	}

	customType := result.Config.CustomServiceTypes[0]
	if customType.Protocol != "TCP" {
		t.Errorf("Expected protocol 'TCP', got %s", customType.Protocol)
	}

	if len(customType.Port) != 2 {
		t.Errorf("Expected 2 port ranges, got %d", len(customType.Port))
	}
}

func TestHostIPServicesService_List_UpdatedDTO(t *testing.T) {
	// Test that the list endpoint works with the updated DTO structure
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Mock paginated response with updated DTO
		// Note: The List() method calls GetIPByPage() which may make multiple requests
		response := HostIPServicePageResponse{
			Content: []HostIPServiceResponse{
				{
					ID:              "service-1",
					Name:            "Service 1",
					Description:     "First service",
					NetworkItemType: "HOST",
					NetworkItemID:   "host-1",
					Type:            "PREDEFINED",
					Config: &IPServiceConfig{
						ServiceTypes: []string{"SSH"},
					},
				},
				{
					ID:              "service-2",
					Name:            "Service 2",
					Description:     "Second service",
					NetworkItemType: "HOST",
					NetworkItemID:   "host-2",
					Type:            "CUSTOM",
					Config: &IPServiceConfig{
						CustomServiceTypes: []*CustomIPServiceType{
							{
								Protocol: "UDP",
								Port: []Range{
									{LowerValue: 53, UpperValue: 53},
								},
							},
						},
					},
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

	client := createTestHostIPServicesClient(server)

	// Test the GetIPByPage method directly to avoid pagination issues
	result, err := client.HostIPServices.GetIPByPage(0, 10)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Content) != 2 {
		t.Errorf("Expected 2 services, got %d", len(result.Content))
	}

	// Verify first service
	service1 := result.Content[0]
	if service1.Type != "PREDEFINED" {
		t.Errorf("Expected first service type 'PREDEFINED', got %s", service1.Type)
	}

	if len(service1.Config.ServiceTypes) != 1 {
		t.Errorf("Expected 1 predefined service type, got %d", len(service1.Config.ServiceTypes))
	}

	// Verify second service
	service2 := result.Content[1]
	if service2.Type != "CUSTOM" {
		t.Errorf("Expected second service type 'CUSTOM', got %s", service2.Type)
	}

	if len(service2.Config.CustomServiceTypes) != 1 {
		t.Errorf("Expected 1 custom service type, got %d", len(service2.Config.CustomServiceTypes))
	}
}

func TestHostIPServicesService_Create_UpdatedDTO(t *testing.T) {
	// Test that create operations work with the updated DTO structure
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Mock response with created service using updated DTO
		service := HostIPServiceResponse{
			ID:              "new-service-123",
			Name:            "New Service",
			Description:     "Newly created service",
			NetworkItemType: "HOST",
			NetworkItemID:   "host-789",
			Type:            "CUSTOM",
			Config: &IPServiceConfig{
				CustomServiceTypes: []*CustomIPServiceType{
					{
						Protocol: "TCP",
						Port: []Range{
							{LowerValue: 8080, UpperValue: 8080},
						},
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(service)
	}))
	defer server.Close()

	client := createTestHostIPServicesClient(server)

	// Create a new IP service
	newService := &IPService{
		Name:            "New Service",
		Description:     "Newly created service",
		NetworkItemType: "HOST",
		NetworkItemID:   "host-789",
		Type:            "CUSTOM",
		Config: &IPServiceConfig{
			CustomServiceTypes: []*CustomIPServiceType{
				{
					Protocol: "TCP",
					Port: []Range{
						{LowerValue: 8080, UpperValue: 8080},
					},
				},
			},
		},
	}

	result, err := client.HostIPServices.Create(newService)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	// Verify the created service response
	if result.ID != "new-service-123" {
		t.Errorf("Expected service ID 'new-service-123', got %s", result.ID)
	}

	if result.Name != "New Service" {
		t.Errorf("Expected service name 'New Service', got %s", result.Name)
	}

	// Verify that the response doesn't contain duplicate routing information
	if result.Config == nil {
		t.Fatal("Expected Config to be populated, got nil")
	}

	if len(result.Config.CustomServiceTypes) != 1 {
		t.Errorf("Expected 1 custom service type, got %d", len(result.Config.CustomServiceTypes))
	}
}

func TestIPServiceResponse_NoRoutesDuplication(t *testing.T) {
	// Test that the updated HostIPServiceResponse structure doesn't have duplicate routes
	// This is a structural test to ensure API v1.1.0 compliance

	service := HostIPServiceResponse{
		ID:              "test-service",
		Name:            "Test Service",
		Description:     "Test description",
		NetworkItemType: "HOST",
		NetworkItemID:   "host-123",
		Type:            "CUSTOM",
		Config: &IPServiceConfig{
			ServiceTypes: []string{"HTTP"},
		},
	}

	// Serialize to JSON to verify structure
	jsonData, err := json.Marshal(service)
	if err != nil {
		t.Fatalf("Failed to marshal service: %v", err)
	}

	// Parse back to verify structure
	var parsed map[string]interface{}
	err = json.Unmarshal(jsonData, &parsed)
	if err != nil {
		t.Fatalf("Failed to unmarshal service: %v", err)
	}

	// Verify that there's no duplicate 'routes' field at the top level
	// (routes should only be in the config if needed)
	if _, exists := parsed["routes"]; exists {
		t.Error("HostIPServiceResponse should not have a top-level 'routes' field in API v1.1.0")
	}

	// Verify expected fields are present
	expectedFields := []string{"id", "name", "description", "networkItemType", "networkItemId", "type", "config"}
	for _, field := range expectedFields {
		if _, exists := parsed[field]; !exists {
			t.Errorf("Expected field '%s' not found in HostIPServiceResponse", field)
		}
	}

	t.Logf("HostIPServiceResponse JSON structure: %s", string(jsonData))
}
