package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
)

// DeviceDetail represents a device in CloudConnexa.
// Fields match the API v1.2.0 DeviceResponse schema.
type DeviceDetail struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	Platform         string `json:"platform,omitempty"`
	UserID           string `json:"userId"`
	ClientUUID       string `json:"clientUUID,omitempty"`
	IPV4Address      string `json:"ipV4Address,omitempty"`
	IPV6Address      string `json:"ipV6Address,omitempty"`
	ConnectionStatus string `json:"connectionStatus,omitempty"`
}

// DevicePageResponse represents a paginated response of devices.
type DevicePageResponse struct {
	Content          []DeviceDetail `json:"content"`
	NumberOfElements int            `json:"numberOfElements"`
	Page             int            `json:"page"`
	Size             int            `json:"size"`
	Success          bool           `json:"success"`
	TotalElements    int            `json:"totalElements"`
	TotalPages       int            `json:"totalPages"`
}

// DeviceListOptions represents the options for listing devices.
type DeviceListOptions struct {
	UserID string `json:"userId,omitempty"`
	Page   int    `json:"page,omitempty"`
	Size   int    `json:"size,omitempty"`
}

// DeviceUpdateRequest represents the request body for updating a device.
type DeviceUpdateRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

// DeviceCreateRequest represents the request body for creating a device.
// Mirrors the API DeviceRequest schema.
type DeviceCreateRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	ClientUUID  string `json:"clientUUID,omitempty"`
}

// DevicesService provides methods for managing devices.
type DevicesService service

// List retrieves a list of devices with optional filtering and pagination.
func (d *DevicesService) List(options DeviceListOptions) (*DevicePageResponse, error) {
	// Build query parameters
	params := url.Values{}

	if options.UserID != "" {
		params.Set("userId", options.UserID)
	}

	if options.Page > 0 {
		params.Set("page", strconv.Itoa(options.Page))
	}

	if options.Size > 0 {
		// Validate size parameter (1-1000 according to API docs)
		if options.Size < 1 || options.Size > 1000 {
			return nil, fmt.Errorf("size must be between 1 and 1000, got %d", options.Size)
		}
		params.Set("size", strconv.Itoa(options.Size))
	}

	endpoint := fmt.Sprintf("%s/devices", d.client.GetV1Url())
	if len(params) > 0 {
		endpoint += "?" + params.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := d.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var response DevicePageResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetByPage retrieves devices using pagination.
func (d *DevicesService) GetByPage(page int, pageSize int) (*DevicePageResponse, error) {
	options := DeviceListOptions{
		Page: page,
		Size: pageSize,
	}
	return d.List(options)
}

// ListAll retrieves all devices by paginating through all available pages.
func (d *DevicesService) ListAll() ([]DeviceDetail, error) {
	var allDevices []DeviceDetail
	page := 0

	for {
		response, err := d.GetByPage(page, defaultPageSize)
		if err != nil {
			return nil, err
		}

		allDevices = append(allDevices, response.Content...)

		// If we've reached the last page, break
		if page >= response.TotalPages-1 {
			break
		}
		page++
	}

	return allDevices, nil
}

// GetByID retrieves a specific device by its ID.
func (d *DevicesService) GetByID(deviceID string) (*DeviceDetail, error) {
	if err := validateID(deviceID); err != nil {
		return nil, err
	}
	endpoint := buildURL(d.client.GetV1Url(), "devices", deviceID)
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := d.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var device DeviceDetail
	err = json.Unmarshal(body, &device)
	if err != nil {
		return nil, err
	}

	return &device, nil
}

// Update updates an existing device by its ID.
func (d *DevicesService) Update(deviceID string, updateRequest DeviceUpdateRequest) (*DeviceDetail, error) {
	if err := validateID(deviceID); err != nil {
		return nil, err
	}
	requestJSON, err := json.Marshal(updateRequest)
	if err != nil {
		return nil, err
	}

	endpoint := buildURL(d.client.GetV1Url(), "devices", deviceID)
	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(requestJSON))
	if err != nil {
		return nil, err
	}

	body, err := d.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var device DeviceDetail
	err = json.Unmarshal(body, &device)
	if err != nil {
		return nil, err
	}

	return &device, nil
}

// ListByUserID retrieves all devices for a specific user.
func (d *DevicesService) ListByUserID(userID string) ([]DeviceDetail, error) {
	var allDevices []DeviceDetail
	page := 0

	for {
		options := DeviceListOptions{
			UserID: userID,
			Page:   page,
			Size:   defaultPageSize,
		}

		response, err := d.List(options)
		if err != nil {
			return nil, err
		}

		allDevices = append(allDevices, response.Content...)

		// If we've reached the last page, break
		if page >= response.TotalPages-1 {
			break
		}
		page++
	}

	return allDevices, nil
}

// UpdateName updates the name of a device.
func (d *DevicesService) UpdateName(deviceID string, name string) (*DeviceDetail, error) {
	updateRequest := DeviceUpdateRequest{
		Name: name,
	}
	return d.Update(deviceID, updateRequest)
}

// UpdateDescription updates the description of a device.
func (d *DevicesService) UpdateDescription(deviceID string, description string) (*DeviceDetail, error) {
	updateRequest := DeviceUpdateRequest{
		Description: description,
	}
	return d.Update(deviceID, updateRequest)
}

// Create creates a new device for the given user.
// userID is sent as the required ?userId= query parameter.
func (d *DevicesService) Create(userID string, req DeviceCreateRequest) (*DeviceDetail, error) {
	if err := validateID(userID); err != nil {
		return nil, err
	}
	requestJSON, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	params := url.Values{}
	params.Set("userId", userID)
	endpoint := fmt.Sprintf("%s/devices?%s", d.client.GetV1Url(), params.Encode())
	httpReq, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(requestJSON))
	if err != nil {
		return nil, err
	}

	body, err := d.client.DoRequest(httpReq)
	if err != nil {
		return nil, err
	}

	var device DeviceDetail
	if err := json.Unmarshal(body, &device); err != nil {
		return nil, err
	}
	return &device, nil
}

