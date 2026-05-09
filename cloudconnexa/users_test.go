package cloudconnexa

import (
	"encoding/json"
	"errors"
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

func TestUsersService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/v1/users" {
			t.Errorf("Expected path /api/v1/users, got %s", r.URL.Path)
		}

		response := UserPageResponse{
			Content: []User{
				{ID: "user-1", Username: "alice", Role: "ADMIN"},
				{ID: "user-2", Username: "bob", Role: "MEMBER"},
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

	client := createTestClientWithUsers(server)

	users, err := client.Users.List()
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	if users[0].Username != "alice" {
		t.Errorf("Expected first user 'alice', got %s", users[0].Username)
	}
}

func TestUsersService_FindByUsernameAndRole(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := UserPageResponse{
			Content: []User{
				{ID: "user-1", Username: "alice", Role: "ADMIN"},
				{ID: "user-2", Username: "bob", Role: "MEMBER"},
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

	client := createTestClientWithUsers(server)

	user, err := client.Users.FindByUsernameAndRole("bob", "MEMBER")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if user.ID != "user-2" {
		t.Errorf("Expected user ID 'user-2', got %s", user.ID)
	}
}

func TestUsersService_FindByUsernameAndRole_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		response := UserPageResponse{
			Content: []User{
				{ID: "user-1", Username: "alice", Role: "ADMIN"},
			},
			NumberOfElements: 1,
			Page:             0,
			Size:             100,
			Success:          true,
			TotalElements:    1,
			TotalPages:       1,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := createTestClientWithUsers(server)

	_, err := client.Users.FindByUsernameAndRole("nobody", "ADMIN")
	if !errors.Is(err, ErrUserNotFound) {
		t.Errorf("Expected ErrUserNotFound, got %v", err)
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
