package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Range struct {
	LowerValue int `json:"lowerValue"`
	UpperValue int `json:"upperValue"`
	Value      int `json:"value,omitempty"`
}

type CustomIPServiceType struct {
	IcmpType []Range `json:"icmpType"`
	Port     []Range `json:"port"`
	Protocol string  `json:"protocol"`
}

type IPServiceRoute struct {
	Description string `json:"description"`
	Value       string `json:"value"`
}

type IPServiceConfig struct {
	CustomServiceTypes []*CustomIPServiceType `json:"customServiceTypes"`
	ServiceTypes       []string               `json:"serviceTypes"`
}

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

type IPServiceResponse struct {
	IPService
	Routes []*Route `json:"routes"`
}

type IPServicePageResponse struct {
	Content          []IPServiceResponse `json:"content"`
	NumberOfElements int                 `json:"numberOfElements"`
	Page             int                 `json:"page"`
	Size             int                 `json:"size"`
	Success          bool                `json:"success"`
	TotalElements    int                 `json:"totalElements"`
	TotalPages       int                 `json:"totalPages"`
}

type HostIPServicesService service

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
