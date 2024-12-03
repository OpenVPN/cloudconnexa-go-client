package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type AccessGroupRequest struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Source      []AccessItemRequest `json:"source"`
	Destination []AccessItemRequest `json:"destination"`
}

type AccessGroupResponse struct {
	Id          string               `json:"id"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Source      []AccessItemResponse `json:"source"`
	Destination []AccessItemResponse `json:"destination"`
}

type AccessItemRequest struct {
	Type       string   `json:"type"`
	AllCovered bool     `json:"allCovered"`
	Parent     string   `json:"parent,omitempty"`
	Children   []string `json:"children,omitempty"`
}

type AccessItemResponse struct {
	Type       string `json:"type"`
	AllCovered bool   `json:"allCovered"`
	Parent     *Item  `json:"parent"`
	Children   []Item `json:"children"`
}

type Item struct {
	Id string `json:"id"`
}

type AccessGroupPageResponse struct {
	Content          []AccessGroupResponse `json:"content"`
	NumberOfElements int                   `json:"numberOfElements"`
	Page             int                   `json:"page"`
	Size             int                   `json:"size"`
	Success          bool                  `json:"success"`
	TotalElements    int                   `json:"totalElements"`
	TotalPages       int                   `json:"totalPages"`
}

type AccessGroupsService service

func (c *AccessGroupsService) GetAccessGroupsByPage(page int, size int) (AccessGroupPageResponse, error) {
	endpoint := fmt.Sprintf("%s/api/beta/access-groups/page?page=%d&size=%d", c.client.BaseURL, page, size)
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

func (c *AccessGroupsService) List() ([]AccessGroupResponse, error) {
	var allGroups []AccessGroupResponse
	page := 0
	pageSize := 10

	for {
		response, err := c.GetAccessGroupsByPage(page, pageSize)
		if err != nil {
			return nil, err
		}

		allGroups = append(allGroups, response.Content...)
		if page >= response.TotalPages {
			break
		}
		page++
	}
	return allGroups, nil
}

func (c *AccessGroupsService) Get(id string) (*AccessGroupResponse, error) {
	groups, err := c.List()
	if err != nil {
		return nil, err
	}

	for _, n := range groups {
		if n.Id == id {
			return &n, nil
		}
	}
	return nil, nil
}

func (c *AccessGroupsService) Create(accessGroup *AccessGroupRequest) (*AccessGroupResponse, error) {
	accessGroupJson, err := json.Marshal(accessGroup)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/api/beta/access-groups", c.client.BaseURL)

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(accessGroupJson))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var s AccessGroupResponse
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *AccessGroupsService) Update(id string, accessGroup *AccessGroupRequest) (*AccessGroupResponse, error) {
	accessGroupJson, err := json.Marshal(accessGroup)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/api/beta/access-groups/%s", c.client.BaseURL, id)

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(accessGroupJson))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var s AccessGroupResponse
	err = json.Unmarshal(body, &s)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (c *AccessGroupsService) Delete(id string) error {
	endpoint := fmt.Sprintf("%s/api/beta/access-groups/%s", c.client.BaseURL, id)
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
