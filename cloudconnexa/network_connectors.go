package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

type NetworkConnector struct {
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

type NetworkConnectorPageResponse struct {
	Content          []NetworkConnector `json:"content"`
	NumberOfElements int                `json:"numberOfElements"`
	Page             int                `json:"page"`
	Size             int                `json:"size"`
	Success          bool               `json:"success"`
	TotalElements    int                `json:"totalElements"`
	TotalPages       int                `json:"totalPages"`
}

type NetworkConnectorsService service

func (c *NetworkConnectorsService) GetByPage(page int, pageSize int) (NetworkConnectorPageResponse, error) {
	return c.GetByPageAndNetworkID(page, pageSize, "")
}

func (c *NetworkConnectorsService) GetByPageAndNetworkID(page int, pageSize int, networkID string) (NetworkConnectorPageResponse, error) {
	params := url.Values{}
	params.Add("page", strconv.Itoa(page))
	params.Add("size", strconv.Itoa(pageSize))
	if networkID != "" {
		params.Add("networkId", networkID)
	}

	endpoint := fmt.Sprintf("%s/networks/connectors?%s", c.client.GetV1Url(), params.Encode())
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return NetworkConnectorPageResponse{}, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return NetworkConnectorPageResponse{}, err
	}

	var response NetworkConnectorPageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return NetworkConnectorPageResponse{}, err
	}
	return response, nil
}

func (c *NetworkConnectorsService) Update(connector NetworkConnector) (*NetworkConnector, error) {
	connectorJSON, err := json.Marshal(connector)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/networks/connectors/%s", c.client.GetV1Url(), connector.ID), bytes.NewBuffer(connectorJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var conn NetworkConnector
	err = json.Unmarshal(body, &conn)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func (c *NetworkConnectorsService) List() ([]NetworkConnector, error) {
	var allConnectors []NetworkConnector
	page := 0
	pageSize := 10

	for {
		response, err := c.GetByPage(page, pageSize)
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

func (c *NetworkConnectorsService) ListByNetworkID(networkID string) ([]NetworkConnector, error) {
	var allConnectors []NetworkConnector
	page := 0
	pageSize := 10

	for {
		response, err := c.GetByPageAndNetworkID(page, pageSize, networkID)
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

func (c *NetworkConnectorsService) GetByID(id string) (*NetworkConnector, error) {
	endpoint := fmt.Sprintf("%s/networks/connectors/%s", c.client.GetV1Url(), id)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var connector NetworkConnector
	err = json.Unmarshal(body, &connector)
	if err != nil {
		return nil, err
	}
	return &connector, nil
}

func (c *NetworkConnectorsService) GetProfile(id string) (string, error) {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/networks/connectors/%s/profile", c.client.GetV1Url(), id), nil)
	if err != nil {
		return "", err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *NetworkConnectorsService) GetToken(id string) (string, error) {
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/networks/connectors/%s/profile/encrypt", c.client.GetV1Url(), id), nil)
	if err != nil {
		return "", err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *NetworkConnectorsService) Create(connector NetworkConnector, networkID string) (*NetworkConnector, error) {
	connectorJSON, err := json.Marshal(connector)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/networks/connectors?networkId=%s", c.client.GetV1Url(), networkID), bytes.NewBuffer(connectorJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var conn NetworkConnector
	err = json.Unmarshal(body, &conn)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func (c *NetworkConnectorsService) Delete(connectorID string, networkID string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/networks/connectors/%s?networkId=%s", c.client.GetV1Url(), connectorID, networkID), nil)
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}
