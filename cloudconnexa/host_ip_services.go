package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// Range represents a range of values with lower and upper bounds, or a single value.
type Range struct {
	LowerValue int `json:"lowerValue"`
	UpperValue int `json:"upperValue"`
	Value      int `json:"value,omitempty"`
}

// CustomIPServiceType represents a custom IP service type configuration with ICMP type, port ranges, and protocol.
type CustomIPServiceType struct {
	IcmpType []Range `json:"icmpType"`
	Port     []Range `json:"port"`
	Protocol string  `json:"protocol"`
}

// IPServiceRoute represents a route configuration for an IP service.
type IPServiceRoute struct {
	Description string `json:"description"`
	Value       string `json:"value"`
}

// IPServiceConfig represents the configuration for an IP service including custom service types and predefined service types.
type IPServiceConfig struct {
	CustomServiceTypes []*CustomIPServiceType `json:"customServiceTypes"`
	ServiceTypes       []string               `json:"serviceTypes"`
}

// IPService represents an IP service with its configuration and routing information.
type IPService struct {
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	NetworkItemType string            `json:"networkItemType"`
	NetworkItemID   string            `json:"networkItemId"`
	ID              string            `json:"id"`
	Type            string            `json:"type"`
	Routes          []*IPServiceRoute `json:"routes"`
	Config          *IPServiceConfig  `json:"config"`
}

// HostIPServiceResponse represents the response structure for IP service operations.
// Updated for API v1.1.0: Removed duplicate routing information to match the simplified DTO.
type HostIPServiceResponse struct {
	Name            string           `json:"name"`
	Description     string           `json:"description"`
	NetworkItemType string           `json:"networkItemType"`
	NetworkItemID   string           `json:"networkItemId"`
	ID              string           `json:"id"`
	Type            string           `json:"type"`
	Config          *IPServiceConfig `json:"config"`
}

// HostIPServicePageResponse represents a paginated response from the CloudConnexa API
// containing a list of IP services and pagination metadata.
type HostIPServicePageResponse struct {
	Content          []HostIPServiceResponse `json:"content"`
	NumberOfElements int                     `json:"numberOfElements"`
	Page             int                     `json:"page"`
	Size             int                     `json:"size"`
	Success          bool                    `json:"success"`
	TotalElements    int                     `json:"totalElements"`
	TotalPages       int                     `json:"totalPages"`
}

// HostIPServicesService provides methods for managing IP services.
type HostIPServicesService service

// GetIPByPage retrieves IP services using pagination.
func (c *HostIPServicesService) GetIPByPage(page int, pageSize int) (HostIPServicePageResponse, error) {
	endpoint := fmt.Sprintf("%s/hosts/ip-services?page=%d&size=%d", c.client.GetV1Url(), page, pageSize)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return HostIPServicePageResponse{}, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return HostIPServicePageResponse{}, err
	}

	var response HostIPServicePageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return HostIPServicePageResponse{}, err
	}
	return response, nil
}

// List retrieves all IP services by paginating through all available pages.
func (c *HostIPServicesService) List() ([]HostIPServiceResponse, error) {
	var allIPServices []HostIPServiceResponse
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

// Get retrieves a specific IP service by its ID.
func (c *HostIPServicesService) Get(id string) (*HostIPServiceResponse, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	endpoint := buildURL(c.client.GetV1Url(), "hosts", "ip-services", id)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var service HostIPServiceResponse
	err = json.Unmarshal(body, &service)
	if err != nil {
		return nil, err
	}
	return &service, nil
}

// GetByName retrieves an IP Service by its name
// name: The name of the IP Service to retrieve
// Returns the IP Service and any error that occurred
func (c *HostIPServicesService) GetByName(name string) (*HostIPServiceResponse, error) {
	items, err := c.List()
	if err != nil {
		return nil, err
	}

	filtered := make([]HostIPServiceResponse, 0)
	for _, item := range items {
		if item.Name == name {
			filtered = append(filtered, item)
		}
	}
	if len(filtered) > 1 {
		return nil, errors.New("different host IP services found with name: " + name)
	}
	if len(filtered) == 1 {
		return &filtered[0], nil
	}
	return nil, errors.New("host IP service not found")
}

// Create creates a new IP service.
func (c *HostIPServicesService) Create(ipService *IPService) (*HostIPServiceResponse, error) {
	ipServiceJSON, err := json.Marshal(ipService)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/hosts/ip-services?hostId=%s", c.client.GetV1Url(), ipService.NetworkItemID)

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(ipServiceJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var s HostIPServiceResponse
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Update updates an existing IP service by its ID.
func (c *HostIPServicesService) Update(id string, service *IPService) (*HostIPServiceResponse, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	serviceJSON, err := json.Marshal(service)
	if err != nil {
		return nil, err
	}

	endpoint := buildURL(c.client.GetV1Url(), "hosts", "ip-services", id)

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(serviceJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var s HostIPServiceResponse
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Delete removes an IP service by its ID.
func (c *HostIPServicesService) Delete(ipServiceID string) error {
	if err := validateID(ipServiceID); err != nil {
		return err
	}
	endpoint := buildURL(c.client.GetV1Url(), "hosts", "ip-services", ipServiceID)
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
