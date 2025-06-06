package cloudconnexa

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// VpnRegion represents a VPN region configuration
type VpnRegion struct {
	ID         string `json:"id"`
	Continent  string `json:"continent"`
	Country    string `json:"country"`
	CountryISO string `json:"countryIso"`
	RegionName string `json:"regionName"`
}

// VPNRegionsService provides methods for managing VPN regions
type VPNRegionsService service

// List retrieves all VPN regions
// Returns a slice of VPN regions and any error that occurred
func (c *VPNRegionsService) List() ([]VpnRegion, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/regions", c.client.GetV1Url()), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var regions []VpnRegion
	err = json.Unmarshal(body, &regions)
	if err != nil {
		return nil, err
	}
	return regions, nil
}

// GetByID retrieves a specific VPN region by ID
// regionID: The ID of the VPN region to retrieve
// Returns the VPN region and any error that occurred
func (c *VPNRegionsService) GetByID(regionID string) (*VpnRegion, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/regions", c.client.GetV1Url()), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var vpnRegions []VpnRegion
	err = json.Unmarshal(body, &vpnRegions)
	if err != nil {
		return nil, err
	}

	for _, r := range vpnRegions {
		if r.ID == regionID {
			return &r, nil
		}
	}
	return nil, nil
}
