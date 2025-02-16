package cloudconnexa

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

const (
	userAgent = "cloudconnexa-go"
)

type Client struct {
	client *http.Client

	BaseURL     string
	Token       string
	RateLimiter *rate.Limiter

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
	Users               *UsersService
	UserGroups          *UserGroupsService
	VPNRegions          *VPNRegionsService
	LocationContexts    *LocationContextsService
	AccessGroups        *AccessGroupsService
}

type service struct {
	client *Client
}

type Credentials struct {
	AccessToken string `json:"access_token"`
}

type ErrClientResponse struct {
	status int
	body   string
}

func (e ErrClientResponse) Error() string {
	return fmt.Sprintf("status code: %d, response body: %s", e.status, e.body)
}

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

	defer resp.Body.Close()
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
		client:      httpClient,
		BaseURL:     baseURL,
		Token:       credentials.AccessToken,
		UserAgent:   userAgent,
		RateLimiter: rate.NewLimiter(rate.Every(1*time.Second), 5),
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
	c.Users = (*UsersService)(&c.common)
	c.UserGroups = (*UserGroupsService)(&c.common)
	c.VPNRegions = (*VPNRegionsService)(&c.common)
	c.LocationContexts = (*LocationContextsService)(&c.common)
	c.AccessGroups = (*AccessGroupsService)(&c.common)
	return c, nil
}

func (c *Client) DoRequest(req *http.Request) ([]byte, error) {
	err := c.RateLimiter.Wait(context.Background())
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("User-Agent", c.UserAgent)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, &ErrClientResponse{status: res.StatusCode, body: string(body)}
	}

	return body, nil
}

func (c *Client) GetV1Url() string {
	return c.BaseURL + "/api/v1"
}
