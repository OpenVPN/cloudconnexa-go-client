package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

const (
	userAgent       = "cloudconnexa-go"
	defaultPageSize = 100

	// DefaultMaxResponseSize is the maximum API response body size (10 MB).
	// This prevents memory exhaustion from malicious or compromised servers (CWE-400).
	DefaultMaxResponseSize int64 = 10 << 20

	// DefaultMaxTokenResponseSize is the maximum OAuth token response size (1 MB).
	DefaultMaxTokenResponseSize int64 = 1 << 20
)

// ClientOptions provides optional configuration for the API client.
type ClientOptions struct {
	// AllowInsecureHTTP permits HTTP connections for loopback addresses only
	// (localhost, 127.0.0.1, ::1). This is intended for local development and testing.
	// WARNING: HTTP connections to non-loopback addresses are always rejected.
	AllowInsecureHTTP bool

	// MaxRetries is how many times a request rejected with HTTP 429 Too Many Requests
	// is retried before the error is returned. Zero means DefaultMaxRetries; a negative
	// value disables retries.
	MaxRetries int

	// OnRetry, if set, is called before each retry of a request rejected with HTTP 429.
	// attempt is the number of the retry about to be made, starting at 1, and wait is how
	// long the client will hold the request before sending it again. It is called from the
	// goroutine that issued the request and must not block.
	OnRetry func(req *http.Request, attempt int, wait time.Duration)
}

// validateBaseURL validates the base URL for the API client.
// It ensures the URL is well-formed and uses HTTPS unless allowHTTP is true
// and the host is a loopback address (per RFC 9700 OAuth security best practices).
func validateBaseURL(rawURL string, allowHTTP bool) (string, error) {
	if rawURL == "" {
		return "", fmt.Errorf("%w: URL cannot be empty", ErrInvalidBaseURL)
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrInvalidBaseURL, err)
	}

	// Validate scheme is present
	if parsed.Scheme == "" {
		return "", fmt.Errorf("%w: URL must include scheme (e.g., https://)", ErrInvalidBaseURL)
	}

	// Normalize scheme to lowercase
	scheme := strings.ToLower(parsed.Scheme)

	// Validate host is present
	if parsed.Host == "" {
		return "", fmt.Errorf("%w: URL must include a host", ErrInvalidBaseURL)
	}

	// Reject URLs with embedded credentials (security: prevents logging leaks)
	if parsed.User != nil {
		return "", fmt.Errorf("%w: URL must not contain credentials", ErrInvalidBaseURL)
	}

	// Check scheme - only allow http or https
	switch scheme {
	case "https":
		// HTTPS is always allowed
	case "http":
		if !allowHTTP {
			return "", ErrHTTPSRequired
		}
		// Even with allowHTTP, only permit loopback addresses
		if !isLoopbackHost(parsed.Host) {
			return "", fmt.Errorf("%w: HTTP is only allowed for localhost/127.0.0.1/::1", ErrHTTPSRequired)
		}
	default:
		return "", fmt.Errorf("%w: unsupported scheme %q; only https is allowed", ErrInvalidBaseURL, scheme)
	}

	// Normalize: strip path, query, fragment - only keep scheme://host
	normalized := fmt.Sprintf("%s://%s", scheme, parsed.Host)
	return normalized, nil
}

// isLoopbackHost checks if the host is a loopback address.
// This includes localhost, 127.0.0.0/8 range, and ::1.
func isLoopbackHost(host string) bool {
	// Extract hostname without port using net.SplitHostPort for robustness
	hostname := host

	// Try to split host and port
	if h, _, err := net.SplitHostPort(host); err == nil {
		hostname = h
	}

	// Strip brackets from IPv6 (may remain if no port was present, e.g., "[::1]")
	hostname = strings.TrimPrefix(hostname, "[")
	hostname = strings.TrimSuffix(hostname, "]")
	hostname = strings.ToLower(hostname)

	// Check common loopback names
	if hostname == "localhost" || hostname == "::1" {
		return true
	}

	// Check 127.0.0.0/8 range
	ip := net.ParseIP(hostname)
	if ip != nil {
		return ip.IsLoopback()
	}

	return false
}

