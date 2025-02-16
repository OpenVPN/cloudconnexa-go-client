package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type HostConnector struct {
	ID               string `json:"id,omitempty"`
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	NetworkItemID    string `json:"networkItemId"`
	NetworkItemType  string `json:"networkItemType"`
	VpnRegionID      string `json:"vpnRegionId"`
	IPv4Address      string `json:"ipV4Address"`
	IPv6Address      string `json:"ipV6Address"`
	Profile          string `json:"profile"`
	ConnectionStatus string `json:"connectionStatus"`
}

type HostConnectorPageResponse struct {
	Content          []HostConnector `json:"content"`
	NumberOfElements int             `json:"numberOfElements"`
	Page             int             `json:"page"`
	Size             int             `json:"size"`
	Success          bool            `json:"success"`
	TotalElements    int             `json:"totalElements"`
	TotalPages       int             `json:"totalPages"`
}

type HostConnectorsService service

func (c *HostConnectorsService) GetByPage(page int, pageSize int) (HostConnectorPageResponse, error) {
	return c.GetByPageAndHostID(page, pageSize, "")
}

func (c *HostConnectorsService) GetByPageAndHostID(page int, pageSize int, hostID string) (HostConnectorPageResponse, error) {
	params := url.Values{}
	params.Add("page", strconv.Itoa(page))
	params.Add("size", strconv.Itoa(pageSize))
	if hostID != "" {
		params.Add("hostId", hostID)
	}

	endpoint := fmt.Sprintf("%s/hosts/connectors?%s", c.client.GetV1Url(), params.Encode())
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return HostConnectorPageResponse{}, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return HostConnectorPageResponse{}, err
	}

	var response HostConnectorPageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return HostConnectorPageResponse{}, err
	}
	return response, nil
}

func (c *HostConnectorsService) Update(connector HostConnector) (*HostConnector, error) {
	connectorJSON, err := json.Marshal(connector)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/hosts/connectors/%s", c.client.GetV1Url(), connector.ID), bytes.NewBuffer(connectorJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var conn HostConnector
	err = json.Unmarshal(body, &conn)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func (c *HostConnectorsService) List() ([]HostConnector, error) {
	return c.ListByHostID("")
}

func (c *HostConnectorsService) ListByHostID(hostID string) ([]HostConnector, error) {
	var allConnectors []HostConnector
	page := 0
	pageSize := 10

	for {
		response, err := c.GetByPageAndHostID(page, pageSize, hostID)
		if err != nil {
			return nil, err
		}

		allConnectors = append(allConnectors, response.Content...)

		if page >= response.TotalPages {
			break
		}
		page++
	}
	return allConnectors, nil
}

func (c *HostConnectorsService) GetByID(id string) (*HostConnector, error) {
	endpoint := fmt.Sprintf("%s/hosts/connectors/%s", c.client.GetV1Url(), id)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var connector HostConnector
	err = json.Unmarshal(body, &connector)
	if err != nil {
		return nil, err
	}
	return &connector, nil
}

func (c *HostConnectorsService) GetProfile(id string) (string, error) {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/hosts/connectors/%s/profile", c.client.GetV1Url(), id), nil)
	if err != nil {
		return "", err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *HostConnectorsService) GetToken(id string) (string, error) {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/hosts/connectors/%s/profile/encrypt", c.client.GetV1Url(), id), nil)
	if err != nil {
		return "", err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *HostConnectorsService) Create(connector HostConnector, hostID string) (*HostConnector, error) {
	connectorJSON, err := json.Marshal(connector)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/hosts/connectors?hostId=%s", c.client.GetV1Url(), hostID), bytes.NewBuffer(connectorJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var conn HostConnector
	err = json.Unmarshal(body, &conn)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func (c *HostConnectorsService) Delete(connectorID string, hostID string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/hosts/connectors/%s?hostId=%s", c.client.GetV1Url(), connectorID, hostID), nil)
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}
