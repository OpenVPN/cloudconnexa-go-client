package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// HostRoute represents a host route configuration.
type HostRoute struct {
	ID              string `json:"id,omitempty"`
	Type            string `json:"type,omitempty"`
	Subnet          string `json:"subnet,omitempty"`
	Domain          string `json:"domain,omitempty"`
	Description     string `json:"description,omitempty"`
	ParentRouteID   string `json:"parentRouteId,omitempty"`
	NetworkItemID   string `json:"networkItemId,omitempty"`
	AllowEmbeddedIP bool   `json:"allowEmbeddedIp,omitempty"`
}

// HostRoutePageResponse represents a paginated response of host routes.
type HostRoutePageResponse struct {
	Success          bool        `json:"success"`
	Content          []HostRoute `json:"content"`
	TotalElements    int         `json:"totalElements"`
	TotalPages       int         `json:"totalPages"`
	NumberOfElements int         `json:"numberOfElements"`
	Page             int         `json:"page"`
	Size             int         `json:"size"`
}

// HostRoutesService provides methods for managing host routes.
type HostRoutesService service

// GetByPage retrieves host routes using pagination.
// hostID: The ID of the host to get routes for
// page: The page number to retrieve
// size: The number of items per page
// Returns a page of routes and any error that occurred
func (c *HostRoutesService) GetByPage(hostID string, page int, size int) (HostRoutePageResponse, error) {
	if err := validateID(hostID); err != nil {
		return HostRoutePageResponse{}, err
	}
	params := url.Values{}
	params.Set("hostId", hostID)
	params.Set("page", strconv.Itoa(page))
	params.Set("size", strconv.Itoa(size))
	endpoint := fmt.Sprintf("%s/hosts/routes?%s", c.client.GetV1Url(), params.Encode())
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return HostRoutePageResponse{}, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return HostRoutePageResponse{}, err
	}

	var response HostRoutePageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return HostRoutePageResponse{}, err
	}
	return response, nil
}

// List retrieves all routes for a host by paginating through all available pages.
// hostID: The ID of the host to get routes for
// Returns a slice of routes and any error that occurred
func (c *HostRoutesService) List(hostID string) ([]HostRoute, error) {
	var allRoutes []HostRoute
	page := 0

	for {
		response, err := c.GetByPage(hostID, page, defaultPageSize)
		if err != nil {
			return nil, err
		}

		allRoutes = append(allRoutes, response.Content...)

		page++
		if page >= response.TotalPages {
			break
		}
	}
	return allRoutes, nil
}

// GetByID retrieves a specific host route by its ID.
// routeID: The ID of the route to retrieve
// Returns the route and any error that occurred
func (c *HostRoutesService) GetByID(routeID string) (*HostRoute, error) {
	if err := validateID(routeID); err != nil {
		return nil, err
	}
	endpoint := buildURL(c.client.GetV1Url(), "hosts", "routes", routeID)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var route HostRoute
	err = json.Unmarshal(body, &route)
	if err != nil {
		return nil, err
	}
	return &route, nil
}

// Create creates a new route for a host.
// hostID: The ID of the host to create the route for
// route: The route configuration to create
// Returns the created route and any error that occurred
func (c *HostRoutesService) Create(hostID string, route HostRoute) (*HostRoute, error) {
	if err := validateID(hostID); err != nil {
		return nil, err
	}
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

	params := url.Values{}
	params.Set("hostId", hostID)
	endpoint := fmt.Sprintf("%s/hosts/routes?%s", c.client.GetV1Url(), params.Encode())
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(routeJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var r HostRoute
	err = json.Unmarshal(body, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

// Update updates an existing host route.
// route: The updated route configuration (must include ID)
// Returns any error that occurred during the update
func (c *HostRoutesService) Update(route HostRoute) error {
	if err := validateID(route.ID); err != nil {
		return err
	}
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

	endpoint := buildURL(c.client.GetV1Url(), "hosts", "routes", route.ID)
	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(routeJSON))
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}

// Delete removes a host route by its ID.
// routeID: The ID of the route to delete
// Returns any error that occurred during deletion
func (c *HostRoutesService) Delete(routeID string) error {
	if err := validateID(routeID); err != nil {
		return err
	}
	endpoint := buildURL(c.client.GetV1Url(), "hosts", "routes", routeID)
	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}
