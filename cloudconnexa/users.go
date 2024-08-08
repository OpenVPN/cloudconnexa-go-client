package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrUserNotFound = errors.New("user not found")
)

type User struct {
	Id               string           `json:"id"`
	Username         string           `json:"username"`
	Role             string           `json:"role"`
	Email            string           `json:"email,omitempty"`
	AuthType         string           `json:"authType"`
	FirstName        string           `json:"firstName,omitempty"`
	LastName         string           `json:"lastName,omitempty"`
	GroupId          string           `json:"groupId"`
	Status           string           `json:"status"`
	Devices          []Device         `json:"devices"`
	ConnectionStatus ConnectionStatus `json:"connectionStatus"`
}

type UserPageResponse struct {
	Content          []User `json:"content"`
	NumberOfElements int    `json:"numberOfElements"`
	Page             int    `json:"page"`
	Size             int    `json:"size"`
	Success          bool   `json:"success"`
	TotalElements    int    `json:"totalElements"`
	TotalPages       int    `json:"totalPages"`
}

type Device struct {
	Id          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IPv4Address string `json:"ipV4Address"`
	IPv6Address string `json:"ipV6Address"`
}

type UsersService service

func (c *UsersService) GetByPage(page int, pageSize int) (UserPageResponse, error) {
	endpoint := fmt.Sprintf("%s/api/beta/users/page?page=%d&size=%d", c.client.BaseURL, page, pageSize)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return UserPageResponse{}, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return UserPageResponse{}, err
	}

	var response UserPageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return UserPageResponse{}, err
	}
	return response, nil
}

func (c *UsersService) List(username string, role string) (*User, error) {
	pageSize := 10
	page := 0

	for {
		response, err := c.GetByPage(page, pageSize)
		if err != nil {
			return nil, err
		}

		for _, user := range response.Content {
			if user.Username == username && user.Role == role {
				return &user, nil
			}
		}

		if page >= response.TotalPages {
			break
		}
		page++
	}
	return nil, ErrUserNotFound
}

func (c *UsersService) Get(userId string) (*User, error) {
	return c.GetById(userId)
}

func (c *UsersService) GetById(userId string) (*User, error) {
	endpoint := fmt.Sprintf("%s/api/beta/users/%s", c.client.BaseURL, userId)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var user User
	err = json.Unmarshal(body, &user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (c *UsersService) GetByUsername(username string) (*User, error) {
	pageSize := 10
	page := 0

	for {
		response, err := c.GetByPage(page, pageSize)
		if err != nil {
			return nil, err
		}

		for _, user := range response.Content {
			if user.Username == username {
				return &user, nil
			}
		}

		if page >= response.TotalPages {
			break
		}
		page++
	}
	return nil, ErrUserNotFound
}

func (c *UsersService) Create(user User) (*User, error) {
	userJson, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/api/beta/users", c.client.BaseURL), bytes.NewBuffer(userJson))
	if err != nil {
		return nil, err
	}

	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var u User
	err = json.Unmarshal(body, &u)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (c *UsersService) Update(user User) error {
	userJson, err := json.Marshal(user)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/api/beta/users/%s", c.client.BaseURL, user.Id), bytes.NewBuffer(userJson))
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	if err != nil {
		return err
	}
	return nil
}

func (c *UsersService) Delete(userId string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/api/beta/users/%s", c.client.BaseURL, userId), nil)
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}
