package cloudconnexa

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/time/rate"
)

// createTestClientWithUsers creates a test client with Users service
func createTestClientWithUsers(server *httptest.Server) *Client {
	client := &Client{
		client:            server.Client(),
		BaseURL:           server.URL,
		Token:             "test-token",
		ReadRateLimiter:   rate.NewLimiter(rate.Every(1), 5),
		UpdateRateLimiter: rate.NewLimiter(rate.Every(1), 5),
	}
	client.Users = (*UsersService)(&service{client: client})
	return client
}

func TestUsersService_Activate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		// Check the path
		if r.URL.Path != "/api/v1/users/user-123/activate" {
			t.Errorf("Expected path /api/v1/users/user-123/activate, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestClientWithUsers(server)

	err := client.Users.Activate("user-123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestUsersService_Suspend(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request method
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		// Check the path
		if r.URL.Path != "/api/v1/users/user-456/suspend" {
			t.Errorf("Expected path /api/v1/users/user-456/suspend, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := createTestClientWithUsers(server)

	err := client.Users.Suspend("user-456")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestUsersService_Activate_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"error": "user not found"}`))
	}))
	defer server.Close()

	client := createTestClientWithUsers(server)

	err := client.Users.Activate("nonexistent-user")
	if err == nil {
		t.Error("Expected error for non-existent user, got nil")
	}
}

func TestUsersService_Suspend_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "user already suspended"}`))
	}))
	defer server.Close()

	client := createTestClientWithUsers(server)

	err := client.Users.Suspend("already-suspended-user")
	if err == nil {
		t.Error("Expected error for already suspended user, got nil")
	}
}

func TestUser_LicensedField(t *testing.T) {
	user := User{
		ID:       "user-123",
		Username: "testuser",
		Licensed: true,
	}

	if !user.Licensed {
		t.Error("Expected Licensed to be true")
	}

	user.Licensed = false
	if user.Licensed {
		t.Error("Expected Licensed to be false")
	}
}
