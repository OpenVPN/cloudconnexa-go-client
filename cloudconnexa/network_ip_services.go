package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// NetworkIPServiceResponse represents the response structure for Network IP service operations.
type NetworkIPServiceResponse struct {
	Name            string           `json:"name"`
	Description     string           `json:"description"`
	NetworkItemType string           `json:"networkItemType"`
	NetworkItemID   string           `json:"networkItemId"`
	ID              string           `json:"id"`
	Type            string           `json:"type"`
	Config          *IPServiceConfig `json:"config"`
	Routes          []*Route         `json:"routes"`
}

// NetworkIPServicePageResponse represents a paginated response from the CloudConnexa API
// containing a list of IP services and pagination metadata.
type NetworkIPServicePageResponse struct {
	Content          []NetworkIPServiceResponse `json:"content"`
	NumberOfElements int                        `json:"numberOfElements"`
	Page             int                        `json:"page"`
	Size             int                        `json:"size"`
	Success          bool                       `json:"success"`
	TotalElements    int                        `json:"totalElements"`
	TotalPages       int                        `json:"totalPages"`
}

// NetworkIPServicesService handles communication with the CloudConnexa IP Services API
type NetworkIPServicesService service

// GetIPByPage retrieves a page of IP services with pagination
// page: The page number to retrieve
// pageSize: The number of items per page
// Returns a page of IP services and any error that occurred
func (c *NetworkIPServicesService) GetIPByPage(page int, pageSize int) (NetworkIPServicePageResponse, error) {
	endpoint := fmt.Sprintf("%s/networks/ip-services?page=%d&size=%d", c.client.GetV1Url(), page, pageSize)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return NetworkIPServicePageResponse{}, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return NetworkIPServicePageResponse{}, err
	}

	var response NetworkIPServicePageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return NetworkIPServicePageResponse{}, err
	}
	return response, nil
}

// List retrieves all IP services by paginating through all available pages
// Returns a slice of IP services and any error that occurred
func (c *NetworkIPServicesService) List() ([]NetworkIPServiceResponse, error) {
	var allIPServices []NetworkIPServiceResponse
	page := 0

	for {
		response, err := c.GetIPByPage(page, defaultPageSize)
		if err != nil {
			return nil, err
		}

		allIPServices = append(allIPServices, response.Content...)

		page++
		if page >= response.TotalPages {
			break
		}
	}
	return allIPServices, nil
}

// Get retrieves a specific IP service by its ID
// id: The ID of the IP service to retrieve
// Returns the IP service and any error that occurred
func (c *NetworkIPServicesService) Get(id string) (*NetworkIPServiceResponse, error) {
	endpoint := fmt.Sprintf("%s/networks/ip-services/%s", c.client.GetV1Url(), id)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var service NetworkIPServiceResponse
	err = json.Unmarshal(body, &service)
	if err != nil {
		return nil, err
	}
	return &service, nil
}

// GetByName retrieves a network IP service by its name
// name: The name of the network IP service to retrieve
// Returns the network IP service and any error that occurred
func (c *NetworkIPServicesService) GetByName(name string) (*NetworkIPServiceResponse, error) {
	items, err := c.List()
	if err != nil {
		return nil, err
	}

	filtered := make([]NetworkIPServiceResponse, 0)
	for _, item := range items {
		if item.Name == name {
			filtered = append(filtered, item)
		}
	}
	if len(filtered) > 1 {
		return nil, errors.New("different network IP services found with name: " + name)
	}
	if len(filtered) == 1 {
		return &filtered[0], nil
	}
	return nil, errors.New("network IP service not found")
}

// Create creates a new IP service
// ipService: The IP service configuration to create
// Returns the created IP service and any error that occurred
func (c *NetworkIPServicesService) Create(ipService *IPService) (*NetworkIPServiceResponse, error) {
	ipServiceJSON, err := json.Marshal(ipService)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/networks/ip-services?networkId=%s", c.client.GetV1Url(), ipService.NetworkItemID)

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(ipServiceJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var s NetworkIPServiceResponse
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Update updates an existing IP service
// id: The ID of the IP service to update
// service: The updated IP service configuration
// Returns the updated IP service and any error that occurred
func (c *NetworkIPServicesService) Update(id string, service *IPService) (*NetworkIPServiceResponse, error) {
	serviceJSON, err := json.Marshal(service)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/networks/ip-services/%s", c.client.GetV1Url(), id)

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(serviceJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var s NetworkIPServiceResponse
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Delete removes an IP service by its ID
// IPServiceID: The ID of the IP service to delete
// Returns any error that occurred during deletion
func (c *NetworkIPServicesService) Delete(IPServiceID string) error {
	endpoint := fmt.Sprintf("%s/networks/ip-services/%s", c.client.GetV1Url(), IPServiceID)
	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	if err != nil {
		return err
	}
	return nil
}
