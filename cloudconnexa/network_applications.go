package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// NetworkApplicationsService provides methods for managing network applications.
type NetworkApplicationsService service

// GetApplicationsByPage retrieves network applications using pagination.
func (c *NetworkApplicationsService) GetApplicationsByPage(page int, pageSize int) (ApplicationPageResponse, error) {
	endpoint := fmt.Sprintf("%s/networks/applications?page=%d&size=%d", c.client.GetV1Url(), page, pageSize)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return ApplicationPageResponse{}, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return ApplicationPageResponse{}, err
	}

	var response ApplicationPageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return ApplicationPageResponse{}, err
	}
	return response, nil
}

// List retrieves all network applications by paginating through all available pages.
func (c *NetworkApplicationsService) List() ([]ApplicationResponse, error) {
	var allApplications []ApplicationResponse
	page := 0
	pageSize := 10

	for {
		response, err := c.GetApplicationsByPage(page, pageSize)
		if err != nil {
			return nil, err
		}

		allApplications = append(allApplications, response.Content...)
		if page >= response.TotalPages {
			break
		}
		page++
	}
	return allApplications, nil
}

// Get retrieves a specific network application by its ID.
func (c *NetworkApplicationsService) Get(id string) (*ApplicationResponse, error) {
	endpoint := fmt.Sprintf("%s/networks/applications/%s", c.client.GetV1Url(), id)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var application ApplicationResponse
	err = json.Unmarshal(body, &application)
	if err != nil {
		return nil, err
	}
	return &application, nil
}

// Create creates a new network application.
func (c *NetworkApplicationsService) Create(application *Application) (*ApplicationResponse, error) {
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

	var s ApplicationResponse
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Update updates an existing network application by its ID.
func (c *NetworkApplicationsService) Update(id string, application *Application) (*ApplicationResponse, error) {
	applicationJSON, err := json.Marshal(application)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/networks/applications/%s", c.client.GetV1Url(), id)

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(applicationJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var s ApplicationResponse
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Delete removes a network application by its ID.
func (c *NetworkApplicationsService) Delete(id string) error {
	endpoint := fmt.Sprintf("%s/networks/applications/%s", c.client.GetV1Url(), id)
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
