package cloudconnexa

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVPNRegionsService_GetByPage(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/v1/vpn-regions", r.URL.Path)
		assert.Equal(t, "page=0&size=10", r.URL.RawQuery)

		response := VPNRegionPageResponse{
			Content: []VpnRegion{
				{
					ID:         "test-region",
					Country:    "Test Country",
					Continent:  "Test Continent",
					CountryISO: "TC",
				},
			},
			Success:          true,
			NumberOfElements: 1,
			TotalElements:    1,
		}

		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client, _ := NewClient(server.URL, "test", "test")
	response, err := client.VPNRegions.GetByPage(0, 10)

	assert.NoError(t, err)
	assert.True(t, response.Success)
	assert.Equal(t, 1, len(response.Content))
	assert.Equal(t, "test-region", response.Content[0].ID)
}
