package cloudconnexa

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

func TestNewClient(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	tests := []struct {
		name         string
		baseURL      string
		clientId     string
		clientSecret string
		wantErr      bool
	}{
		{"Valid Credentials", server.URL, "test-id", "test-secret", false},
		{"Empty ClientId", server.URL, "", "test-secret", true},
		{"Empty ClientSecret", server.URL, "test-id", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.baseURL, tt.clientId, tt.clientSecret)

			if tt.wantErr {
				assert.Error(t, err, "NewClient should return an error for invalid credentials")
			} else {
				assert.NoError(t, err, "NewClient should not return an error for valid credentials")
				assert.NotNil(t, client, "Client should not be nil for valid credentials")
			}
		})
	}
}

func TestDoRequest(t *testing.T) {
	server := setupMockServer()
	defer server.Close()

	client := &Client{
		client:      server.Client(),
		BaseURL:     server.URL,
		Token:       "mock-access-token",
		RateLimiter: rate.NewLimiter(rate.Every(1), 5),
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
