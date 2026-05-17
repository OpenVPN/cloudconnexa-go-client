// Package cloudconnexa provides a Go client library for the CloudConnexa API.
// It offers comprehensive functionality for managing VPN networks, hosts, connectors,
// routes, users, and other CloudConnexa resources through a simple Go interface.
package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// AccessGroup represents a group of access rules that define network access permissions.
// It contains source and destination rules that determine what resources can access each other.
type AccessGroup struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description,omitempty"`
	Source      []AccessItem `json:"source"`
	Destination []AccessItem `json:"destination"`
}

// AccessItem represents a single access rule item that can be either a source or destination.
// It defines what resources are covered by the access rule and their relationships.
type AccessItem struct {
	Type       string   `json:"type"`
	AllCovered bool     `json:"allCovered"`
	Parent     string   `json:"parent,omitempty"`
	Children   []string `json:"children,omitempty"`
}

// Item represents a basic resource with an identifier.
type Item struct {
	ID string `json:"id"`
}

// AccessGroupPageResponse represents a paginated response from the CloudConnexa API
// containing a list of access groups and pagination metadata.
type AccessGroupPageResponse struct {
	Content          []AccessGroup `json:"content"`
	NumberOfElements int           `json:"numberOfElements"`
	Page             int           `json:"page"`
	Size             int           `json:"size"`
	Success          bool          `json:"success"`
	TotalElements    int           `json:"totalElements"`
	TotalPages       int           `json:"totalPages"`
}

// AccessGroupsService handles communication with the CloudConnexa API for access group operations.
type AccessGroupsService service

// GetAccessGroupsByPage retrieves a paginated list of access groups from the CloudConnexa API.
// It returns the access groups for the specified page and page size.
func (c *AccessGroupsService) GetAccessGroupsByPage(page int, size int) (AccessGroupPageResponse, error) {
	endpoint := fmt.Sprintf("%s/access-groups?page=%d&size=%d", c.client.GetV1Url(), page, size)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return AccessGroupPageResponse{}, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return AccessGroupPageResponse{}, err
	}

	var response AccessGroupPageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return AccessGroupPageResponse{}, err
	}
	return response, nil
}

// List retrieves all access groups from the CloudConnexa API.
// It handles pagination internally and returns a complete list of access groups.
func (c *AccessGroupsService) List() ([]AccessGroup, error) {
	var allGroups []AccessGroup
	page := 0

	for {
		response, err := c.GetAccessGroupsByPage(page, defaultPageSize)
		if err != nil {
			return nil, err
		}

		allGroups = append(allGroups, response.Content...)

		page++
		if page >= response.TotalPages {
			break
		}
	}
	return allGroups, nil
}

// Get retrieves a specific access group by its ID from the CloudConnexa API.
func (c *AccessGroupsService) Get(id string) (*AccessGroup, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	endpoint := buildURL(c.client.GetV1Url(), "access-groups", id)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var accessGroup AccessGroup
	err = json.Unmarshal(body, &accessGroup)
	if err != nil {
		return nil, err
	}
	return &accessGroup, nil
}

// GetByName retrieves an access group by its name
// name: The name of the access group to retrieve
// Returns the access group and any error that occurred
func (c *AccessGroupsService) GetByName(name string) (*AccessGroup, error) {
	items, err := c.List()
	if err != nil {
		return nil, err
	}

	for _, item := range items {
		if item.Name == name {
			return &item, nil
		}
	}
	return nil, errors.New("access group not found")
}

// Create creates a new access group in the CloudConnexa API.
// It returns the created access group with its assigned ID.
func (c *AccessGroupsService) Create(accessGroup *AccessGroup) (*AccessGroup, error) {
	accessGroupJSON, err := json.Marshal(accessGroup)
	if err != nil {
		return nil, err
	}

	endpoint := buildURL(c.client.GetV1Url(), "access-groups")

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(accessGroupJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var s AccessGroup
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Update updates an existing access group in the CloudConnexa API.
// It returns the updated access group.
func (c *AccessGroupsService) Update(id string, accessGroup *AccessGroup) (*AccessGroup, error) {
	if err := validateID(id); err != nil {
		return nil, err
	}
	accessGroupJSON, err := json.Marshal(accessGroup)
	if err != nil {
		return nil, err
	}

	endpoint := buildURL(c.client.GetV1Url(), "access-groups", id)

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(accessGroupJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var s AccessGroup
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// Delete removes an access group from the CloudConnexa API by its ID.
func (c *AccessGroupsService) Delete(id string) error {
	if err := validateID(id); err != nil {
		return err
	}
	endpoint := buildURL(c.client.GetV1Url(), "access-groups", id)
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