// Delete removes a device.
// userID is sent as the required ?userId= query parameter.
func (d *DevicesService) Delete(userID, deviceID string) error {
	if err := validateID(userID); err != nil {
		return err
	}
	if err := validateID(deviceID); err != nil {
		return err
	}

	params := url.Values{}
	params.Set("userId", userID)
	endpoint := fmt.Sprintf("%s?%s", buildURL(d.client.GetV1Url(), "devices", deviceID), params.Encode())
	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	_, err = d.client.DoRequest(req)
	return err
}

// GenerateProfile generates a .ovpn profile for an existing device.
// userID and regionID are sent as the required query parameters.
// Returns the OpenVPN profile body as a string.
func (d *DevicesService) GenerateProfile(userID, deviceID, regionID string) (string, error) {
	if err := validateID(userID); err != nil {
		return "", err
	}
	if err := validateID(deviceID); err != nil {
		return "", err
	}
	if err := validateID(regionID); err != nil {
		return "", err
	}

	params := url.Values{}
	params.Set("userId", userID)
	params.Set("regionId", regionID)
	endpoint := fmt.Sprintf("%s?%s", buildURL(d.client.GetV1Url(), "devices", deviceID, "profile"), params.Encode())
	req, err := http.NewRequest(http.MethodPost, endpoint, nil)
	if err != nil {
		return "", err
	}

	body, err := d.client.DoRequest(req)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// RevokeProfile revokes the currently active profile for a device.
// userID is sent as the required ?userId= query parameter.
func (d *DevicesService) RevokeProfile(userID, deviceID string) error {
	if err := validateID(userID); err != nil {
		return err
	}
	if err := validateID(deviceID); err != nil {
		return err
	}

	params := url.Values{}
	params.Set("userId", userID)
	endpoint := fmt.Sprintf("%s?%s", buildURL(d.client.GetV1Url(), "devices", deviceID, "profile"), params.Encode())
	req, err := http.NewRequest(http.MethodDelete, endpoint, nil)
	if err != nil {
		return err
	}

	_, err = d.client.DoRequest(req)
	return err
}
