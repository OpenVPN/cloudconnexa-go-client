package cloudconnexa

import (
	"encoding/json"
	"errors"
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
			// Use NewClientWithOptions with AllowInsecureHTTP for localhost test server
			client, err := NewClientWithOptions(tt.baseURL, tt.clientID, tt.clientSecret, &ClientOptions{
				AllowInsecureHTTP: true,
			})

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

// TestDoRequest_ResponseUnderLimit verifies that normal-sized responses succeed.
func TestDoRequest_ResponseUnderLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer server.Close()

	client := &Client{
		client:            server.Client(),
		BaseURL:           server.URL,
		Token:             "mock-access-token",
		ReadRateLimiter:   rate.NewLimiter(rate.Every(1), 5),
		UpdateRateLimiter: rate.NewLimiter(rate.Every(1), 5),
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	body, err := client.DoRequest(req)

	assert.NoError(t, err, "DoRequest should succeed for small response")
	assert.NotEmpty(t, body, "Response body should not be empty")
}

// TestDoRequest_ResponseOverLimit verifies that oversized responses are rejected with ErrResponseTooLarge.
func TestDoRequest_ResponseOverLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Write more than DefaultMaxResponseSize to trigger the limit
		data := make([]byte, DefaultMaxResponseSize+1)
		_, _ = w.Write(data)
	}))
	defer server.Close()

	client := &Client{
		client:            server.Client(),
		BaseURL:           server.URL,
		Token:             "mock-access-token",
		ReadRateLimiter:   rate.NewLimiter(rate.Every(1), 5),
		UpdateRateLimiter: rate.NewLimiter(rate.Every(1), 5),
	}

	req, _ := http.NewRequest("GET", server.URL+"/test", nil)
	_, err := client.DoRequest(req)

	assert.Error(t, err, "DoRequest should fail for oversized response")
	assert.True(t, errors.Is(err, ErrResponseTooLarge), "Error should be ErrResponseTooLarge, got: %v", err)
}

// TestNewClient_TokenResponseOverLimit verifies that oversized OAuth token responses are rejected.
func TestNewClient_TokenResponseOverLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/oauth/token" {
			// Return response larger than token limit
			data := make([]byte, DefaultMaxTokenResponseSize+1)
			_, _ = w.Write(data)
		}
	}))
	defer server.Close()

	_, err := NewClientWithOptions(server.URL, "client-id", "client-secret", &ClientOptions{
		AllowInsecureHTTP: true,
	})

	assert.Error(t, err, "NewClient should fail for oversized OAuth response")
	assert.True(t, errors.Is(err, ErrResponseTooLarge), "Error should be ErrResponseTooLarge, got: %v", err)
}

// TestValidateBaseURL tests the validateBaseURL function that validates base URL parameters.
func TestValidateBaseURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		allowHTTP bool
		wantErr   error
		wantURL   string
	}{
		// Valid HTTPS URLs
		{"valid https", "https://api.example.com", false, nil, "https://api.example.com"},
		{"https with port", "https://api.example.com:443", false, nil, "https://api.example.com:443"},
		{"https with path stripped", "https://api.example.com/v1/", false, nil, "https://api.example.com"},
		{"https trailing slash stripped", "https://api.example.com/", false, nil, "https://api.example.com"},

		// Invalid: HTTP without allowHTTP
		{"http not allowed", "http://api.example.com", false, ErrHTTPSRequired, ""},

		// Valid: HTTP with allowHTTP for loopback
		{"http localhost allowed", "http://localhost:8080", true, nil, "http://localhost:8080"},
		{"http 127.0.0.1 allowed", "http://127.0.0.1:8080", true, nil, "http://127.0.0.1:8080"},
		{"http 127.0.0.2 allowed", "http://127.0.0.2:9999", true, nil, "http://127.0.0.2:9999"},
		{"http ::1 allowed", "http://[::1]:8080", true, nil, "http://[::1]:8080"},

		// Invalid: HTTP with allowHTTP for non-loopback
		{"http remote rejected even with allowHTTP", "http://api.example.com", true, ErrHTTPSRequired, ""},

		// Invalid URLs
		{"empty url", "", false, ErrInvalidBaseURL, ""},
		{"missing scheme", "api.example.com", false, ErrInvalidBaseURL, ""},
		{"ftp scheme", "ftp://api.example.com", false, ErrInvalidBaseURL, ""},
		{"missing host", "https://", false, ErrInvalidBaseURL, ""},

		// Security: credentials in URL rejected
		{"url with userinfo", "https://user:pass@api.example.com", false, ErrInvalidBaseURL, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validateBaseURL(tt.url, tt.allowHTTP)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tt.wantErr), "expected error %v, got %v", tt.wantErr, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantURL, result)
			}
		})
	}
}

// TestIsLoopbackHost tests the isLoopbackHost function that checks for loopback addresses.
func TestIsLoopbackHost(t *testing.T) {
	tests := []struct {
		host     string
		expected bool
	}{
		{"localhost", true},
		{"localhost:8080", true},
		{"LOCALHOST", true},
		{"127.0.0.1", true},
		{"127.0.0.1:8080", true},
		{"127.0.0.2", true},
		{"127.255.255.255", true},
		{"[::1]", true},
		{"[::1]:8080", true},
		{"::1", true},
		{"api.example.com", false},
		{"192.168.1.1", false},
		{"10.0.0.1", false},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			result := isLoopbackHost(tt.host)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestNewClient_HTTPSRequired tests that NewClient rejects HTTP URLs by default.
func TestNewClient_HTTPSRequired(t *testing.T) {
	_, err := NewClient("http://api.example.com", "id", "secret")
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrHTTPSRequired), "expected ErrHTTPSRequired, got: %v", err)
}

// TestNewClientWithOptions_AllowHTTPForLocalhost tests that NewClientWithOptions allows HTTP for localhost.
func TestNewClientWithOptions_AllowHTTPForLocalhost(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	client, err := NewClientWithOptions(server.URL, "test-id", "test-secret", &ClientOptions{
		AllowInsecureHTTP: true,
	})
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

// TestNewClientWithOptions_HTTPRemoteRejected tests that HTTP to non-localhost is always rejected.
func TestNewClientWithOptions_HTTPRemoteRejected(t *testing.T) {
	_, err := NewClientWithOptions("http://api.example.com", "test-id", "test-secret", &ClientOptions{
		AllowInsecureHTTP: true,
	})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrHTTPSRequired), "expected ErrHTTPSRequired, got: %v", err)
}