// Client represents a CloudConnexa API client with all service endpoints.
type Client struct {
	client     *http.Client
	maxRetries int
	onRetry    func(req *http.Request, attempt int, wait time.Duration)
	// The API limits GET requests and all other requests in separate buckets,
	// so retries are scheduled separately too.
	pauseRead  retryPause
	pauseWrite retryPause

	BaseURL string
	Token   string

	// ReadRateLimiter and UpdateRateLimiter are optional client-side throttles for GET
	// requests and for all other requests respectively. They default to rate.Inf, which
	// never waits; set them to pace requests proactively on top of the automatic 429 retry.
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
// The baseURL must use HTTPS. For development with localhost HTTP, use NewClientWithOptions.
// It authenticates using OAuth2 client credentials flow and returns a configured client.
func NewClient(baseURL, clientID, clientSecret string) (*Client, error) {
	return NewClientWithOptions(baseURL, clientID, clientSecret, nil)
}

// NewClientWithOptions creates a new CloudConnexa API client with custom options.
// It authenticates using OAuth2 client credentials flow and returns a configured client.
func NewClientWithOptions(baseURL, clientID, clientSecret string, opts *ClientOptions) (*Client, error) {
	if clientID == "" || clientSecret == "" {
		return nil, ErrCredentialsRequired
	}

	allowHTTP := false
	maxRetries := 0
	var onRetry func(*http.Request, int, time.Duration)
	if opts != nil {
		allowHTTP = opts.AllowInsecureHTTP
		maxRetries = opts.MaxRetries
		onRetry = opts.OnRetry
	}

	normalizedURL, err := validateBaseURL(baseURL, allowHTTP)
	if err != nil {
		return nil, err
	}

	values := map[string]string{"grant_type": "client_credentials", "scope": "default"}
	jsonData, err := json.Marshal(values)
	if err != nil {
		return nil, err
	}

	tokenURL := fmt.Sprintf("%s/api/v1/oauth/token", normalizedURL)
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

	// Bound OAuth response size to prevent memory exhaustion (CWE-400)
	limitedReader := io.LimitReader(resp.Body, DefaultMaxTokenResponseSize+1)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, err
	}

	if int64(len(body)) > DefaultMaxTokenResponseSize {
		return nil, fmt.Errorf("%w: OAuth response exceeded %d bytes", ErrResponseTooLarge, DefaultMaxTokenResponseSize)
	}

	var credentials Credentials
	err = json.Unmarshal(body, &credentials)
	if err != nil {
		return nil, err
	}

	c := &Client{
		client:            httpClient,
		BaseURL:           normalizedURL,
		Token:             credentials.AccessToken,
		UserAgent:         userAgent,
		ReadRateLimiter:   rate.NewLimiter(rate.Inf, 0),
		UpdateRateLimiter: rate.NewLimiter(rate.Inf, 0),
		maxRetries:        maxRetries,
		onRetry:           onRetry,
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

// DoRequest executes an HTTP request with authentication, retries and optional rate limiting.
// It adds the Bearer token and common headers, and returns *ErrClientResponse for non-2xx
// responses. A request rejected with HTTP 429 Too Many Requests is retried up to MaxRetries
// times, waiting as long as the server's Retry-After or X-RateLimit-Replenish-* headers ask,
// or with exponential backoff and jitter when they are absent. Retries on the same client
// are queued one wait apart, and new requests queue behind them, so concurrent callers do
// not race each other for the same replenished token. GET requests and other requests are
// queued separately, matching the API's separate limits. No other error is retried. Waits
// honor the request's context. ClientOptions.OnRetry, if set, is called before each retry.
func (c *Client) DoRequest(req *http.Request) ([]byte, error) {
	ctx := req.Context()
	rateLimiter, pause := c.UpdateRateLimiter, &c.pauseWrite
	if req.Method == http.MethodGet {
		rateLimiter, pause = c.ReadRateLimiter, &c.pauseRead
	}
	maxRetries := c.maxRetries
	switch {
	case maxRetries == 0:
		maxRetries = DefaultMaxRetries
	case maxRetries < 0:
		maxRetries = 0
	}
	// A body that cannot be re-read must not be replayed.
	canReplay := req.Body == nil || req.Body == http.NoBody || req.GetBody != nil

	c.setCommonHeaders(req)

	slot := pause.deadline()
	for attempt := 0; ; attempt++ {
		if err := waitUntil(ctx, slot); err != nil {
			return nil, err
		}
		if rateLimiter != nil {
			if err := rateLimiter.Wait(ctx); err != nil {
				return nil, err
			}
		}

		body, res, err := c.send(req)
		if err != nil {
			return nil, err
		}
		if res.StatusCode >= 200 && res.StatusCode < 300 {
			return body, nil
		}
		apiErr := &ErrClientResponse{status: res.StatusCode, body: string(body)}
		if res.StatusCode != http.StatusTooManyRequests || attempt >= maxRetries || !canReplay {
			return nil, apiErr
		}

		slot = pause.reserve(retryWait(res.Header, attempt))
		if c.onRetry != nil {
			c.onRetry(req, attempt+1, time.Until(slot))
		}
		if req.GetBody != nil {
			if req.Body, err = req.GetBody(); err != nil {
				return nil, err
			}
		}
	}
}

// send performs one round trip and reads the response body, bounded to DefaultMaxResponseSize.
// The returned response's body is already closed; only its status and headers may be used.
func (c *Client) send(req *http.Request) ([]byte, *http.Response, error) {
	res, err := c.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			// Log the error if you have a logger, otherwise this is acceptable for library code
			_ = closeErr
		}
	}()

	// Bound response body size to prevent memory exhaustion (CWE-400)
	limitedReader := io.LimitReader(res.Body, DefaultMaxResponseSize+1)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, nil, err
	}

	if int64(len(body)) > DefaultMaxResponseSize {
		return nil, nil, fmt.Errorf("%w: response exceeded %d bytes", ErrResponseTooLarge, DefaultMaxResponseSize)
	}

	return body, res, nil
}

// AssignLimits is a no-op kept for backward compatibility.
//
// Deprecated: the client no longer paces requests from X-RateLimit-* response headers.
// Requests rejected with HTTP 429 are retried instead; see ClientOptions.MaxRetries.
func (c *Client) AssignLimits(_ *http.Response, _ *rate.Limiter) error {
	return nil
}

// GetV1Url returns the base URL for CloudConnexa API v1 endpoints.
func (c *Client) GetV1Url() string {
	return c.BaseURL + "/api/v1"
}

// buildURL constructs a URL with escaped path segments for safe API calls.
// Example: buildURL(c.GetV1Url(), "users", userID, "activate")
// Returns: https://api.example.com/api/v1/users/{escaped-id}/activate
func buildURL(base string, segments ...string) string {
	if len(segments) == 0 {
		return base
	}
	escaped := make([]string, len(segments))
	for i, seg := range segments {
		escaped[i] = url.PathEscape(seg)
	}
	return base + "/" + strings.Join(escaped, "/")
}

// validateID returns an error if the provided ID is empty.
// This should be called before making API calls that require an ID parameter.
func validateID(id string) error {
	if id == "" {
		return ErrEmptyID
	}
	return nil
}
