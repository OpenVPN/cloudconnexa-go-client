package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type LocationContext struct {
	Id            string        `json:"id"`
	Name          string        `json:"name"`
	Description   string        `json:"description,omitempty"`
	UserGroupsIds []string      `json:"userGroupsIds"`
	IpCheck       *IpCheck      `json:"ipCheck,omitempty"`
	CountryCheck  *CountryCheck `json:"countryCheck,omitempty"`
	DefaultCheck  *DefaultCheck `json:"defaultCheck"`
}

type IpCheck struct {
	Allowed bool `json:"allowed"`
	Ips     []Ip `json:"ips"`
}

type CountryCheck struct {
	Allowed   bool     `json:"allowed"`
	Countries []string `json:"countries"`
}

type DefaultCheck struct {
	Allowed bool `json:"allowed"`
}

type Ip struct {
	Ip          string `json:"ip"`
	Description string `json:"description"`
}

type LocationContextPageResponse struct {
	Content          []LocationContext `json:"content"`
	NumberOfElements int               `json:"numberOfElements"`
	Page             int               `json:"page"`
	Size             int               `json:"size"`
	Success          bool              `json:"success"`
	TotalElements    int               `json:"totalElements"`
	TotalPages       int               `json:"totalPages"`
}

type LocationContextsService service

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

func (c *LocationContextsService) Create(locationContext *LocationContext) (*LocationContext, error) {
	locationContextJson, err := json.Marshal(locationContext)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/location-contexts/", c.client.GetV1Url())
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(locationContextJson))
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

func (c *LocationContextsService) Update(id string, locationContext *LocationContext) (*LocationContext, error) {
	locationContextJson, err := json.Marshal(locationContext)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/location-contexts/%s", c.client.GetV1Url(), id)
	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(locationContextJson))
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
