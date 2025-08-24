package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// DeviceStatus represents the possible statuses of a device.
type DeviceStatus string

const (
	// DeviceStatusActive represents an active device.
	DeviceStatusActive DeviceStatus = "ACTIVE"
	// DeviceStatusInactive represents an inactive device.
	DeviceStatusInactive DeviceStatus = "INACTIVE"
	// DeviceStatusBlocked represents a blocked device.
	DeviceStatusBlocked DeviceStatus = "BLOCKED"
	// DeviceStatusPending represents a pending device.
	DeviceStatusPending DeviceStatus = "PENDING"
)

// DeviceType represents the type of device.
type DeviceType string

const (
	// DeviceTypeClient represents a client device.
	DeviceTypeClient DeviceType = "CLIENT"
	// DeviceTypeConnector represents a connector device.
	DeviceTypeConnector DeviceType = "CONNECTOR"
)

// DeviceDetail represents a detailed device in CloudConnexa.
type DeviceDetail struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	Description        string                 `json:"description,omitempty"`
	UserID             string                 `json:"userId"`
	Status             string                 `json:"status"`
	Type               string                 `json:"type"`
	Platform           string                 `json:"platform,omitempty"`
	Version            string                 `json:"version,omitempty"`
	LastSeen           *time.Time             `json:"lastSeen,omitempty"`
	CreatedAt          time.Time              `json:"createdAt"`
	UpdatedAt          time.Time              `json:"updatedAt"`
	IPv4Address        string                 `json:"ipv4Address,omitempty"`
	IPv6Address        string                 `json:"ipv6Address,omitempty"`
	PublicKey          string                 `json:"publicKey,omitempty"`
	Fingerprint        string                 `json:"fingerprint,omitempty"`
	SerialNumber       string                 `json:"serialNumber,omitempty"`
	MACAddress         string                 `json:"macAddress,omitempty"`
	Hostname           string                 `json:"hostname,omitempty"`
	OperatingSystem    string                 `json:"operatingSystem,omitempty"`
	OSVersion          string                 `json:"osVersion,omitempty"`
	ClientVersion      string                 `json:"clientVersion,omitempty"`
	IsOnline           bool                   `json:"isOnline"`
	LastConnectedAt    *time.Time             `json:"lastConnectedAt,omitempty"`
	LastDisconnectedAt *time.Time             `json:"lastDisconnectedAt,omitempty"`
	TotalBytesIn       int64                  `json:"totalBytesIn"`
	TotalBytesOut      int64                  `json:"totalBytesOut"`
	SessionCount       int                    `json:"sessionCount"`
	Region             string                 `json:"region,omitempty"`
	Gateway            string                 `json:"gateway,omitempty"`
	UserGroupID        string                 `json:"userGroupId,omitempty"`
	NetworkID          string                 `json:"networkId,omitempty"`
	Tags               []string               `json:"tags,omitempty"`
	Metadata           map[string]interface{} `json:"metadata,omitempty"`
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
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Status      string                 `json:"status,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
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
	endpoint := fmt.Sprintf("%s/devices/%s", d.client.GetV1Url(), deviceID)
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
	requestJSON, err := json.Marshal(updateRequest)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/devices/%s", d.client.GetV1Url(), deviceID)
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

// Block blocks a device by updating its status to BLOCKED.
func (d *DevicesService) Block(deviceID string) (*DeviceDetail, error) {
	updateRequest := DeviceUpdateRequest{
		Status: string(DeviceStatusBlocked),
	}
	return d.Update(deviceID, updateRequest)
}

// Unblock unblocks a device by updating its status to ACTIVE.
func (d *DevicesService) Unblock(deviceID string) (*DeviceDetail, error) {
	updateRequest := DeviceUpdateRequest{
		Status: string(DeviceStatusActive),
	}
	return d.Update(deviceID, updateRequest)
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

// UpdateTags updates the tags of a device.
func (d *DevicesService) UpdateTags(deviceID string, tags []string) (*DeviceDetail, error) {
	updateRequest := DeviceUpdateRequest{
		Tags: tags,
	}
	return d.Update(deviceID, updateRequest)
}
