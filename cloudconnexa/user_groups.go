package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	// ErrUserGroupNotFound is returned when a user group cannot be found
	ErrUserGroupNotFound = errors.New("user group not found")
)

// UserGroupPageResponse represents a paginated response of user groups
type UserGroupPageResponse struct {
	Content          []UserGroup `json:"content"`
	NumberOfElements int         `json:"numberOfElements"`
	Page             int         `json:"page"`
	Size             int         `json:"size"`
	Success          bool        `json:"success"`
	TotalElements    int         `json:"totalElements"`
	TotalPages       int         `json:"totalPages"`
}

// UserGroup represents a user group configuration
type UserGroup struct {
	ConnectAuth        string   `json:"connectAuth"`
	ID                 string   `json:"id"`
	InternetAccess     string   `json:"internetAccess"`
	MaxDevice          int      `json:"maxDevice"`
	Name               string   `json:"name"`
	SystemSubnets      []string `json:"systemSubnets"`
	VpnRegionIDs       []string `json:"vpnRegionIds"`
	AllRegionsIncluded bool     `json:"allRegionsIncluded"`
}

// UserGroupsService provides methods for managing user groups
type UserGroupsService service

// GetByPage retrieves user groups using pagination
// page: The page number to retrieve
// pageSize: The number of items per page
// Returns a page of user groups and any error that occurred
func (c *UserGroupsService) GetByPage(page int, pageSize int) (UserGroupPageResponse, error) {
	endpoint := fmt.Sprintf("%s/user-groups?page=%d&size=%d", c.client.GetV1Url(), page, pageSize)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return UserGroupPageResponse{}, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return UserGroupPageResponse{}, err
	}

	var response UserGroupPageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return UserGroupPageResponse{}, err
	}
	return response, nil
}

// List retrieves all user groups by paginating through all available pages
// Returns a slice of user groups and any error that occurred
func (c *UserGroupsService) List() ([]UserGroup, error) {
	var allUserGroups []UserGroup
	page := 0

	for {
		response, err := c.GetByPage(page, defaultPageSize)
		if err != nil {
			return nil, err
		}

		allUserGroups = append(allUserGroups, response.Content...)

		if page >= response.TotalPages {
			break
		}
		page++
	}
	return allUserGroups, nil
}

// GetByName retrieves a user group by its name
// name: The name of the user group to retrieve
// Returns the user group and any error that occurred
func (c *UserGroupsService) GetByName(name string) (*UserGroup, error) {
	userGroups, err := c.List()
	if err != nil {
		return nil, err
	}

	for _, ug := range userGroups {
		if ug.Name == name {
			return &ug, nil
		}
	}
	return nil, ErrUserGroupNotFound
}

// GetByID retrieves a user group by its ID using the direct API endpoint.
// This is the preferred method for getting a single user group as it uses the direct
// GET /api/v1/user-groups/{id} endpoint introduced in API v1.1.0.
// id: The ID of the user group to retrieve
// Returns the user group and any error that occurred
func (c *UserGroupsService) GetByID(id string) (*UserGroup, error) {
	endpoint := fmt.Sprintf("%s/user-groups/%s", c.client.GetV1Url(), id)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var userGroup UserGroup
	err = json.Unmarshal(body, &userGroup)
	if err != nil {
		return nil, err
	}
	return &userGroup, nil
}

// Get retrieves a user group by its ID using pagination search.
// Deprecated: Use GetByID() instead for better performance with the direct API endpoint.
// id: The ID of the user group to retrieve
// Returns the user group and any error that occurred
func (c *UserGroupsService) Get(id string) (*UserGroup, error) {
	userGroups, err := c.List()
	if err != nil {
		return nil, err
	}

	for _, ug := range userGroups {
		if ug.ID == id {
			return &ug, nil
		}
	}
	return nil, ErrUserGroupNotFound
}

// Create creates a new user group
// userGroup: The user group configuration to create
// Returns the created user group and any error that occurred
func (c *UserGroupsService) Create(userGroup *UserGroup) (*UserGroup, error) {
	userGroupJSON, err := json.Marshal(userGroup)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/user-groups", c.client.GetV1Url()), bytes.NewBuffer(userGroupJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var ug UserGroup
	err = json.Unmarshal(body, &ug)
	if err != nil {
		return nil, err
	}
	return &ug, nil
}

// Update updates an existing user group
// id: The ID of the user group to update
// userGroup: The updated user group configuration
// Returns the updated user group and any error that occurred
func (c *UserGroupsService) Update(id string, userGroup *UserGroup) (*UserGroup, error) {
	userGroupJSON, err := json.Marshal(userGroup)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/user-groups/%s", c.client.GetV1Url(), id), bytes.NewBuffer(userGroupJSON))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}
	var ug UserGroup
	err = json.Unmarshal(body, &ug)
	if err != nil {
		return nil, err
	}
	return &ug, nil
}

// Delete removes a user group by its ID
// id: The ID of the user group to delete
// Returns any error that occurred during deletion
func (c *UserGroupsService) Delete(id string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/user-groups/%s", c.client.GetV1Url(), id), nil)
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	if err != nil {
		return err
	}
	return nil
}
