package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type NetworkApplicationsService service

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

func (c *NetworkApplicationsService) Create(application *Application) (*ApplicationResponse, error) {
	applicationJson, err := json.Marshal(application)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/networks/applications?networkId=%s", c.client.GetV1Url(), application.NetworkItemId)

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(applicationJson))
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

func (c *NetworkApplicationsService) Update(id string, application *Application) (*ApplicationResponse, error) {
	applicationJson, err := json.Marshal(application)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/networks/applications/%s", c.client.GetV1Url(), id)

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(applicationJson))
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
