package cloudconnexa

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type VpnRegion struct {
	ID         string `json:"id"`
	Continent  string `json:"continent"`
	Country    string `json:"country"`
	CountryISO string `json:"countryIso"`
	RegionName string `json:"regionName"`
}

type VPNRegionsService service

type VPNRegionPageResponse struct {
	Content          []VpnRegion `json:"content"`
	NumberOfElements int         `json:"numberOfElements"`
	Page             int         `json:"page"`
	Size             int         `json:"size"`
	Success          bool        `json:"success"`
	TotalElements    int         `json:"totalElements"`
	TotalPages       int         `json:"totalPages"`
}

// GetByPage retrieves a page of VPN regions
func (c *VPNRegionsService) GetByPage(page int, pageSize int) (VPNRegionPageResponse, error) {
	endpoint := fmt.Sprintf("%s/vpn-regions?page=%d&size=%d", c.client.GetV1Url(), page, pageSize)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return VPNRegionPageResponse{}, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return VPNRegionPageResponse{}, err
	}

	var response VPNRegionPageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return VPNRegionPageResponse{}, err
	}
	return response, nil
}

// List retrieves all VPN regions
func (c *VPNRegionsService) List() ([]VpnRegion, error) {
	var allRegions []VpnRegion
	pageSize := 10
	page := 0

	for {
		response, err := c.GetByPage(page, pageSize)
		if err != nil {
			return nil, err
		}

		allRegions = append(allRegions, response.Content...)

		if page >= response.TotalPages {
			break
		}
		page++
	}
	return allRegions, nil
}

// GetByID retrieves a specific VPN region by ID
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
