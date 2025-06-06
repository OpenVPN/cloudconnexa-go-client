package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// LocationContext represents a location context in CloudConnexa with its associated checks and user groups.
type LocationContext struct {
	ID            string        `json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description,omitempty"`
	UserGroupsIDs []string      `json:"userGroupsIds"`
	IPCheck       *IPCheck      `json:"ipCheck,omitempty"`
	CountryCheck  *CountryCheck `json:"countryCheck,omitempty"`
	DefaultCheck  *DefaultCheck `json:"defaultCheck"`
}

// IPCheck represents the IP-based access control configuration.
type IPCheck struct {
	Allowed bool `json:"allowed"`
	Ips     []IP `json:"ips"`
}

// CountryCheck represents the country-based access control configuration.
type CountryCheck struct {
	Allowed   bool     `json:"allowed"`
	Countries []string `json:"countries"`
}

// DefaultCheck represents the default access control configuration.
type DefaultCheck struct {
	Allowed bool `json:"allowed"`
}

// IP represents an IP address with its description.
type IP struct {
	IP          string `json:"ip"`
	Description string `json:"description"`
}

// LocationContextPageResponse represents a paginated response from the CloudConnexa API
// containing a list of location contexts and pagination metadata.
type LocationContextPageResponse struct {
	Content          []LocationContext `json:"content"`
	NumberOfElements int               `json:"numberOfElements"`
	Page             int               `json:"page"`
	Size             int               `json:"size"`
	Success          bool              `json:"success"`
	TotalElements    int               `json:"totalElements"`
	TotalPages       int               `json:"totalPages"`
}

// LocationContextsService provides methods for managing location contexts.
type LocationContextsService service

// GetLocationContextByPage retrieves location contexts using pagination.
func (c *LocationContextsService) GetLocationContextByPage(page int, pageSize int) (LocationContextPageResponse, error) {
	endpoint := fmt.Sprintf("%s/location-contexts?page=%d&size=%d", c.client.GetV1Url(), page, pageSize)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return LocationContextPageResponse{}, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return LocationContextPageResponse{}, err
	}

	var response LocationContextPageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return LocationContextPageResponse{}, err
	}
	return response, nil
}

// List retrieves all location contexts by paginating through all available pages.
func (c *LocationContextsService) List() ([]LocationContext, error) {
	var allLocationContexts []LocationContext
	page := 0
	pageSize := 10

	for {
		response, err := c.GetLocationContextByPage(page, pageSize)
		if err != nil {
			return nil, err
		}

		allLocationContexts = append(allLocationContexts, response.Content...)
		if page >= response.TotalPages {
			break
		}
		page++
	}
	return allLocationContexts, nil
}

// Get retrieves a specific location context by its ID.
func (c *LocationContextsService) Get(id string) (*LocationContext, error) {
	endpoint := fmt.Sprintf("%s/location-contexts/%s", c.client.GetV1Url(), id)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var locationContext LocationContext
	err = json.Unmarshal(body, &locationContext)
	if err != nil {
		return nil, err
	}
	return &locationContext, nil
}

// Create creates a new location context.
func (c *LocationContextsService) Create(locationContext *LocationContext) (*LocationContext, error) {
	locationContextJSON, err := json.Marshal(locationContext)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/location-contexts/", c.client.GetV1Url())
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(locationContextJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var s LocationContext
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Update updates an existing location context by its ID.
func (c *LocationContextsService) Update(id string, locationContext *LocationContext) (*LocationContext, error) {
	locationContextJSON, err := json.Marshal(locationContext)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/location-contexts/%s", c.client.GetV1Url(), id)
	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(locationContextJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var s LocationContext
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Delete removes a location context by its ID.
func (c *LocationContextsService) Delete(id string) error {
	endpoint := fmt.Sprintf("%s/location-contexts/%s", c.client.GetV1Url(), id)
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
