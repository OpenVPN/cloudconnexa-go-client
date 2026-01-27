package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
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
	Licensed         bool   `json:"licensed"`
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
	params := url.Values{}
	params.Set("page", strconv.Itoa(page))
	params.Set("size", strconv.Itoa(pageSize))
	if hostID != "" {
		params.Set("hostId", hostID)
	}

	endpoint := fmt.Sprintf("%s/hosts/connectors?%s", c.client.GetV1Url(), params.Encode())
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
	if err := validateID(connector.ID); err != nil {
		return nil, err
	}
	connectorJSON, err := json.Marshal(connector)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, buildURL(c.client.GetV1Url(), "hosts", "connectors", connector.ID), bytes.NewBuffer(connectorJSON))
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

	for {
		response, err := c.GetByPageAndHostID(page, defaultPageSize, hostID)
		if err != nil {
			return nil, err
		}

		allConnectors = append(allConnectors, response.Content...)

		page++
		if page >= response.TotalPages {
			break
		}
	}
	return allConnectors, nil
}

// GetByID retrieves a specific host connector by ID.
func (c *HostConnectorsService) GetByID(id string) (*HostConnector, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	endpoint := buildURL(c.client.GetV1Url(), "hosts", "connectors", id)
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

// GetByName retrieves a connector by its name
// name: The name of the connector to retrieve
// Returns the connector and any error that occurred
func (c *HostConnectorsService) GetByName(name string) (*HostConnector, error) {
	items, err := c.List()
	if err != nil {
		return nil, err
	}

	filtered := make([]HostConnector, 0)
	for _, item := range items {
		if item.Name == name {
			filtered = append(filtered, item)
		}
	}
	if len(filtered) > 1 {
		return nil, errors.New("different host connectors found with name: " + name)
	}
	if len(filtered) == 1 {
		return &filtered[0], nil
	}
	return nil, errors.New("host connector not found")
}

// GetProfile retrieves the profile configuration for a host connector.
func (c *HostConnectorsService) GetProfile(id string) (string, error) {
	if err := validateID(id); err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, buildURL(c.client.GetV1Url(), "hosts", "connectors", id, "profile"), nil)
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
	if err := validateID(id); err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, buildURL(c.client.GetV1Url(), "hosts", "connectors", id, "profile", "encrypt"), nil)
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
	if err := validateID(hostID); err != nil {
		return nil, err
	}
	connectorJSON, err := json.Marshal(connector)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("hostId", hostID)
	endpoint := fmt.Sprintf("%s/hosts/connectors?%s", c.client.GetV1Url(), params.Encode())
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(connectorJSON))
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
	if err := validateID(connectorID); err != nil {
		return err
	}
	if err := validateID(hostID); err != nil {
		return err
	}
	params := url.Values{}
	params.Set("hostId", hostID)
	endpoint := fmt.Sprintf("%s?%s", buildURL(c.client.GetV1Url(), "hosts", "connectors", connectorID), params.Encode())
	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}

// Activate activates a suspended host connector.
// connectorID: The ID of the connector to activate
// Returns any error that occurred
func (c *HostConnectorsService) Activate(connectorID string) error {
	if err := validateID(connectorID); err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, buildURL(c.client.GetV1Url(), "hosts", "connectors", connectorID, "activate"), nil)
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}

// Suspend suspends an active host connector.
// connectorID: The ID of the connector to suspend
// Returns any error that occurred
func (c *HostConnectorsService) Suspend(connectorID string) error {
	if err := validateID(connectorID); err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPut, buildURL(c.client.GetV1Url(), "hosts", "connectors", connectorID, "suspend"), nil)
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}
