package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type NetworkIPServicesService service

func (c *NetworkIPServicesService) GetIPByPage(page int, pageSize int) (IPServicePageResponse, error) {
	endpoint := fmt.Sprintf("%s/networks/ip-services?page=%d&size=%d", c.client.GetV1Url(), page, pageSize)
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

func (c *NetworkIPServicesService) List() ([]IPServiceResponse, error) {
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

func (c *NetworkIPServicesService) Get(id string) (*IPServiceResponse, error) {
	endpoint := fmt.Sprintf("%s/networks/ip-services/%s", c.client.GetV1Url(), id)
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

func (c *NetworkIPServicesService) Create(ipService *IPService) (*IPServiceResponse, error) {
	ipServiceJSON, err := json.Marshal(ipService)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/networks/ip-services?networkId=%s", c.client.GetV1Url(), ipService.NetworkItemID)

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

func (c *NetworkIPServicesService) Update(id string, service *IPService) (*IPServiceResponse, error) {
	serviceJSON, err := json.Marshal(service)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/networks/ip-services/%s", c.client.GetV1Url(), id)

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

func (c *NetworkIPServicesService) Delete(IPServiceID string) error {
	endpoint := fmt.Sprintf("%s/networks/ip-services/%s", c.client.GetV1Url(), IPServiceID)
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
