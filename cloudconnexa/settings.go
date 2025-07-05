package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

// DNSServers represents DNS server configuration with primary and secondary IPv4 addresses
type DNSServers struct {
	PrimaryIPV4   string `json:"primaryIpV4,omitempty"`
	SecondaryIPV4 string `json:"secondaryIpV4,omitempty"`
}

// DNSZones represents a collection of DNS zones
type DNSZones struct {
	Zones []DNSZone `json:"zones"`
}

// DNSZone represents a single DNS zone with name and associated addresses
type DNSZone struct {
	Name      string   `json:"name"`
	Addresses []string `json:"addresses"`
}

// DomainRoutingSubnet represents subnet configuration for domain routing with IPv4 and IPv6 addresses
type DomainRoutingSubnet struct {
	IPV4Address string `json:"ipV4Address"`
	IPV6Address string `json:"ipV6Address"`
}

// Subnet represents network subnet configuration with IPv4 and IPv6 address ranges
type Subnet struct {
	IPV4Address []string `json:"ipV4Address"`
	IPV6Address []string `json:"ipV6Address"`
}

// SettingsService handles operations related to system settings
type SettingsService service

// GetTrustedDevicesAllowed retrieves whether trusted devices are allowed
func (c *SettingsService) GetTrustedDevicesAllowed() (bool, error) {
	return c.getBool("%s/settings/auth/trusted-devices-allowed")
}

// SetTrustedDevicesAllowed sets whether trusted devices are allowed
func (c *SettingsService) SetTrustedDevicesAllowed(value bool) (bool, error) {
	return c.setBool("%s/settings/auth/trusted-devices-allowed", value)
}

// GetTwoFactorAuthEnabled retrieves whether two-factor authentication is enabled
func (c *SettingsService) GetTwoFactorAuthEnabled() (bool, error) {
	return c.getBool("%s/settings/auth/two-factor-auth")
}

// SetTwoFactorAuthEnabled sets whether two-factor authentication is enabled
func (c *SettingsService) SetTwoFactorAuthEnabled(value bool) (bool, error) {
	return c.setBool("%s/settings/auth/two-factor-auth", value)
}

