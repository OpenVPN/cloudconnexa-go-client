package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Network struct {
	Connectors     []NetworkConnector `json:"connectors"`
	Description    string             `json:"description"`
	Egress         bool               `json:"egress"`
	Id             string             `json:"id"`
	InternetAccess string             `json:"internetAccess"`
	Name           string             `json:"name"`
	Routes         []Route            `json:"routes"`
	SystemSubnets  []string           `json:"systemSubnets"`
}

type NetworkPageResponse struct {
	Content          []Network `json:"content"`
	NumberOfElements int       `json:"numberOfElements"`
	Page             int       `json:"page"`
	Size             int       `json:"size"`
	Success          bool      `json:"success"`
	TotalElements    int       `json:"totalElements"`
	TotalPages       int       `json:"totalPages"`
}

type NetworksService service

func (c *NetworksService) GetByPage(page int, size int) (NetworkPageResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/networks?page=%d&size=%d", c.client.GetV1Url(), page, size), nil)
	if err != nil {
		return NetworkPageResponse{}, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return NetworkPageResponse{}, err
	}

	var response NetworkPageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return NetworkPageResponse{}, err
	}

	return response, nil
}

func (c *NetworksService) List() ([]Network, error) {
	var allNetworks []Network
	pageSize := 10
	page := 0

	for {
		response, err := c.GetByPage(page, pageSize)
		if err != nil {
			return nil, err
		}

		allNetworks = append(allNetworks, response.Content...)

		if page >= response.TotalPages {
			break
		}
		page++
	}
	return allNetworks, nil
}

func (c *NetworksService) Get(id string) (*Network, error) {
	endpoint := fmt.Sprintf("%s/networks/%s", c.client.GetV1Url(), id)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var network Network
	err = json.Unmarshal(body, &network)
	if err != nil {
		return nil, err
	}
	return &network, nil
}

func (c *NetworksService) Create(network Network) (*Network, error) {
	networkJson, err := json.Marshal(network)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/networks", c.client.GetV1Url()), bytes.NewBuffer(networkJson))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var n Network
	err = json.Unmarshal(body, &n)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (c *NetworksService) Update(network Network) error {
	networkJson, err := json.Marshal(network)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/networks/%s", c.client.GetV1Url(), network.Id), bytes.NewBuffer(networkJson))
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}

func (c *NetworksService) Delete(networkId string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/networks/%s", c.client.GetV1Url(), networkId), nil)
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}
