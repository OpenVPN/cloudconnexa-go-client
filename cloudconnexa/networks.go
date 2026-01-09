package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

const (
	// InternetAccessSplitTunnelOn enables split tunneling for internet access.
	InternetAccessSplitTunnelOn = "SPLIT_TUNNEL_ON"
	// InternetAccessSplitTunnelOff disables split tunneling for internet access.
	InternetAccessSplitTunnelOff = "SPLIT_TUNNEL_OFF"
	// InternetAccessRestrictedInternet restricts internet access.
	InternetAccessRestrictedInternet = "RESTRICTED_INTERNET"
)

// Network represents a network in CloudConnexa.
type Network struct {
	ID                string             `json:"id"`
	Name              string             `json:"name"`
	Description       string             `json:"description"`
	Egress            bool               `json:"egress"`
	InternetAccess    string             `json:"internetAccess"`
	SystemSubnets     []string           `json:"systemSubnets"`
	Connectors        []NetworkConnector `json:"connectors"`
	Routes            []Route            `json:"routes"`
	TunnelingProtocol string             `json:"tunnelingProtocol"`
}

// NetworkPageResponse represents a paginated response of networks.
type NetworkPageResponse struct {
	Content          []Network `json:"content"`
	NumberOfElements int       `json:"numberOfElements"`
	Page             int       `json:"page"`
	Size             int       `json:"size"`
	Success          bool      `json:"success"`
	TotalElements    int       `json:"totalElements"`
	TotalPages       int       `json:"totalPages"`
}

// NetworksService provides methods for managing networks.
type NetworksService service

// GetByPage retrieves networks using pagination.
// page: The page number to retrieve
// size: The number of items per page
// Returns a NetworkPageResponse containing the networks and pagination information
func (c *NetworksService) GetByPage(page int, size int) (NetworkPageResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/networks?page=%d&size=%d", c.client.GetV1Url(), page, size), nil)
	if err != nil {
		return NetworkPageResponse{}, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return NetworkPageResponse{}, err
	}

	var response NetworkPageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return NetworkPageResponse{}, err
	}

	return response, nil
}

// List retrieves all networks by paginating through all available pages.
// Returns a slice of all networks and any error that occurred
func (c *NetworksService) List() ([]Network, error) {
	var allNetworks []Network
	page := 0

	for {
		response, err := c.GetByPage(page, defaultPageSize)
		if err != nil {
			return nil, err
		}

		allNetworks = append(allNetworks, response.Content...)

		if page >= response.TotalPages {
			break
		}
		page++
	}
	return allNetworks, nil
}

// Get retrieves a specific network by its ID.
// id: The ID of the network to retrieve
// Returns the network and any error that occurred
func (c *NetworksService) Get(id string) (*Network, error) {
	endpoint := fmt.Sprintf("%s/networks/%s", c.client.GetV1Url(), id)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var network Network
	err = json.Unmarshal(body, &network)
	if err != nil {
		return nil, err
	}
	return &network, nil
}

// GetByName retrieves a network by its name
// name: The name of the network to retrieve
// Returns the network and any error that occurred
func (c *NetworksService) GetByName(name string) (*Network, error) {
	items, err := c.List()
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if item.Name == name {
			return &item, nil
		}
	}
	return nil, errors.New("network not found")
}

// Create creates a new network.
// network: The network configuration to create
// Returns the created network and any error that occurred
func (c *NetworksService) Create(network Network) (*Network, error) {
	networkJSON, err := json.Marshal(network)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/networks", c.client.GetV1Url()), bytes.NewBuffer(networkJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var n Network
	err = json.Unmarshal(body, &n)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

// Update updates an existing network.
// network: The updated network configuration
// Returns any error that occurred during the update
func (c *NetworksService) Update(network Network) error {
	networkJSON, err := json.Marshal(network)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/networks/%s", c.client.GetV1Url(), network.ID), bytes.NewBuffer(networkJSON))
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}

// Delete removes a network by its ID.
// networkID: The ID of the network to delete
// Returns any error that occurred during deletion
func (c *NetworksService) Delete(networkID string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/networks/%s", c.client.GetV1Url(), networkID), nil)
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}
