package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Route represents a network route configuration
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

// RoutePageResponse represents a paginated response of network routes
type RoutePageResponse struct {
	Success          bool    `json:"success"`
	Content          []Route `json:"content"`
	TotalElements    int     `json:"totalElements"`
	TotalPages       int     `json:"totalPages"`
	NumberOfElements int     `json:"numberOfElements"`
	Page             int     `json:"page"`
	Size             int     `json:"size"`
}

// RoutesService provides methods for managing network routes
type RoutesService service

// GetByPage retrieves network routes using pagination
// networkID: The ID of the network to get routes for
// page: The page number to retrieve
// size: The number of items per page
// Returns a page of routes and any error that occurred
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

// List retrieves all routes for a network by paginating through all available pages
// networkID: The ID of the network to get routes for
// Returns a slice of routes and any error that occurred
func (c *RoutesService) List(networkID string) ([]Route, error) {
	var allRoutes []Route
	page := 0

	for {
		response, err := c.GetByPage(networkID, page, defaultPageSize)
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

// GetNetworkRoute retrieves a specific route from a network
// networkID: The ID of the network containing the route
// routeID: The ID of the route to retrieve
// Returns the route and any error that occurred
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

// Get retrieves a specific route by searching through all networks
// routeID: The ID of the route to retrieve
// Returns the route and any error that occurred
func (c *RoutesService) Get(routeID string) (*Route, error) {
	networks, err := c.client.Networks.List()
	if err != nil {
		return nil, err
	}

	for _, n := range networks {
		for _, r := range n.Routes {
			if r.ID == routeID {
				r.NetworkItemID = n.ID
				return &r, nil
			}
		}
	}
	return nil, nil
}

// Create creates a new route in a network
// networkID: The ID of the network to create the route in
// route: The route configuration to create
// Returns the created route and any error that occurred
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

// Update updates an existing route
// route: The updated route configuration
// Returns any error that occurred during the update
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

// Delete removes a route by its ID
// id: The ID of the route to delete
// Returns any error that occurred during deletion
func (c *RoutesService) Delete(id string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/networks/routes/%s", c.client.GetV1Url(), id), nil)
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}
