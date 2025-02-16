package cloudconnexa

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testVpnRegion = VpnRegion{
	ID:         "test-region",
	Country:    "Test Country",
	Continent:  "Test Continent",
	CountryISO: "TC",
	RegionName: "Test Region",
}

func TestVPNRegionsService_List(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle auth token request
		if r.URL.Path == "/api/v1/oauth/token" {
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(map[string]string{
				"access_token": "test-token",
			})
			assert.NoError(t, err)
			return
		}

		// Handle VPN regions request
		assert.Equal(t, "/api/v1/regions", r.URL.Path)

		regions := []VpnRegion{testVpnRegion}

		err := json.NewEncoder(w).Encode(regions)
		assert.NoError(t, err)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test", "test")
	assert.NoError(t, err)
	regions, err := client.VPNRegions.List()

	assert.NoError(t, err)
	assert.Equal(t, 1, len(regions))
	assert.Equal(t, testVpnRegion.ID, regions[0].ID)
}

func TestVPNRegionsService_GetByID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Handle auth token request
		if r.URL.Path == "/api/v1/oauth/token" {
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(map[string]string{
				"access_token": "test-token",
			})
			assert.NoError(t, err)
			return
		}

		// Handle VPN regions request
		assert.Equal(t, "/api/v1/regions", r.URL.Path)

		err := json.NewEncoder(w).Encode([]VpnRegion{testVpnRegion})
		assert.NoError(t, err)
	}))
	defer server.Close()

	client, err := NewClient(server.URL, "test", "test")
	assert.NoError(t, err)

	// Test existing region
	region, err := client.VPNRegions.GetByID(testVpnRegion.ID)
	assert.NoError(t, err)
	assert.NotNil(t, region)
	assert.Equal(t, testVpnRegion.ID, region.ID)

	// Test non-existent region
	region, err = client.VPNRegions.GetByID("non-existent")
	assert.NoError(t, err)
	assert.Nil(t, region)
}
