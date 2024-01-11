package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Range struct {
	LowerValue int `json:"lowerValue"`
	UpperValue int `json:"upperValue"`
	Value      int `json:"value"`
}

type CustomIPServiceType struct {
	IcmpType []Range `json:"icmpType"`
	Port     []Range `json:"port"`
	Protocol string  `json:"protocol"`
}

type IPServiceConfig struct {
	CustomServiceTypes []*CustomIPServiceType `json:"customServiceTypes"`
	ServiceTypes       []string               `json:"serviceTypes"`
}

type IPService struct {
	Name            string           `json:"name"`
	Description     string           `json:"description"`
	NetworkItemType string           `json:"networkItemType"`
	NetworkItemId   string           `json:"networkItemId"`
	Id              string           `json:"id"`
	Type            string           `json:"type"`
	Routes          []*Route         `json:"routes"`
	Config          *IPServiceConfig `json:"config"`
}

type IPServicePageResponse struct {
	Content          []IPService `json:"content"`
	NumberOfElements int         `json:"numberOfElements"`
	Page             int         `json:"page"`
	Size             int         `json:"size"`
	Success          bool        `json:"success"`
	TotalElements    int         `json:"totalElements"`
	TotalPages       int         `json:"totalPages"`
}

func (c *Client) GetIPServicesByPage(page int, pageSize int) (IPServicePageResponse, error) {
	endpoint := fmt.Sprintf("%s/api/beta/ip-services/page?page=%d&size=%d", c.BaseURL, page, pageSize)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return IPServicePageResponse{}, err
	}

	body, err := c.DoRequest(req)
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

func (c *Client) GetAllIPServices() ([]IPService, error) {
	var allIPServices []IPService
	page := 1
	pageSize := 10

	for {
		response, err := c.GetIPServicesByPage(page, pageSize)
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

func (c *Client) GetServiceIPByID(serviceID string) (*IPService, error) {
	endpoint := fmt.Sprintf("%s/api/beta/ip-services/single?serviceId=%s", c.BaseURL, serviceID)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var service IPService
	err = json.Unmarshal(body, &service)
	if err != nil {
		return nil, err
	}
	return &service, nil
}

func (c *Client) CreateIPService(ipService *IPService) (*IPService, error) {
	ipServiceJson, err := json.Marshal(ipService)
	if err != nil {
		return nil, err
	}

	params := networkUrlParams(ipService.NetworkItemType, ipService.NetworkItemId)
	endpoint := fmt.Sprintf("%s/api/beta/ip-services?%s", c.BaseURL, params.Encode())

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(ipServiceJson))
	if err != nil {
		return nil, err
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var s IPService
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *Client) UpdateIPService(id string, service *IPService) (*IPService, error) {
	serviceJson, err := json.Marshal(service)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/api/beta/ip-services/%s?%s", c.BaseURL, id)

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(serviceJson))
	if err != nil {
		return nil, err
	}

	body, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var s IPService
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *Client) DeleteIPService(ipServiceId string) error {
	endpoint := fmt.Sprintf("%s/api/beta/ip-services/%s?%s", c.BaseURL, ipServiceId)
	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	_, err = c.DoRequest(req)
	if err != nil {
		return err
	}
	return nil
}

func networkUrlParams(networkItemType string, networkItemId string) url.Values {
	params := url.Values{}
	params.Add("networkItemId", networkItemId)
	params.Add("networkItemType", networkItemType)
	return params
}
