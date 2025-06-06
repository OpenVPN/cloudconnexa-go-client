package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// NetworkConnector represents a network connector configuration.
type NetworkConnector struct {
	ID               string `json:"id,omitempty"`
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	NetworkItemID    string `json:"networkItemId"`
	NetworkItemType  string `json:"networkItemType"`
	VpnRegionID      string `json:"vpnRegionId"`
	IPv4Address      string `json:"ipV4Address"`
	IPv6Address      string `json:"ipV6Address"`
	Profile          string `json:"profile"`
	ConnectionStatus string `json:"connectionStatus"`
}

// NetworkConnectorPageResponse represents a paginated response of network connectors.
type NetworkConnectorPageResponse struct {
	Content          []NetworkConnector `json:"content"`
	NumberOfElements int                `json:"numberOfElements"`
	Page             int                `json:"page"`
	Size             int                `json:"size"`
	Success          bool               `json:"success"`
	TotalElements    int                `json:"totalElements"`
	TotalPages       int                `json:"totalPages"`
}

// NetworkConnectorsService provides methods for managing network connectors.
type NetworkConnectorsService service

// GetByPage retrieves network connectors using pagination.
func (c *NetworkConnectorsService) GetByPage(page int, pageSize int) (NetworkConnectorPageResponse, error) {
	return c.GetByPageAndNetworkID(page, pageSize, "")
}

// GetByPageAndNetworkID retrieves network connectors for a specific network using pagination.
func (c *NetworkConnectorsService) GetByPageAndNetworkID(page int, pageSize int, networkID string) (NetworkConnectorPageResponse, error) {
	params := url.Values{}
	params.Add("page", strconv.Itoa(page))
	params.Add("size", strconv.Itoa(pageSize))
	if networkID != "" {
		params.Add("networkId", networkID)
	}

	endpoint := fmt.Sprintf("%s/networks/connectors?%s", c.client.GetV1Url(), params.Encode())
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return NetworkConnectorPageResponse{}, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return NetworkConnectorPageResponse{}, err
	}

	var response NetworkConnectorPageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return NetworkConnectorPageResponse{}, err
	}
	return response, nil
}

// Update updates an existing network connector.
func (c *NetworkConnectorsService) Update(connector NetworkConnector) (*NetworkConnector, error) {
	connectorJSON, err := json.Marshal(connector)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/networks/connectors/%s", c.client.GetV1Url(), connector.ID), bytes.NewBuffer(connectorJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var conn NetworkConnector
	err = json.Unmarshal(body, &conn)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

// List retrieves all network connectors by paginating through all available pages.
func (c *NetworkConnectorsService) List() ([]NetworkConnector, error) {
	var allConnectors []NetworkConnector
	page := 0
	pageSize := 10

	for {
		response, err := c.GetByPage(page, pageSize)
		if err != nil {
			return nil, err
		}

		allConnectors = append(allConnectors, response.Content...)

		if page >= response.TotalPages {
			break
		}
		page++
	}
	return allConnectors, nil
}

// ListByNetworkID retrieves all network connectors for a specific network by paginating through all available pages.
func (c *NetworkConnectorsService) ListByNetworkID(networkID string) ([]NetworkConnector, error) {
	var allConnectors []NetworkConnector
	page := 0
	pageSize := 10

	for {
		response, err := c.GetByPageAndNetworkID(page, pageSize, networkID)
		if err != nil {
			return nil, err
		}

		allConnectors = append(allConnectors, response.Content...)

		if page >= response.TotalPages {
			break
		}
		page++
	}
	return allConnectors, nil
}

// GetByID retrieves a specific network connector by its ID.
func (c *NetworkConnectorsService) GetByID(id string) (*NetworkConnector, error) {
	endpoint := fmt.Sprintf("%s/networks/connectors/%s", c.client.GetV1Url(), id)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var connector NetworkConnector
	err = json.Unmarshal(body, &connector)
	if err != nil {
		return nil, err
	}
	return &connector, nil
}

// GetProfile retrieves the profile configuration for a specific network connector.
func (c *NetworkConnectorsService) GetProfile(id string) (string, error) {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/networks/connectors/%s/profile", c.client.GetV1Url(), id), nil)
	if err != nil {
		return "", err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// GetToken retrieves an encrypted token for a specific network connector.
func (c *NetworkConnectorsService) GetToken(id string) (string, error) {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/networks/connectors/%s/profile/encrypt", c.client.GetV1Url(), id), nil)
	if err != nil {
		return "", err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// Create creates a new network connector.
func (c *NetworkConnectorsService) Create(connector NetworkConnector, networkID string) (*NetworkConnector, error) {
	connectorJSON, err := json.Marshal(connector)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/networks/connectors?networkId=%s", c.client.GetV1Url(), networkID), bytes.NewBuffer(connectorJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var conn NetworkConnector
	err = json.Unmarshal(body, &conn)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

// Delete removes a network connector by its ID and network ID.
func (c *NetworkConnectorsService) Delete(connectorID string, networkID string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/networks/connectors/%s?networkId=%s", c.client.GetV1Url(), connectorID, networkID), nil)
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}

// IPsecStartResponse represents the response from starting an IPsec tunnel.
type IPsecStartResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Status  string `json:"status,omitempty"`
}

// IPsecStopResponse represents the response from stopping an IPsec tunnel.
type IPsecStopResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Status  string `json:"status,omitempty"`
}

// StartIPsec starts an IPsec tunnel for the specified network connector.
func (c *NetworkConnectorsService) StartIPsec(connectorID string) (*IPsecStartResponse, error) {
	endpoint := fmt.Sprintf("%s/networks/connectors/%s/ipsec/start", c.client.GetV1Url(), connectorID)
	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var response IPsecStartResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// StopIPsec stops an IPsec tunnel for the specified network connector.
func (c *NetworkConnectorsService) StopIPsec(connectorID string) (*IPsecStopResponse, error) {
	endpoint := fmt.Sprintf("%s/networks/connectors/%s/ipsec/stop", c.client.GetV1Url(), connectorID)
	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var response IPsecStopResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
