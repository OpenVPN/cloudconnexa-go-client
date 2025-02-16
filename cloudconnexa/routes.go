package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type Route struct {
	ID              string `json:"id,omitempty"`
	Type            string `json:"type,omitempty"`
	Subnet          string `json:"subnet,omitempty"`
	Domain          string `json:"domain,omitempty"`
	Description     string `json:"description,omitempty"`
	ParentRouteID   string `json:"parentRouteId,omitempty"`
	NetworkItemID   string `json:"networkItemId,omitempty"`
	AllowEmbeddedIP bool   `json:"allowEmbeddedIp,omitempty"`
}

type RoutePageResponse struct {
	Success          bool    `json:"success"`
	Content          []Route `json:"content"`
	TotalElements    int     `json:"totalElements"`
	TotalPages       int     `json:"totalPages"`
	NumberOfElements int     `json:"numberOfElements"`
	Page             int     `json:"page"`
	Size             int     `json:"size"`
}

type RoutesService service

func (c *RoutesService) GetByPage(networkID string, page int, size int) (RoutePageResponse, error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/networks/routes?networkId=%s&page=%d&size=%d", c.client.GetV1Url(), networkID, page, size), nil)
	if err != nil {
		return RoutePageResponse{}, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return RoutePageResponse{}, err
	}

	var response RoutePageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return RoutePageResponse{}, err
	}
	return response, nil
}

func (c *RoutesService) List(networkID string) ([]Route, error) {
	var allRoutes []Route
	pageSize := 10
	page := 0

	for {
		response, err := c.GetByPage(networkID, page, pageSize)
		if err != nil {
			return nil, err
		}

		allRoutes = append(allRoutes, response.Content...)

		if page >= response.TotalPages {
			break
		}
		page++
	}
	return allRoutes, nil
}

func (c *RoutesService) GetNetworkRoute(networkID string, routeID string) (*Route, error) {
	routes, err := c.List(networkID)
	if err != nil {
		return nil, err
	}

	for _, r := range routes {
		if r.ID == routeID {
			return &r, nil
		}
	}
	return nil, nil
}

func (c *RoutesService) Get(routeID string) (*Route, error) {
	networks, err := c.client.Networks.List()
	if err != nil {
		return nil, err
	}

	for _, n := range networks {
		if err != nil {
			continue
		}
		for _, r := range n.Routes {
			if r.ID == routeID {
				r.NetworkItemID = n.ID
				return &r, nil
			}
		}
	}
	return nil, nil
}

func (c *RoutesService) Create(networkID string, route Route) (*Route, error) {
	type newRoute struct {
		Description string `json:"description"`
		Value       string `json:"value"`
	}
	routeToCreate := newRoute{
		Description: route.Description,
		Value:       route.Subnet,
	}
	routeJSON, err := json.Marshal(routeToCreate)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("%s/networks/routes?networkId=%s", c.client.GetV1Url(), networkID),
		bytes.NewBuffer(routeJSON),
	)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var r Route
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (c *RoutesService) Update(route Route) error {
	type updatedRoute struct {
		Description string `json:"description"`
		Value       string `json:"value"`
	}
	routeToUpdate := updatedRoute{
		Description: route.Description,
		Value:       route.Subnet,
	}

	routeJSON, err := json.Marshal(routeToUpdate)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		http.MethodPut,
		fmt.Sprintf("%s/networks/routes/%s", c.client.GetV1Url(), route.ID),
		bytes.NewBuffer(routeJSON),
	)
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}

func (c *RoutesService) Delete(id string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/networks/routes/%s", c.client.GetV1Url(), id), nil)
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}
