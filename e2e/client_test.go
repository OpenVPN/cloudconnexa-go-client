package client

import (
	"crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	mrand "math/rand"
	"net"
	"os"
	"testing"
	"time"

	"github.com/openvpn/cloudconnexa-go-client/v2/cloudconnexa"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// validateEnvVars checks if all required environment variables are set
func validateEnvVars(t *testing.T) {
	validateEnvVar(t, HostEnvVar)
	validateEnvVar(t, ClientIDEnvVar)
	validateEnvVar(t, ClientSecretEnvVar)
}

// validateEnvVar checks if a specific environment variable is set
// envVar: The name of the environment variable to check
func validateEnvVar(t *testing.T, envVar string) {
	fmt.Println(os.Getenv(envVar))
	require.NotEmptyf(t, os.Getenv(envVar), "%s must be set", envVar)
}

// Environment variable names for client configuration
const (
	HostEnvVar         = "CLOUDCONNEXA_BASE_URL"
	ClientIDEnvVar     = "CLOUDCONNEXA_CLIENT_ID"
	ClientSecretEnvVar = "CLOUDCONNEXA_CLIENT_SECRET" //nolint:gosec // This is an environment variable name, not a credential
)

// TestNewClient tests the creation of a new client
// It verifies that the client is created successfully and has a valid token
func TestNewClient(t *testing.T) {
	c := setUpClient(t)
	assert.NotEmpty(t, c.Token)
}

// setUpClient creates and returns a new client for testing
// It validates environment variables and initializes the client with credentials
func setUpClient(t *testing.T) *cloudconnexa.Client {
	validateEnvVars(t)
	var err error
	client, err := cloudconnexa.NewClient(os.Getenv(HostEnvVar), os.Getenv(ClientIDEnvVar), os.Getenv(ClientSecretEnvVar))
	require.NoError(t, err)
	return client
}

// cidrOverlaps returns true if two IPv4 networks overlap (IPv4 only, safe)
func cidrOverlaps(a *net.IPNet, b *net.IPNet) bool {
	if a == nil || b == nil {
		return false
	}
	if a.IP.To4() == nil || b.IP.To4() == nil {
		return false
	}
	return a.Contains(b.IP) || b.Contains(a.IP)
}

// parseCIDROrNil parses CIDR string and returns *net.IPNet or nil on error
func parseCIDROrNil(cidr string) *net.IPNet {
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil
	}
	return ipnet
}

func shuffledRange(start, end int, rnd *mrand.Rand) []int {
	n := end - start + 1
	arr := make([]int, n)
	for i := 0; i < n; i++ {
		arr[i] = start + i
	}
	rnd.Shuffle(n, func(i, j int) { arr[i], arr[j] = arr[j], arr[i] })
	return arr
}

// findAvailableInRange scans used subnets and returns a free 10.a.b.0/24 within [startA, endA]
func findAvailableInRange(used []*net.IPNet, startA, endA int, rnd *mrand.Rand) (string, bool) {
	// Skip the commonly reserved 10.200.0.0/16 range first
	reserved := parseCIDROrNil("10.200.0.0/16")
	for _, a := range shuffledRange(startA, endA, rnd) {
		for _, b := range shuffledRange(0, 255, rnd) {
			candidate := fmt.Sprintf("10.%d.%d.0/24", a, b)
			_, ipn, err := net.ParseCIDR(candidate)
			if err != nil {
				continue
			}
			if reserved != nil && cidrOverlaps(ipn, reserved) {
				continue
			}
			overlap := false
			for _, u := range used {
				if cidrOverlaps(ipn, u) {
					overlap = true
					break
				}
			}
			if !overlap {
				return candidate, true
			}
		}
	}
	return "", false
}

