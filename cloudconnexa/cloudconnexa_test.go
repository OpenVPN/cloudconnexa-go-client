package cloudconnexa

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

// setupMockServer creates a test HTTP server that simulates the CloudConnexa API endpoints
// for testing purposes. It handles token authentication and basic endpoint responses.
func setupMockServer() *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/oauth/token":
			if r.Method == "POST" {
				response := Credentials{AccessToken: "mocked-token"}
				err := json.NewEncoder(w).Encode(response)
				if err != nil {
					log.Printf("Mock server: error encoding response: %v\n", err)
				}
			} else {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			}
		case "/valid-endpoint":
			if r.Method != "GET" {
				http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			}
		default:
			http.Error(w, "Not Found", http.StatusNotFound)
		}
	}))

	return server
}

// TestNewClient tests the creation of a new CloudConnexa client with various credential combinations.
// It verifies that the client is properly initialized with valid credentials and returns
// appropriate errors for invalid credentials.
func TestNewClient(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	tests := []struct {
		name         string
		baseURL      string
		clientID     string
		clientSecret string
		wantErr      bool
	}{
		{"Valid Credentials", server.URL, "test-id", "test-secret", false},
		{"Empty ClientID", server.URL, "", "test-secret", true},
		{"Empty ClientSecret", server.URL, "test-id", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.baseURL, tt.clientID, tt.clientSecret)

			if tt.wantErr {
				assert.Error(t, err, "NewClient should return an error for invalid credentials")
			} else {
				assert.NoError(t, err, "NewClient should not return an error for valid credentials")
				assert.NotNil(t, client, "Client should not be nil for valid credentials")
			}
		})
	}
}

// TestDoRequest tests the DoRequest method of the CloudConnexa client.
// It verifies that the client correctly handles various HTTP requests and responses,
// including valid requests, invalid endpoints, and incorrect HTTP methods.
func TestDoRequest(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	client := &Client{
		client:            server.Client(),
		BaseURL:           server.URL,
		Token:             "mock-access-token",
		ReadRateLimiter:   rate.NewLimiter(rate.Every(1), 5),
		UpdateRateLimiter: rate.NewLimiter(rate.Every(1), 5),
	}

	tests := []struct {
		name      string
		method    string
		endpoint  string
		wantError bool
	}{
		{"Valid Request", "GET", "/valid-endpoint", false},
		{"Invalid Endpoint", "GET", "/invalid-endpoint", true},
		{"Invalid Method", "POST", "/valid-endpoint", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, client.BaseURL+tt.endpoint, nil)
			resp, err := client.DoRequest(req)

			if tt.wantError {
				assert.Error(t, err, "DoRequest should return an error")
			} else {
				assert.NoError(t, err, "DoRequest should not return an error")
				assert.NotNil(t, resp, "Response should not be nil for valid requests")
			}
		})
	}
}

// TestBuildURL tests the buildURL function that constructs URLs with escaped path segments.
func TestBuildURL(t *testing.T) {
	tests := []struct {
		name     string
		base     string
		segments []string
		expected string
	}{
		{
			name:     "no segments",
			base:     "https://api.example.com/v1",
			segments: []string{},
			expected: "https://api.example.com/v1",
		},
		{
			name:     "single segment",
			base:     "https://api.example.com/v1",
			segments: []string{"users"},
			expected: "https://api.example.com/v1/users",
		},
		{
			name:     "multiple segments",
			base:     "https://api.example.com/v1",
			segments: []string{"users", "abc-123", "activate"},
			expected: "https://api.example.com/v1/users/abc-123/activate",
		},
		{
			name:     "path traversal escaped",
			base:     "https://api.example.com/v1",
			segments: []string{"users", "../admin"},
			expected: "https://api.example.com/v1/users/..%2Fadmin",
		},
		{
			name:     "forward slash escaped",
			base:     "https://api.example.com/v1",
			segments: []string{"users", "user/admin"},
			expected: "https://api.example.com/v1/users/user%2Fadmin",
		},
		{
			name:     "space escaped",
			base:     "https://api.example.com/v1",
			segments: []string{"users", "user 123"},
			expected: "https://api.example.com/v1/users/user%20123",
		},
		{
			name:     "question mark escaped",
			base:     "https://api.example.com/v1",
			segments: []string{"users", "user?role=admin"},
			expected: "https://api.example.com/v1/users/user%3Frole=admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildURL(tt.base, tt.segments...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestValidateID tests the validateID function that validates ID parameters.
func TestValidateID(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		wantErr error
	}{
		{
			name:    "valid ID",
			id:      "abc-123",
			wantErr: nil,
		},
		{
			name:    "valid UUID",
			id:      "550e8400-e29b-41d4-a716-446655440000",
			wantErr: nil,
		},
		{
			name:    "empty ID",
			id:      "",
			wantErr: ErrEmptyID,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateID(tt.id)
			if tt.wantErr != nil {
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
