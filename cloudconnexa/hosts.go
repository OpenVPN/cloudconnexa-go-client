package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Host struct {
	ID             string          `json:"id,omitempty"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	Domain         string          `json:"domain,omitempty"`
	InternetAccess string          `json:"internetAccess"`
	SystemSubnets  []string        `json:"systemSubnets"`
	Connectors     []HostConnector `json:"connectors"`
}

type HostPageResponse struct {
	Content          []Host `json:"content"`
	NumberOfElements int    `json:"numberOfElements"`
	Page             int    `json:"page"`
	Size             int    `json:"size"`
	Success          bool   `json:"success"`
	TotalElements    int    `json:"totalElements"`
	TotalPages       int    `json:"totalPages"`
}

type HostsService service

func (c *HostsService) GetHostsByPage(page int, size int) (HostPageResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/hosts?page=%d&size=%d", c.client.GetV1Url(), page, size), nil)
	if err != nil {
		return HostPageResponse{}, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return HostPageResponse{}, err
	}

	var response HostPageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return HostPageResponse{}, err
	}
	return response, nil
}

func (c *HostsService) List() ([]Host, error) {
	var allHosts []Host
	pageSize := 10
	page := 0

	for {
		response, err := c.GetHostsByPage(page, pageSize)
		if err != nil {
			return nil, err
		}

		allHosts = append(allHosts, response.Content...)

		if page >= response.TotalPages {
			break
		}
		page++
	}
	return allHosts, nil
}

func (c *HostsService) Get(id string) (*Host, error) {
	endpoint := fmt.Sprintf("%s/hosts/%s", c.client.GetV1Url(), id)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var host Host
	err = json.Unmarshal(body, &host)
	if err != nil {
		return nil, err
	}
	return &host, nil
}

func (c *HostsService) Create(host Host) (*Host, error) {
	hostJSON, err := json.Marshal(host)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/hosts", c.client.GetV1Url()), bytes.NewBuffer(hostJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var h Host
	err = json.Unmarshal(body, &h)
	if err != nil {
		return nil, err
	}
	return &h, nil
}

func (c *HostsService) Update(host Host) error {
	hostJSON, err := json.Marshal(host)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/hosts/%s", c.client.GetV1Url(), host.ID), bytes.NewBuffer(hostJSON))
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}

func (c *HostsService) Delete(hostID string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/hosts/%s", c.client.GetV1Url(), hostID), nil)
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}