// findAvailableInRange172 scans used subnets for a free 172.16-31.b.0/24
func findAvailableInRange172(used []*net.IPNet, rnd *mrand.Rand) (string, bool) {
	for _, a := range shuffledRange(16, 31, rnd) {
		for _, b := range shuffledRange(0, 255, rnd) {
			candidate := fmt.Sprintf("172.%d.%d.0/24", a, b)
			_, ipn, err := net.ParseCIDR(candidate)
			if err != nil {
				continue
			}
			overlap := false
			for _, u := range used {
				if cidrOverlaps(ipn, u) {
					overlap = true
					break
				}
			}
			if !overlap {
				return candidate, true
			}
		}
	}
	return "", false
}

// findAvailableInRange192168 scans used subnets for a free 192.168.b.0/24
func findAvailableInRange192168(used []*net.IPNet, rnd *mrand.Rand) (string, bool) {
	for _, b := range shuffledRange(0, 255, rnd) {
		candidate := fmt.Sprintf("192.168.%d.0/24", b)
		_, ipn, err := net.ParseCIDR(candidate)
		if err != nil {
			continue
		}
		overlap := false
		for _, u := range used {
			if cidrOverlaps(ipn, u) {
				overlap = true
				break
			}
		}
		if !overlap {
			return candidate, true
		}
	}
	return "", false
}

// findAvailableIPv4Subnet scans existing networks' routes and system subnets
// and returns an available RFC1918 /24 subnet that does not overlap
func findAvailableIPv4Subnet(c *cloudconnexa.Client) (string, error) {
	var seedBytes [8]byte
	_, _ = rand.Read(seedBytes[:])
	rnd := mrand.New(mrand.NewSource(int64(binary.LittleEndian.Uint64(seedBytes[:]))))

	networks, err := c.Networks.List()
	if err != nil {
		return "", err
	}

	var used []*net.IPNet
	for _, n := range networks {
		// Collect existing routes via API to ensure we see them
		routes, err := c.Routes.List(n.ID)
		if err == nil {
			for _, r := range routes {
				if r.Subnet == "" {
					continue
				}
				if ipn := parseCIDROrNil(r.Subnet); ipn != nil {
					used = append(used, ipn)
				}
			}
		}
		// Collect system subnets from GET network (may not be present in List)
		if nn, err := c.Networks.Get(n.ID); err == nil && nn != nil {
			for _, s := range nn.SystemSubnets {
				if ipn := parseCIDROrNil(s); ipn != nil {
					used = append(used, ipn)
				}
			}
		}
	}

	// Try 10.0.0.0/8 excluding known reserved 10.200.0.0/16, prefer higher ranges
	if candidate, ok := findAvailableInRange(used, 201, 254, rnd); ok {
		return candidate, nil
	}
	if candidate, ok := findAvailableInRange(used, 0, 199, rnd); ok {
		return candidate, nil
	}
	if candidate, ok := findAvailableInRange(used, 200, 200, rnd); ok {
		return candidate, nil
	}
	// Try 172.16.0.0/12
	if candidate, ok := findAvailableInRange172(used, rnd); ok {
		return candidate, nil
	}
	// Try 192.168.0.0/16
	if candidate, ok := findAvailableInRange192168(used, rnd); ok {
		return candidate, nil
	}

	return "", fmt.Errorf("no available /24 subnet found in RFC1918 ranges")
}

// TestListNetworks tests the retrieval of networks using pagination
// It verifies that networks can be retrieved successfully
func TestListNetworks(t *testing.T) {
	c := setUpClient(t)
	response, err := c.Networks.GetByPage(0, 100)
	require.NoError(t, err)
	fmt.Printf("found %d networks\n", len(response.Content))
}

// TestListConnectors tests the retrieval of network connectors using pagination
// It verifies that connectors can be retrieved successfully
func TestListConnectors(t *testing.T) {
	c := setUpClient(t)
	response, err := c.NetworkConnectors.GetByPage(0, 10)
	require.NoError(t, err)
	fmt.Printf("found %d connectors\n", len(response.Content))
}