// GetDNSServers retrieves the current DNS server configuration
func (c *SettingsService) GetDNSServers() (*DNSServers, error) {
	body, err := c.get("%s/settings/dns/custom-servers")
	if err != nil {
		return nil, err
	}

	var response DNSServers
	s := string(body)
	if s == "" {
		return nil, nil
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// SetDNSServers updates the DNS server configuration
func (c *SettingsService) SetDNSServers(value *DNSServers) (*DNSServers, error) {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	body, err := c.set("%s/settings/dns/custom-servers", jsonValue)
	if err != nil {
		return nil, err
	}
	var response DNSServers
	s := string(body)
	if s == "" {
		return nil, nil
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// GetDefaultDNSSuffix retrieves the default DNS suffix
func (c *SettingsService) GetDefaultDNSSuffix() (string, error) {
	return c.getString("%s/settings/dns/default-suffix")
}

// SetDefaultDNSSuffix sets the default DNS suffix
func (c *SettingsService) SetDefaultDNSSuffix(value string) (string, error) {
	return c.setString("%s/settings/dns/default-suffix", value)
}

// GetDNSProxyEnabled retrieves whether DNS proxy is enabled
func (c *SettingsService) GetDNSProxyEnabled() (bool, error) {
	return c.getBool("%s/settings/dns/proxy-enabled")
}

// SetDNSProxyEnabled sets whether DNS proxy is enabled
func (c *SettingsService) SetDNSProxyEnabled(value bool) (bool, error) {
	return c.setBool("%s/settings/dns/proxy-enabled", value)
}

// GetDNSZones retrieves the current DNS zones configuration
func (c *SettingsService) GetDNSZones() ([]DNSZone, error) {
	body, err := c.get("%s/settings/dns/zones")
	if err != nil {
		return nil, err
	}

	var response DNSZones
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return response.Zones, nil
}

// SetDNSZones updates the DNS zones configuration
func (c *SettingsService) SetDNSZones(value []DNSZone) ([]DNSZone, error) {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	body, err := c.set("%s/settings/dns/zones", jsonValue)
	if err != nil {
		return nil, err
	}
	var response DNSZones
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return response.Zones, nil
}

// GetDefaultConnectAuth retrieves the default connection authentication method
func (c *SettingsService) GetDefaultConnectAuth() (string, error) {
	return c.getString("%s/settings/user/connect-auth")
}

// SetDefaultConnectAuth sets the default connection authentication method
func (c *SettingsService) SetDefaultConnectAuth(value string) (string, error) {
	return c.setString("%s/settings/user/connect-auth", value)
}

// GetDefaultDeviceAllowancePerUser retrieves the default device allowance per user
func (c *SettingsService) GetDefaultDeviceAllowancePerUser() (int, error) {
	return c.getInt("%s/settings/user/device-allowance")
}

// SetDefaultDeviceAllowancePerUser sets the default device allowance per user
func (c *SettingsService) SetDefaultDeviceAllowancePerUser(value int) (int, error) {
	return c.setInt("%s/settings/user/device-allowance", value)
}

// GetForceUpdateDeviceAllowanceEnabled retrieves whether force update device allowance is enabled
func (c *SettingsService) GetForceUpdateDeviceAllowanceEnabled() (bool, error) {
	return c.getBool("%s/settings/user/device-allowance-force-update")
}

// SetForceUpdateDeviceAllowanceEnabled sets whether force update device allowance is enabled
func (c *SettingsService) SetForceUpdateDeviceAllowanceEnabled(value bool) (bool, error) {
	return c.setBool("%s/settings/user/device-allowance-force-update", value)
}

// GetDeviceEnforcement retrieves the device enforcement policy
func (c *SettingsService) GetDeviceEnforcement() (string, error) {
	return c.getString("%s/settings/user/device-enforcement")
}

// SetDeviceEnforcement sets the device enforcement policy
func (c *SettingsService) SetDeviceEnforcement(value string) (string, error) {
	return c.setString("%s/settings/user/device-enforcement", value)
}

// GetProfileDistribution retrieves the profile distribution method
func (c *SettingsService) GetProfileDistribution() (string, error) {
	return c.getString("%s/settings/user/profile-distribution")
}

// SetProfileDistribution sets the profile distribution method
func (c *SettingsService) SetProfileDistribution(value string) (string, error) {
	return c.setString("%s/settings/user/profile-distribution", value)
}

// GetConnectionTimeout retrieves the connection timeout value
func (c *SettingsService) GetConnectionTimeout() (int, error) {
	return c.getInt("%s/settings/users/connection-timeout")
}

// SetConnectionTimeout sets the connection timeout value
func (c *SettingsService) SetConnectionTimeout(value int) (int, error) {
	return c.setInt("%s/settings/users/connection-timeout", value)
}

// GetClientOptions retrieves the client options configuration
func (c *SettingsService) GetClientOptions() ([]string, error) {
	body, err := c.get("%s/settings/wpc/client-options")
	if err != nil {
		return nil, err
	}

	var response []string
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// SetClientOptions updates the client options configuration
func (c *SettingsService) SetClientOptions(value []string) ([]string, error) {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	body, err := c.set("%s/settings/wpc/client-options", jsonValue)
	if err != nil {
		return nil, err
	}
	var response []string
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

// GetDefaultRegion retrieves the default region setting
func (c *SettingsService) GetDefaultRegion() (string, error) {
	return c.getString("%s/settings/wpc/default-region")
}

// SetDefaultRegion sets the default region
func (c *SettingsService) SetDefaultRegion(value string) (string, error) {
	return c.setString("%s/settings/wpc/default-region", value)
}

// GetDomainRoutingSubnet retrieves the domain routing subnet configuration
func (c *SettingsService) GetDomainRoutingSubnet() (*DomainRoutingSubnet, error) {
	body, err := c.get("%s/settings/wpc/domain-routing-subnet")
	if err != nil {
		return nil, err
	}

	var response DomainRoutingSubnet
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// SetDomainRoutingSubnet updates the domain routing subnet configuration
func (c *SettingsService) SetDomainRoutingSubnet(value DomainRoutingSubnet) (*DomainRoutingSubnet, error) {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	body, err := c.set("%s/settings/wpc/domain-routing-subnet", jsonValue)
	if err != nil {
		return nil, err
	}
	var response DomainRoutingSubnet
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// GetSnatEnabled retrieves whether SNAT is enabled
func (c *SettingsService) GetSnatEnabled() (bool, error) {
	return c.getBool("%s/settings/wpc/snat")
}

// SetSnatEnabled sets whether SNAT is enabled
func (c *SettingsService) SetSnatEnabled(value bool) (bool, error) {
	return c.setBool("%s/settings/wpc/snat", value)
}

// GetSubnet retrieves the subnet configuration
func (c *SettingsService) GetSubnet() (*Subnet, error) {
	body, err := c.get("%s/settings/wpc/subnet")
	if err != nil {
		return nil, err
	}

	var response Subnet
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// SetSubnet updates the subnet configuration
func (c *SettingsService) SetSubnet(value Subnet) (*Subnet, error) {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	body, err := c.set("%s/settings/wpc/subnet", jsonValue)
	if err != nil {
		return nil, err
	}
	var response Subnet
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

// GetTopology retrieves the network topology setting
func (c *SettingsService) GetTopology() (string, error) {
	return c.getString("%s/settings/wpc/topology")
}

// SetTopology sets the network topology
func (c *SettingsService) SetTopology(value string) (string, error) {
	return c.setString("%s/settings/wpc/topology", value)
}

// GetDnsLogEnabled retrieves whether DNS Log is enabled
func (c *SettingsService) GetDnsLogEnabled() (bool, error) {
	return c.getBool("%s/dns-log/user-dns-resolutions/enabled")
}

// SetDnsLogEnabled sets whether DNS Log is enabled
func (c *SettingsService) SetDnsLogEnabled(value bool) error {
	if value {
		_, err := c.set("%s/dns-log/user-dns-resolutions/enable", []byte(strconv.FormatBool(value)))
		if err != nil {
			return err
		}
	} else {
		_, err := c.set("%s/dns-log/user-dns-resolutions/disable", []byte(strconv.FormatBool(value)))
		if err != nil {
			return err
		}
	}
	return nil
}

// GetAccessVisibilityEnabled retrieves whether Access Visibility is enabled
func (c *SettingsService) GetAccessVisibilityEnabled() (bool, error) {
	return c.getBool("%s/access-visibility/enabled")
}

// SetAccessVisibilityEnabled sets whether Access Visibility is enabled
func (c *SettingsService) SetAccessVisibilityEnabled(value bool) error {
	if value {
		_, err := c.set("%s/access-visibility/enable", []byte(strconv.FormatBool(value)))
		if err != nil {
			return err
		}
	} else {
		_, err := c.set("%s/access-visibility/disable", []byte(strconv.FormatBool(value)))
		if err != nil {
			return err
		}
	}
	return nil
}

// getBool retrieves a boolean value from the specified path
func (c *SettingsService) getBool(path string) (bool, error) {
	body, err := c.get(path)
	if err != nil {
		return false, err
	}

	var response bool
	err = json.Unmarshal(body, &response)
	if err != nil {
		return false, err
	}
	return response, nil
}

// setBool sets a boolean value at the specified path
func (c *SettingsService) setBool(path string, value bool) (bool, error) {
	body, err := c.set(path, []byte(strconv.FormatBool(value)))
	if err != nil {
		return false, err
	}
	var response bool
	err = json.Unmarshal(body, &response)
	if err != nil {
		return false, err
	}
	return response, nil
}

// getString retrieves a string value from the specified path
func (c *SettingsService) getString(path string) (string, error) {
	endpoint := fmt.Sprintf(path, c.client.GetV1Url())
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	body, err := c.client.DoRequest(req)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// setString sets a string value at the specified path
func (c *SettingsService) setString(path string, value string) (string, error) {
	endpoint := fmt.Sprintf(path, c.client.GetV1Url())
	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer([]byte(value)))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "text/plain")
	body, err := c.client.DoRequest(req)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// getInt retrieves an integer value from the specified path
func (c *SettingsService) getInt(path string) (int, error) {
	body, err := c.get(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(body))
}

// setInt sets an integer value at the specified path
func (c *SettingsService) setInt(path string, value int) (int, error) {
	body, err := c.set(path, []byte(strconv.Itoa(value)))
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(body))
}

// get performs a GET request to the specified path
func (c *SettingsService) get(path string) ([]byte, error) {
	endpoint := fmt.Sprintf(path, c.client.GetV1Url())
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}
	return body, nil
}

// set performs a PUT request to the specified path with the given value
func (c *SettingsService) set(path string, value []byte) ([]byte, error) {
	endpoint := fmt.Sprintf(path, c.client.GetV1Url())
	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(value))
	if err != nil {
		return nil, err
	}
	body, err := c.client.DoRequest(req)
	if err != nil {
		return nil, err
	}
	return body, nil
}
