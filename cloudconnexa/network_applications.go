package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// NetworkApplicationRoute represents a route configuration for a network application.
// Mirrors the API NetworkApplicationRouteRequest schema. ExactMatch is supported
// only for network application routes; the host application equivalent
// (ApplicationRoute) does not include it.
type NetworkApplicationRoute struct {
	Value           string `json:"value"`
	AllowEmbeddedIP bool   `json:"allowEmbeddedIp"`
	ExactMatch      bool   `json:"exactMatch,omitempty"`
}

// NetworkApplicationDomainRoute represents a domain route returned by the network
// application endpoints. Mirrors the API NetworkDomainRouteResponse schema, which
// — unlike HostDomainRouteResponse — carries the ExactMatch flag.
type NetworkApplicationDomainRoute struct {
	ID              string `json:"id,omitempty"`
	Type            string `json:"type,omitempty"`
	Domain          string `json:"domain,omitempty"`
	AllowEmbeddedIP bool   `json:"allowEmbeddedIp,omitempty"`
	ExactMatch      bool   `json:"exactMatch,omitempty"`
}

// NetworkApplication represents a network application request body.
// It is structurally similar to Application but uses NetworkApplicationRoute
// for its routes so that ExactMatch is honored.
type NetworkApplication struct {
	Name            string                     `json:"name"`
	Description     string                     `json:"description"`
	NetworkItemType string                     `json:"networkItemType"`
	NetworkItemID   string                     `json:"networkItemId"`
	ID              string                     `json:"id"`
	Routes          []*NetworkApplicationRoute `json:"routes"`
	Config          *ApplicationConfig         `json:"config"`
}

// NetworkApplicationResponse represents the response payload for network application
// operations. Routes carry the network-specific NetworkApplicationDomainRoute,
// which exposes ExactMatch.
type NetworkApplicationResponse struct {
	NetworkApplication
	Routes []*NetworkApplicationDomainRoute `json:"routes"`
}

// NetworkApplicationPageResponse represents a paginated response of network applications.
type NetworkApplicationPageResponse struct {
	Content          []NetworkApplicationResponse `json:"content"`
	NumberOfElements int                          `json:"numberOfElements"`
	Page             int                          `json:"page"`
	Size             int                          `json:"size"`
	Success          bool                         `json:"success"`
	TotalElements    int                          `json:"totalElements"`
	TotalPages       int                          `json:"totalPages"`
}

// NetworkApplicationsService provides methods for managing network applications.
type NetworkApplicationsService service

// GetApplicationsByPage retrieves network applications using pagination.
func (c *NetworkApplicationsService) GetApplicationsByPage(page int, pageSize int) (NetworkApplicationPageResponse, error) {
	endpoint := fmt.Sprintf("%s/networks/applications?page=%d&size=%d", c.client.GetV1Url(), page, pageSize)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return NetworkApplicationPageResponse{}, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return NetworkApplicationPageResponse{}, err
	}

	var response NetworkApplicationPageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return NetworkApplicationPageResponse{}, err
	}
	return response, nil
}

// List retrieves all network applications by paginating through all available pages.
func (c *NetworkApplicationsService) List() ([]NetworkApplicationResponse, error) {
	var allApplications []NetworkApplicationResponse
	page := 0

	for {
		response, err := c.GetApplicationsByPage(page, defaultPageSize)
		if err != nil {
			return nil, err
		}

		allApplications = append(allApplications, response.Content...)

		page++
		if page >= response.TotalPages {
			break
		}
	}
	return allApplications, nil
}

// Get retrieves a specific network application by its ID.
func (c *NetworkApplicationsService) Get(id string) (*NetworkApplicationResponse, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	endpoint := buildURL(c.client.GetV1Url(), "networks", "applications", id)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var application NetworkApplicationResponse
	err = json.Unmarshal(body, &application)
	if err != nil {
		return nil, err
	}
	return &application, nil
}

// GetByName retrieves a network application by its name
// name: The name of the network application to retrieve
// Returns the network application and any error that occurred
func (c *NetworkApplicationsService) GetByName(name string) (*NetworkApplicationResponse, error) {
	items, err := c.List()
	if err != nil {
		return nil, err
	}

	filtered := make([]NetworkApplicationResponse, 0)
	for _, item := range items {
		if item.Name == name {
			filtered = append(filtered, item)
		}
	}
	if len(filtered) > 1 {
		return nil, errors.New("different network applications found with name: " + name)
	}
	if len(filtered) == 1 {
		return &filtered[0], nil
	}
	return nil, errors.New("network application not found")
}

// Create creates a new network application.
func (c *NetworkApplicationsService) Create(application *NetworkApplication) (*NetworkApplicationResponse, error) {
	applicationJSON, err := json.Marshal(application)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/networks/applications?networkId=%s", c.client.GetV1Url(), application.NetworkItemID)

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(applicationJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var s NetworkApplicationResponse
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Update updates an existing network application by its ID.
func (c *NetworkApplicationsService) Update(id string, application *NetworkApplication) (*NetworkApplicationResponse, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	applicationJSON, err := json.Marshal(application)
	if err != nil {
		return nil, err
	}

	endpoint := buildURL(c.client.GetV1Url(), "networks", "applications", id)

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(applicationJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var s NetworkApplicationResponse
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Delete removes a network application by its ID.
func (c *NetworkApplicationsService) Delete(id string) error {
	if err := validateID(id); err != nil {
		return err
	}
	endpoint := buildURL(c.client.GetV1Url(), "networks", "applications", id)
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
