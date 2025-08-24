package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

var (
	// ErrUserNotFound is returned when a user cannot be found
	ErrUserNotFound = errors.New("user not found")
)

// User represents a user configuration
type User struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Username         string   `json:"username"`
	Role             string   `json:"role"`
	Email            string   `json:"email,omitempty"`
	AuthType         string   `json:"authType"`
	FirstName        string   `json:"firstName,omitempty"`
	LastName         string   `json:"lastName,omitempty"`
	GroupID          string   `json:"groupId"`
	Status           string   `json:"status"`
	Devices          []Device `json:"devices"`
	ConnectionStatus string   `json:"connectionStatus"`
}

// UserPageResponse represents a paginated response of users
type UserPageResponse struct {
	Content          []User `json:"content"`
	NumberOfElements int    `json:"numberOfElements"`
	Page             int    `json:"page"`
	Size             int    `json:"size"`
	Success          bool   `json:"success"`
	TotalElements    int    `json:"totalElements"`
	TotalPages       int    `json:"totalPages"`
}

// Device represents a user device configuration
type Device struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IPv4Address string `json:"ipV4Address"`
	IPv6Address string `json:"ipV6Address"`
}

// UsersService provides methods for managing users
type UsersService service

// GetByPage retrieves users using pagination
// page: The page number to retrieve
// pageSize: The number of items per page
// Returns a page of users and any error that occurred
func (c *UsersService) GetByPage(page int, pageSize int) (UserPageResponse, error) {
	endpoint := fmt.Sprintf("%s/users?page=%d&size=%d", c.client.GetV1Url(), page, pageSize)
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

// List retrieves a user by username and role
// username: The username to search for
// role: The role to filter by
// Returns the user and any error that occurred
func (c *UsersService) List(username string, role string) (*User, error) {
	page := 0

	for {
		response, err := c.GetByPage(page, defaultPageSize)
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

// Get retrieves a user by ID
// userID: The ID of the user to retrieve
// Returns the user and any error that occurred
func (c *UsersService) Get(userID string) (*User, error) {
	return c.GetByID(userID)
}

// GetByID retrieves a user by ID
// userID: The ID of the user to retrieve
// Returns the user and any error that occurred
func (c *UsersService) GetByID(userID string) (*User, error) {
	endpoint := fmt.Sprintf("%s/users/%s", c.client.GetV1Url(), userID)
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

// GetByUsername retrieves a user by username
// username: The username to search for
// Returns the user and any error that occurred
func (c *UsersService) GetByUsername(username string) (*User, error) {
	page := 0

	for {
		response, err := c.GetByPage(page, defaultPageSize)
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

// Create creates a new user
// user: The user configuration to create
// Returns the created user and any error that occurred
func (c *UsersService) Create(user User) (*User, error) {
	userJSON, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s/users", c.client.GetV1Url()), bytes.NewBuffer(userJSON))
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

// Update updates an existing user
// user: The user configuration to update
// Returns any error that occurred
func (c *UsersService) Update(user User) error {
	userJSON, err := json.Marshal(user)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPut, fmt.Sprintf("%s/users/%s", c.client.GetV1Url(), user.ID), bytes.NewBuffer(userJSON))
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	if err != nil {
		return err
	}
	return nil
}

// Delete deletes a user by ID
// userID: The ID of the user to delete
// Returns any error that occurred
func (c *UsersService) Delete(userID string) error {
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/users/%s", c.client.GetV1Url(), userID), nil)
	if err != nil {
		return err
	}

	_, err = c.client.DoRequest(req)
	return err
}
