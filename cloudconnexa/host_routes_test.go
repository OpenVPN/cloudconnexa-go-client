package cloudconnexa

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/time/rate"
)

// createTestClientWithHostRoutes creates a test client with HostRoutes service
func createTestClientWithHostRoutes(server *httptest.Server) *Client {
	client := &Client{
		client:            server.Client(),
		BaseURL:           server.URL,
		Token:             "test-token",
		ReadRateLimiter:   rate.NewLimiter(rate.Every(1), 5),
		UpdateRateLimiter: rate.NewLimiter(rate.Every(1), 5),
	}
	client.HostRoutes = (*HostRoutesService)(&service{client: client})
	return client
}

func TestHostRoutesService_GetByPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/hosts/routes" {
			t.Errorf("Expected path /api/v1/hosts/routes, got %s", r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("hostId") != "host-123" {
			t.Errorf("Expected hostId=host-123, got %s", query.Get("hostId"))
		}

		response := HostRoutePageResponse{
			Content: []HostRoute{
				{ID: "route-1", Subnet: "10.0.0.0/24", Description: "Route 1"},
				{ID: "route-2", Subnet: "10.0.1.0/24", Description: "Route 2"},
			},
			TotalElements:    2,
			TotalPages:       1,
			NumberOfElements: 2,
			Page:             0,
			Size:             100,
			Success:          true,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := createTestClientWithHostRoutes(server)

	result, err := client.HostRoutes.GetByPage("host-123", 0, 100)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(result.Content) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(result.Content))
	}

	if result.Content[0].ID != "route-1" {
		t.Errorf("Expected route ID 'route-1', got %s", result.Content[0].ID)
	}
}

func TestHostRoutesService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := HostRoutePageResponse{
			Content: []HostRoute{
				{ID: "route-1", Subnet: "10.0.0.0/24"},
				{ID: "route-2", Subnet: "10.0.1.0/24"},
			},
			TotalElements:    2,
			TotalPages:       1,
			NumberOfElements: 2,
			Page:             0,
			Size:             100,
			Success:          true,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := createTestClientWithHostRoutes(server)

	routes, err := client.HostRoutes.List("host-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(routes) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(routes))
	}
}

func TestHostRoutesService_GetByID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/hosts/routes/route-123" {
			t.Errorf("Expected path /api/v1/hosts/routes/route-123, got %s", r.URL.Path)
		}

		route := HostRoute{
			ID:          "route-123",
			Subnet:      "10.0.0.0/24",
			Description: "Test route",
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(route)
	}))
	defer server.Close()

	client := createTestClientWithHostRoutes(server)

	route, err := client.HostRoutes.GetByID("route-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if route.ID != "route-123" {
		t.Errorf("Expected route ID 'route-123', got %s", route.ID)
	}

	if route.Subnet != "10.0.0.0/24" {
		t.Errorf("Expected subnet '10.0.0.0/24', got %s", route.Subnet)
	}
}

func TestHostRoutesService_Create(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/hosts/routes" {
			t.Errorf("Expected path /api/v1/hosts/routes, got %s", r.URL.Path)
		}

		query := r.URL.Query()
		if query.Get("hostId") != "host-123" {
			t.Errorf("Expected hostId=host-123, got %s", query.Get("hostId"))
		}

		route := HostRoute{
			ID:          "route-new",
			Subnet:      "10.0.2.0/24",
			Description: "New route",
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(route)
	}))
	defer server.Close()

	client := createTestClientWithHostRoutes(server)

	newRoute := HostRoute{
		Subnet:      "10.0.2.0/24",
		Description: "New route",
	}

	route, err := client.HostRoutes.Create("host-123", newRoute)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if route.ID != "route-new" {
		t.Errorf("Expected route ID 'route-new', got %s", route.ID)
	}
}

func TestHostRoutesService_Update(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/hosts/routes/route-123" {
			t.Errorf("Expected path /api/v1/hosts/routes/route-123, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestClientWithHostRoutes(server)

	route := HostRoute{
		ID:          "route-123",
		Subnet:      "10.0.3.0/24",
		Description: "Updated route",
	}

	err := client.HostRoutes.Update(route)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestHostRoutesService_Delete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/hosts/routes/route-123" {
			t.Errorf("Expected path /api/v1/hosts/routes/route-123, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := createTestClientWithHostRoutes(server)

	err := client.HostRoutes.Delete("route-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestHostRoutesService_GetByID_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "route not found"}`))
	}))
	defer server.Close()

	client := createTestClientWithHostRoutes(server)

	_, err := client.HostRoutes.GetByID("nonexistent-route")
	if err == nil {
		t.Error("Expected error for non-existent route, got nil")
	}
}
