package cloudconnexa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

const (
	userAgent       = "cloudconnexa-go"
	defaultPageSize = 100
)

// Client represents a CloudConnexa API client with all service endpoints.
type Client struct {
	client *http.Client

	BaseURL           string
	Token             string
	ReadRateLimiter   *rate.Limiter
	UpdateRateLimiter *rate.Limiter

	UserAgent string

	common service

	HostConnectors      *HostConnectorsService
	NetworkConnectors   *NetworkConnectorsService
	DNSRecords          *DNSRecordsService
	Hosts               *HostsService
	HostIPServices      *HostIPServicesService
	NetworkIPServices   *NetworkIPServicesService
	HostApplications    *HostApplicationsService
	NetworkApplications *NetworkApplicationsService
	Networks            *NetworksService
	Routes              *RoutesService
	HostRoutes          *HostRoutesService
	Users               *UsersService
	UserGroups          *UserGroupsService
	VPNRegions          *VPNRegionsService
	LocationContexts    *LocationContextsService
	AccessGroups        *AccessGroupsService
	Settings            *SettingsService
	Sessions            *SessionsService
	Devices             *DevicesService
}

type service struct {
	client *Client
}

// Credentials represents the OAuth2 token response from CloudConnexa API.
type Credentials struct {
	AccessToken string `json:"access_token"`
}

// ErrClientResponse represents an error response from the CloudConnexa API.
type ErrClientResponse struct {
	status int
	body   string
}

func (e ErrClientResponse) Error() string {
	return fmt.Sprintf("status code: %d, response body: %s", e.status, e.body)
}

// StatusCode returns the HTTP status code of the API error response.
func (e ErrClientResponse) StatusCode() int { return e.status }

// Body returns the raw response body of the API error response.
func (e ErrClientResponse) Body() string { return e.body }

// NewClient creates a new CloudConnexa API client with the given credentials.
// It authenticates using OAuth2 client credentials flow and returns a configured client.
func NewClient(baseURL, clientID, clientSecret string) (*Client, error) {
	if clientID == "" || clientSecret == "" {
		return nil, ErrCredentialsRequired
	}

	values := map[string]string{"grant_type": "client_credentials", "scope": "default"}
	jsonData, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	tokenURL := fmt.Sprintf("%s/api/v1/oauth/token", strings.TrimRight(baseURL, "/"))
	req, err := http.NewRequest(http.MethodPost, tokenURL, bytes.NewBuffer(jsonData))

	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(clientID, clientSecret)
	req.Header.Add("Accept", "application/json")
	httpClient := &http.Client{Timeout: 30 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			// Log the error if you have a logger, otherwise this is acceptable for library code
			_ = closeErr
		}
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var credentials Credentials
	err = json.Unmarshal(body, &credentials)
	if err != nil {
		return nil, err
	}

	c := &Client{
		client:            httpClient,
		BaseURL:           baseURL,
		Token:             credentials.AccessToken,
		UserAgent:         userAgent,
		ReadRateLimiter:   rate.NewLimiter(rate.Every(1*time.Second), 1),
		UpdateRateLimiter: rate.NewLimiter(rate.Every(4*time.Second), 1),
	}
	c.common.client = c
	c.HostConnectors = (*HostConnectorsService)(&c.common)
	c.NetworkConnectors = (*NetworkConnectorsService)(&c.common)
	c.DNSRecords = (*DNSRecordsService)(&c.common)
	c.Hosts = (*HostsService)(&c.common)
	c.HostIPServices = (*HostIPServicesService)(&c.common)
	c.NetworkIPServices = (*NetworkIPServicesService)(&c.common)
	c.HostApplications = (*HostApplicationsService)(&c.common)
	c.NetworkApplications = (*NetworkApplicationsService)(&c.common)
	c.Networks = (*NetworksService)(&c.common)
	c.Routes = (*RoutesService)(&c.common)
	c.HostRoutes = (*HostRoutesService)(&c.common)
	c.Users = (*UsersService)(&c.common)
	c.UserGroups = (*UserGroupsService)(&c.common)
	c.VPNRegions = (*VPNRegionsService)(&c.common)
	c.LocationContexts = (*LocationContextsService)(&c.common)
	c.AccessGroups = (*AccessGroupsService)(&c.common)
	c.Settings = (*SettingsService)(&c.common)
	c.Sessions = (*SessionsService)(&c.common)
	c.Devices = (*DevicesService)(&c.common)
	return c, nil
}

// setCommonHeaders sets the standard headers for API requests.
// It sets Authorization and User-Agent headers, and sets Content-Type to application/json
// if no Content-Type header is already present.
func (c *Client) setCommonHeaders(req *http.Request) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("User-Agent", c.UserAgent)

	if req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}
}

// DoRequest executes an HTTP request with authentication and rate limiting.
// It automatically adds the Bearer token, sets headers, and handles errors.
func (c *Client) DoRequest(req *http.Request) ([]byte, error) {
	var rateLimiter *rate.Limiter
	if req.Method == "GET" {
		rateLimiter = c.ReadRateLimiter
	} else {
		rateLimiter = c.UpdateRateLimiter
	}
	err := rateLimiter.Wait(context.Background())
	if err != nil {
		return nil, err
	}

	c.setCommonHeaders(req)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			// Log the error if you have a logger, otherwise this is acceptable for library code
			_ = closeErr
		}
	}()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, &ErrClientResponse{status: res.StatusCode, body: string(body)}
	}

	err = c.AssignLimits(res, rateLimiter)
	if err != nil {
		return nil, err
	}

	return body, nil
}

// AssignLimits adjusts the rate limiter according to values received in response headers from the API
func (c *Client) AssignLimits(res *http.Response, rateLimiter *rate.Limiter) error {
	rateHeader := res.Header.Get("X-RateLimit-Replenish-Rate")
	timeHeader := res.Header.Get("X-RateLimit-Replenish-Time")
	remainingHeader := res.Header.Get("X-RateLimit-Remaining")

	if rateHeader != "" && timeHeader != "" && remainingHeader != "" {
		rateValue, err := strconv.Atoi(rateHeader)
		if err != nil {
			return err
		}
		timeValue, err := strconv.Atoi(timeHeader)
		if err != nil {
			return err
		}
		remainingValue, err := strconv.Atoi(remainingHeader)
		if err != nil {
			return err
		}
		if remainingValue <= 0 {
			remainingValue = 1
		}
		rateLimiter.SetLimit(rate.Every(time.Duration(timeValue * 1_000_000_000 / rateValue)))
		rateLimiter.SetBurst(remainingValue)
	}
	return nil
}

// GetV1Url returns the base URL for CloudConnexa API v1 endpoints.
func (c *Client) GetV1Url() string {
	return c.BaseURL + "/api/v1"
}
