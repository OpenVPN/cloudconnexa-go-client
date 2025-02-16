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

func (c *VPNRegionsService) GetVpnRegion(regionID string) (*VpnRegion, error) {
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