// TestVPNRegions tests the VPN regions functionality
// It verifies that regions can be listed and retrieved by ID
func TestVPNRegions(t *testing.T) {
	c := setUpClient(t)

	// Test List
	regions, err := c.VPNRegions.List()
	require.NoError(t, err)
	require.NotNil(t, regions)
	fmt.Printf("found %d VPN regions\n", len(regions))

	// If regions exist, test GetByID
	if len(regions) > 0 {
		region := regions[0]
		foundRegion, err := c.VPNRegions.GetByID(region.ID)
		require.NoError(t, err)
		require.NotNil(t, foundRegion)
		require.Equal(t, region.ID, foundRegion.ID)
		require.Equal(t, region.Country, foundRegion.Country)
		require.Equal(t, region.Continent, foundRegion.Continent)
		fmt.Printf("successfully found region %s in %s, %s\n",
			foundRegion.ID, foundRegion.Country, foundRegion.Continent)
	}

	// Test GetByID with non-existent ID
	nonExistentRegion, err := c.VPNRegions.GetByID("non-existent-id")
	require.NoError(t, err)
	require.Nil(t, nonExistentRegion)
}

// TestCreateNetwork tests the creation of a network with associated resources
// It verifies that a network can be created with routes and services, and then deleted
func TestCreateNetwork(t *testing.T) {
	c := setUpClient(t)
	timestamp := time.Now().Unix()
	testName := fmt.Sprintf("test-%d-%d", timestamp, time.Now().Nanosecond())

	networks, err := c.Networks.List()
	require.NoError(t, err)
	for _, n := range networks {
		require.NotEqual(t, testName, n.Name, "Network with name %s already exists", testName)
	}

	connector := cloudconnexa.NetworkConnector{
		Description: "test",
		Name:        testName,
		VpnRegionID: "it-mxp",
	}

	network := cloudconnexa.Network{
		Description:       "test",
		Egress:            false,
		Name:              testName,
		InternetAccess:    cloudconnexa.InternetAccessSplitTunnelOn,
		Connectors:        []cloudconnexa.NetworkConnector{connector},
		TunnelingProtocol: "OPENVPN",
	}
	response, err := c.Networks.Create(network)
	require.NoError(t, err)
	fmt.Printf("created %s network\n", response.ID)
	// Ensure cleanup even if subsequent steps fail
	defer func() { _ = c.Networks.Delete(response.ID) }()

	// Attempt to create a non-overlapping route with retries to avoid CI matrix collisions
	var testRoute *cloudconnexa.Route
	var lastErr error
	for attempts := 0; attempts < 20; attempts++ {
		subnet, serr := findAvailableIPv4Subnet(c)
		require.NoError(t, serr)
		route := cloudconnexa.Route{
			Description: "test",
			Type:        "IP_V4",
			Subnet:      subnet,
		}
		testRoute, err = c.Routes.Create(response.ID, route)
		if err == nil {
			fmt.Printf("created %s route\n", testRoute.ID)
			break
		}
		lastErr = err
		var apiErr *cloudconnexa.ErrClientResponse
		if errors.As(err, &apiErr) {
			if apiErr.StatusCode() == 400 {
				// Overlap or validation error, refresh and retry
				time.Sleep(500 * time.Millisecond)
				continue
			}
		}
		// Unexpected error
		require.NoError(t, err)
	}
	require.NoError(t, lastErr)
	require.NotNil(t, testRoute)

	serviceConfig := cloudconnexa.IPServiceConfig{
		ServiceTypes: []string{"ANY"},
	}
	ipServiceRoute := cloudconnexa.IPServiceRoute{
		Description: "test",
		Value:       testRoute.Subnet,
	}
	service := cloudconnexa.IPService{
		Name:            testName,
		Description:     "test",
		NetworkItemID:   response.ID,
		Type:            "IP_SOURCE",
		NetworkItemType: "NETWORK",
		Config:          &serviceConfig,
		Routes:          []*cloudconnexa.IPServiceRoute{&ipServiceRoute},
	}
	s, err := c.NetworkIPServices.Create(&service)
	require.NoError(t, err)
	fmt.Printf("created %s service\n", s.ID)
	err = c.Networks.Delete(response.ID)
	require.NoError(t, err)
	fmt.Printf("deleted %s network\n", response.ID)
}
