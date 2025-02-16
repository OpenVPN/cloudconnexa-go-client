package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type ApplicationRoute struct {
	Value           string `json:"value"`
	AllowEmbeddedIP bool   `json:"allowEmbeddedIp"`
}

type CustomApplicationType struct {
	IcmpType []Range `json:"icmpType"`
	Port     []Range `json:"port"`
	Protocol string  `json:"protocol"`
}

type ApplicationConfig struct {
	CustomServiceTypes []*CustomApplicationType `json:"customServiceTypes"`
	ServiceTypes       []string                 `json:"serviceTypes"`
}

type Application struct {
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	NetworkItemType string              `json:"networkItemType"`
	NetworkItemID   string              `json:"networkItemId"`
	ID              string              `json:"id"`
	Routes          []*ApplicationRoute `json:"routes"`
	Config          *ApplicationConfig  `json:"config"`
}

type ApplicationResponse struct {
	Application
	Routes []*Route `json:"routes"`
}

type ApplicationPageResponse struct {
	Content          []ApplicationResponse `json:"content"`
	NumberOfElements int                   `json:"numberOfElements"`
	Page             int                   `json:"page"`
	Size             int                   `json:"size"`
	Success          bool                  `json:"success"`
	TotalElements    int                   `json:"totalElements"`
	TotalPages       int                   `json:"totalPages"`
}

type HostApplicationsService service

func (c *HostApplicationsService) GetApplicationsByPage(page int, pageSize int) (ApplicationPageResponse, error) {
	endpoint := fmt.Sprintf("%s/hosts/applications?page=%d&size=%d", c.client.GetV1Url(), page, pageSize)
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

func (c *HostApplicationsService) List() ([]ApplicationResponse, error) {
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

func (c *HostApplicationsService) Get(id string) (*ApplicationResponse, error) {
	endpoint := fmt.Sprintf("%s/hosts/applications/%s", c.client.GetV1Url(), id)
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

func (c *HostApplicationsService) Create(application *Application) (*ApplicationResponse, error) {
	applicationJSON, err := json.Marshal(application)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/hosts/applications?hostId=%s", c.client.GetV1Url(), application.NetworkItemID)

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

func (c *HostApplicationsService) Update(id string, application *Application) (*ApplicationResponse, error) {
	applicationJSON, err := json.Marshal(application)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/hosts/applications/%s", c.client.GetV1Url(), id)

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

func (c *HostApplicationsService) Delete(id string) error {
	endpoint := fmt.Sprintf("%s/hosts/applications/%s", c.client.GetV1Url(), id)
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
