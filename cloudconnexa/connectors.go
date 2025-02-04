package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type ConnectionStatus string

type Connector struct {
	Description      string           `json:"description,omitempty"`
	Id               string           `json:"id,omitempty"`
	Name             string           `json:"name"`
	NetworkItemId    string           `json:"networkItemId"`
	NetworkItemType  string           `json:"networkItemType"`
	VpnRegionId      string           `json:"vpnRegionId"`
	IPv4Address      string           `json:"ipV4Address"`
	IPv6Address      string           `json:"ipV6Address"`
	Profile          string           `json:"profile"`
	ConnectionStatus ConnectionStatus `json:"connectionStatus"`
}

type ConnectorPageResponse struct {
	Content          []Connector `json:"content"`
	NumberOfElements int         `json:"numberOfElements"`
	Page             int         `json:"page"`
	Size             int         `json:"size"`
	Success          bool        `json:"success"`
	TotalElements    int         `json:"totalElements"`
	TotalPages       int         `json:"totalPages"`
}

type ConnectorsService service

func (c *ConnectorsService) GetByPage(page int, pageSize int, networkItemType string) (ConnectorPageResponse, error) {
	path, err := GetPath(networkItemType)
	if err != nil {
		return ConnectorPageResponse{}, err
	}
	endpoint := fmt.Sprintf("%s/%s/connectors?page=%d&size=%d", c.client.GetV1Url(), path, page, pageSize)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return ConnectorPageResponse{}, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return ConnectorPageResponse{}, err
	}

	var response ConnectorPageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return ConnectorPageResponse{}, err
	}
	return response, nil
}

func (c *ConnectorsService) Update(connector Connector) (*Connector, error) {
	path, err := GetPath(connector.NetworkItemType)
	if err != nil {
		return nil, err
	}
	connectorJson, err := json.Marshal(connector)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/%s/connectors/%s", c.client.GetV1Url(), path, connector.Id), bytes.NewBuffer(connectorJson))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var conn Connector
	err = json.Unmarshal(body, &conn)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func (c *ConnectorsService) List(networkItemType string) ([]Connector, error) {
	var allConnectors []Connector
	page := 0
	pageSize := 10

	for {
		response, err := c.GetByPage(page, pageSize, networkItemType)
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

func (c *ConnectorsService) GetByName(name string, networkItemType string) (*Connector, error) {
	connectors, err := c.List(networkItemType)
	if err != nil {
		return nil, err
	}

	for _, connector := range connectors {
		if connector.Name == name {
			return &connector, nil
		}
	}
	return nil, nil
}

func (c *ConnectorsService) GetByID(connectorID string, networkItemType string) (*Connector, error) {
	connectors, err := c.List(networkItemType)
	if err != nil {
		return nil, err
	}

	for _, connector := range connectors {
		if connector.Id == connectorID {
			return &connector, nil
		}
	}
	return nil, nil
}

func (c *ConnectorsService) GetByNetworkID(networkId string) ([]Connector, error) {
	connectors, err := c.List("NETWORK")
	if err != nil {
		return nil, err
	}

	var networkConnectors []Connector
	for _, connector := range connectors {
		if connector.NetworkItemId == networkId {
			networkConnectors = append(networkConnectors, connector)
		}
	}
	return networkConnectors, nil
}

func (c *ConnectorsService) GetProfile(id string, networkItemType string) (string, error) {
	path, err := GetPath(networkItemType)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s/connectors/%s/profile", c.client.GetV1Url(), path, id), nil)
	if err != nil {
		return "", err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *ConnectorsService) GetToken(id string, networkItemType string) (string, error) {
	path, err := GetPath(networkItemType)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s/connectors/%s/profile/encrypt", c.client.GetV1Url(), path, id), nil)
	if err != nil {
		return "", err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *ConnectorsService) Create(connector Connector, networkItemId string) (*Connector, error) {
	path, err := GetPath(connector.NetworkItemType)
	if err != nil {
		return nil, err
	}
	connectorJson, err := json.Marshal(connector)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/%s/connectors?networkItemId=%s&networkItemType=%s", c.client.GetV1Url(), path, networkItemId, connector.NetworkItemType), bytes.NewBuffer(connectorJson))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var conn Connector
	err = json.Unmarshal(body, &conn)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func (c *ConnectorsService) Delete(connectorId string, networkItemId string, networkItemType string) error {
	path, err := GetPath(networkItemType)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/%s/connectors/%s?networkItemId=%s&networkItemType=%s", c.client.GetV1Url(), path, connectorId, networkItemId, networkItemType), nil)
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}

func GetPath(networkItemType string) (string, error) {
	if networkItemType == "NETWORK" {
		return "networks", nil
	}

	if networkItemType == "HOST" {
		return "hosts", nil
	}

	return "undefined", errors.New("unknown network item type")
}
