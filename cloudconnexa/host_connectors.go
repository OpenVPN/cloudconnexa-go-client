package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// HostConnector represents a host connector in CloudConnexa.
type HostConnector struct {
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

// HostConnectorPageResponse represents a paginated response of host connectors.
type HostConnectorPageResponse struct {
	Content          []HostConnector `json:"content"`
	NumberOfElements int             `json:"numberOfElements"`
	Page             int             `json:"page"`
	Size             int             `json:"size"`
	Success          bool            `json:"success"`
	TotalElements    int             `json:"totalElements"`
	TotalPages       int             `json:"totalPages"`
}

// HostConnectorsService provides methods for managing host connectors.
type HostConnectorsService service

// GetByPage retrieves host connectors using pagination.
func (c *HostConnectorsService) GetByPage(page int, pageSize int) (HostConnectorPageResponse, error) {
	return c.GetByPageAndHostID(page, pageSize, "")
}

// GetByPageAndHostID retrieves host connectors using pagination, optionally filtered by host ID.
func (c *HostConnectorsService) GetByPageAndHostID(page int, pageSize int, hostID string) (HostConnectorPageResponse, error) {
	var endpoint string
	if hostID != "" {
		endpoint = fmt.Sprintf("%s/hosts/connectors?hostId=%s&page=%d&size=%d", c.client.GetV1Url(), hostID, page, pageSize)
	} else {
		endpoint = fmt.Sprintf("%s/hosts/connectors?page=%d&size=%d", c.client.GetV1Url(), page, pageSize)
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return HostConnectorPageResponse{}, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return HostConnectorPageResponse{}, err
	}

	var response HostConnectorPageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return HostConnectorPageResponse{}, err
	}
	return response, nil
}

// Update updates an existing host connector.
func (c *HostConnectorsService) Update(connector HostConnector) (*HostConnector, error) {
	connectorJSON, err := json.Marshal(connector)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/hosts/connectors/%s", c.client.GetV1Url(), connector.ID), bytes.NewBuffer(connectorJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var conn HostConnector
	err = json.Unmarshal(body, &conn)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

// List retrieves all host connectors.
func (c *HostConnectorsService) List() ([]HostConnector, error) {
	return c.ListByHostID("")
}

// ListByHostID retrieves all host connectors for a specific host ID.
func (c *HostConnectorsService) ListByHostID(hostID string) ([]HostConnector, error) {
	var allConnectors []HostConnector
	page := 0
	pageSize := 10

	for {
		response, err := c.GetByPageAndHostID(page, pageSize, hostID)
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

// GetByID retrieves a specific host connector by ID.
func (c *HostConnectorsService) GetByID(id string) (*HostConnector, error) {
	endpoint := fmt.Sprintf("%s/hosts/connectors/%s", c.client.GetV1Url(), id)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var connector HostConnector
	err = json.Unmarshal(body, &connector)
	if err != nil {
		return nil, err
	}
	return &connector, nil
}

// GetProfile retrieves the profile configuration for a host connector.
func (c *HostConnectorsService) GetProfile(id string) (string, error) {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/hosts/connectors/%s/profile", c.client.GetV1Url(), id), nil)
	if err != nil {
		return "", err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// GetToken retrieves an encrypted token for a host connector.
func (c *HostConnectorsService) GetToken(id string) (string, error) {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/hosts/connectors/%s/profile/encrypt", c.client.GetV1Url(), id), nil)
	if err != nil {
		return "", err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// Create creates a new host connector for the specified host.
func (c *HostConnectorsService) Create(connector HostConnector, hostID string) (*HostConnector, error) {
	connectorJSON, err := json.Marshal(connector)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/hosts/connectors?hostId=%s", c.client.GetV1Url(), hostID), bytes.NewBuffer(connectorJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var conn HostConnector
	err = json.Unmarshal(body, &conn)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

// Delete deletes a host connector by ID.
func (c *HostConnectorsService) Delete(connectorID string, hostID string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/hosts/connectors/%s?hostId=%s", c.client.GetV1Url(), connectorID, hostID), nil)
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}
