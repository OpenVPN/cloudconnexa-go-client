package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type UserGroupPageResponse struct {
	Content          []UserGroup `json:"content"`
	NumberOfElements int         `json:"numberOfElements"`
	Page             int         `json:"page"`
	Size             int         `json:"size"`
	Success          bool        `json:"success"`
	TotalElements    int         `json:"totalElements"`
	TotalPages       int         `json:"totalPages"`
}

type UserGroup struct {
	ConnectAuth    string   `json:"connectAuth"`
	ID             string   `json:"id"`
	InternetAccess string   `json:"internetAccess"`
	MaxDevice      int      `json:"maxDevice"`
	Name           string   `json:"name"`
	SystemSubnets  []string `json:"systemSubnets"`
	VpnRegionIds   []string `json:"vpnRegionIds"`
}

type UserGroupsService service

func (c *UserGroupsService) GetByPage(page int, pageSize int) (UserGroupPageResponse, error) {
	endpoint := fmt.Sprintf("%s/api/beta/user-groups/page?page=%d&size=%d", c.client.BaseURL, page, pageSize)
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

func (c *UserGroupsService) List() ([]UserGroup, error) {
	var allUserGroups []UserGroup
	pageSize := 10
	page := 0

	for {
		response, err := c.GetByPage(page, pageSize)
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
	return nil, fmt.Errorf("group %s does not exist", name)
}

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
	return nil, fmt.Errorf("group %s does not exist", id)
}

func (c *UserGroupsService) Create(userGroup *UserGroup) (*UserGroup, error) {
	userGroupJson, err := json.Marshal(userGroup)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/beta/user-groups", c.client.BaseURL), bytes.NewBuffer(userGroupJson))
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

func (c *UserGroupsService) Update(id string, userGroup *UserGroup) (*UserGroup, error) {
	userGroupJson, err := json.Marshal(userGroup)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/beta/user-groups/%s", c.client.BaseURL, id), bytes.NewBuffer(userGroupJson))
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

func (c *UserGroupsService) Delete(id string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/beta/user-groups/%s", c.client.BaseURL, id), nil)
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	if err != nil {
		return err
	}
	return nil
}
