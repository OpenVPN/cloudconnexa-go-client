package cloudconnexa

import (
	"bytes"
	"encoding/json"
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

// IPServiceResponse represents the response structure for IP service operations,
// extending the base IPService with additional route information.
type IPServiceResponse struct {
	IPService
	Routes []*Route `json:"routes"`
}

// IPServicePageResponse represents a paginated response from the CloudConnexa API
// containing a list of IP services and pagination metadata.
type IPServicePageResponse struct {
	Content          []IPServiceResponse `json:"content"`
	NumberOfElements int                 `json:"numberOfElements"`
	Page             int                 `json:"page"`
	Size             int                 `json:"size"`
	Success          bool                `json:"success"`
	TotalElements    int                 `json:"totalElements"`
	TotalPages       int                 `json:"totalPages"`
}

// HostIPServicesService provides methods for managing IP services.
type HostIPServicesService service

// GetIPByPage retrieves IP services using pagination.
func (c *HostIPServicesService) GetIPByPage(page int, pageSize int) (IPServicePageResponse, error) {
	endpoint := fmt.Sprintf("%s/hosts/ip-services?page=%d&size=%d", c.client.GetV1Url(), page, pageSize)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return IPServicePageResponse{}, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return IPServicePageResponse{}, err
	}

	var response IPServicePageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return IPServicePageResponse{}, err
	}
	return response, nil
}

// List retrieves all IP services by paginating through all available pages.
func (c *HostIPServicesService) List() ([]IPServiceResponse, error) {
	var allIPServices []IPServiceResponse
	page := 0
	pageSize := 10

	for {
		response, err := c.GetIPByPage(page, pageSize)
		if err != nil {
			return nil, err
		}

		allIPServices = append(allIPServices, response.Content...)
		if page >= response.TotalPages {
			break
		}
		page++
	}
	return allIPServices, nil
}

// Get retrieves a specific IP service by its ID.
func (c *HostIPServicesService) Get(id string) (*IPServiceResponse, error) {
	endpoint := fmt.Sprintf("%s/hosts/ip-services/%s", c.client.GetV1Url(), id)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var service IPServiceResponse
	err = json.Unmarshal(body, &service)
	if err != nil {
		return nil, err
	}
	return &service, nil
}

// Create creates a new IP service.
func (c *HostIPServicesService) Create(ipService *IPService) (*IPServiceResponse, error) {
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

	var s IPServiceResponse
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Update updates an existing IP service by its ID.
func (c *HostIPServicesService) Update(id string, service *IPService) (*IPServiceResponse, error) {
	serviceJSON, err := json.Marshal(service)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/hosts/ip-services/%s", c.client.GetV1Url(), id)

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(serviceJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var s IPServiceResponse
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Delete removes an IP service by its ID.
func (c *HostIPServicesService) Delete(ipServiceID string) error {
	endpoint := fmt.Sprintf("%s/hosts/ip-services/%s", c.client.GetV1Url(), ipServiceID)
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
