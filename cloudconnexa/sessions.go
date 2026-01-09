package cloudconnexa

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

// SessionStatus represents the possible statuses of an OpenVPN session.
type SessionStatus string

const (
	// SessionStatusActive represents an active session.
	SessionStatusActive SessionStatus = "ACTIVE"
	// SessionStatusCompleted represents a completed session.
	SessionStatusCompleted SessionStatus = "COMPLETED"
	// SessionStatusFailed represents a failed session.
	SessionStatusFailed SessionStatus = "FAILED"
)

// Session represents an OpenVPN session in CloudConnexa.
// Fields match the API v1.2.0 SessionResponse schema.
type Session struct {
	SessionID        string    `json:"sessionId"`
	UserID           string    `json:"userId"`
	DeviceID         string    `json:"deviceId"`
	RegionID         string    `json:"regionId"`
	BytesIn          int64     `json:"bytesIn"`
	BytesOut         int64     `json:"bytesOut"`
	ConnectorName    string    `json:"connectorName,omitempty"`
	UserName         string    `json:"userName,omitempty"`
	DeviceName       string    `json:"deviceName,omitempty"`
	ClientIP         string    `json:"clientIp,omitempty"`
	StartDateTime    time.Time `json:"startDateTime"`
	VpnIPv4          string    `json:"vpnIpv4,omitempty"`
	NetworkName      string    `json:"networkName,omitempty"`
	RegionName       string    `json:"regionName,omitempty"`
	ConnectionStatus string    `json:"connectionStatus,omitempty"`
}

// SessionsResponse represents the response from the sessions API endpoint.
type SessionsResponse struct {
	Sessions   []Session `json:"sessions"`
	NextCursor string    `json:"nextCursor,omitempty"`
}

// SessionsListOptions represents the options for listing sessions.
type SessionsListOptions struct {
	StartDate     *time.Time    `json:"startDate,omitempty"`
	EndDate       *time.Time    `json:"endDate,omitempty"`
	Status        SessionStatus `json:"status,omitempty"`
	ReturnOnlyNew bool          `json:"returnOnlyNew,omitempty"`
	Size          int           `json:"size"`
	Cursor        string        `json:"cursor,omitempty"`
}

// SessionsService provides methods for managing OpenVPN sessions.
type SessionsService service

// List retrieves a list of OpenVPN sessions with optional filtering.
// The size parameter is required and must be between 1 and 100.
// Returns a SessionsResponse containing sessions and optional next cursor for pagination.
func (s *SessionsService) List(options SessionsListOptions) (*SessionsResponse, error) {
	// Validate size parameter
	if options.Size < 1 || options.Size > 100 {
		return nil, fmt.Errorf("size must be between 1 and 100, got %d", options.Size)
	}

	// Build query parameters
	params := url.Values{}
	params.Set("size", strconv.Itoa(options.Size))

	if options.StartDate != nil {
		params.Set("startDate", options.StartDate.Format(time.RFC3339))
	}

	if options.EndDate != nil {
		params.Set("endDate", options.EndDate.Format(time.RFC3339))
	}

	if options.Status != "" {
		params.Set("status", string(options.Status))
	}

	if options.ReturnOnlyNew {
		params.Set("returnOnlyNew", "true")
	}

	if options.Cursor != "" {
		params.Set("cursor", options.Cursor)
	}

	endpoint := fmt.Sprintf("%s/sessions?%s", s.client.GetV1Url(), params.Encode())
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	body, err := s.client.DoRequest(req)
	if err != nil {
		return nil, err
	}

	var response SessionsResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// ListAll retrieves all sessions by automatically handling pagination.
// This method will make multiple API calls if necessary to retrieve all sessions.
// Use with caution as it may result in many API calls for large datasets.
func (s *SessionsService) ListAll(options SessionsListOptions) ([]Session, error) {
	var allSessions []Session
	cursor := options.Cursor

	// Set a reasonable default size if not specified
	if options.Size == 0 {
		options.Size = 100
	}

	for {
		options.Cursor = cursor
		response, err := s.List(options)
		if err != nil {
			return nil, err
		}

		allSessions = append(allSessions, response.Sessions...)

		// If there's no next cursor, we've retrieved all sessions
		if response.NextCursor == "" {
			break
		}

		cursor = response.NextCursor
	}

	return allSessions, nil
}

// ListActive retrieves all active sessions.
func (s *SessionsService) ListActive(size int) (*SessionsResponse, error) {
	options := SessionsListOptions{
		Status: SessionStatusActive,
		Size:   size,
	}
	return s.List(options)
}

// ListByDateRange retrieves sessions within a specific date range.
func (s *SessionsService) ListByDateRange(startDate, endDate time.Time, size int) (*SessionsResponse, error) {
	options := SessionsListOptions{
		StartDate: &startDate,
		EndDate:   &endDate,
		Size:      size,
	}
	return s.List(options)
}

// ListByStatus retrieves sessions with a specific status.
func (s *SessionsService) ListByStatus(status SessionStatus, size int) (*SessionsResponse, error) {
	options := SessionsListOptions{
		Status: status,
		Size:   size,
	}
	return s.List(options)
}
