package cloudconnexa

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type DNSServers struct {
	PrimaryIPV4   string `json:"primaryIpV4,omitempty"`
	SecondaryIPV4 string `json:"secondaryIpV4,omitempty"`
}

type DNSZones struct {
	Zones []DNSZone `json:"zones"`
}

type DNSZone struct {
	Name      string   `json:"name"`
	Addresses []string `json:"addresses"`
}

type DomainRoutingSubnet struct {
	IPV4Address string `json:"ipV4Address"`
	IPV6Address string `json:"ipV6Address"`
}

type Subnet struct {
	IPV4Address []string `json:"ipV4Address"`
	IPV6Address []string `json:"ipV6Address"`
}

type SettingsService service

func (c *SettingsService) GetTrustedDevicesAllowed() (bool, error) {
	return c.getBool("/auth/trusted-devices-allowed")
}

func (c *SettingsService) SetTrustedDevicesAllowed(value bool) (bool, error) {
	return c.setBool("/auth/trusted-devices-allowed", value)
}

func (c *SettingsService) GetTwoFactorAuthEnabled() (bool, error) {
	return c.getBool("/auth/two-factor-auth")
}

func (c *SettingsService) SetTwoFactorAuthEnabled(value bool) (bool, error) {
	return c.setBool("/auth/two-factor-auth", value)
}

func (c *SettingsService) GetDNSServers() (*DNSServers, error) {
	body, err := c.get("/dns/custom-servers")
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

func (c *SettingsService) SetDNSServers(value *DNSServers) (*DNSServers, error) {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	body, err := c.set("/dns/custom-servers", jsonValue)
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

func (c *SettingsService) GetDefaultDNSSuffix() (string, error) {
	return c.getString("/dns/default-suffix")
}

func (c *SettingsService) SetDefaultDNSSuffix(value string) (string, error) {
	return c.setString("/dns/default-suffix", value)
}

func (c *SettingsService) GetDNSProxyEnabled() (bool, error) {
	return c.getBool("/dns/proxy-enabled")
}

func (c *SettingsService) SetDNSProxyEnabled(value bool) (bool, error) {
	return c.setBool("/dns/proxy-enabled", value)
}

func (c *SettingsService) GetDNSZones() ([]DNSZone, error) {
	body, err := c.get("/dns/zones")
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

func (c *SettingsService) SetDNSZones(value []DNSZone) ([]DNSZone, error) {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	body, err := c.set("/dns/zones", jsonValue)
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

func (c *SettingsService) GetDefaultConnectAuth() (string, error) {
	return c.getString("/user/connect-auth")
}

func (c *SettingsService) SetDefaultConnectAuth(value string) (string, error) {
	return c.setString("/user/connect-auth", value)
}

func (c *SettingsService) GetDefaultDeviceAllowancePerUser() (int, error) {
	return c.getInt("/user/device-allowance")
}

func (c *SettingsService) SetDefaultDeviceAllowancePerUser(value int) (int, error) {
	return c.setInt("/user/device-allowance", value)
}

func (c *SettingsService) GetForceUpdateDeviceAllowanceEnabled() (bool, error) {
	return c.getBool("/user/device-allowance-force-update")
}

func (c *SettingsService) SetForceUpdateDeviceAllowanceEnabled(value bool) (bool, error) {
	return c.setBool("/user/device-allowance-force-update", value)
}

func (c *SettingsService) GetDeviceEnforcement() (string, error) {
	return c.getString("/user/device-enforcement")
}

func (c *SettingsService) SetDeviceEnforcement(value string) (string, error) {
	return c.setString("/user/device-enforcement", value)
}

func (c *SettingsService) GetProfileDistribution() (string, error) {
	return c.getString("/user/profile-distribution")
}

func (c *SettingsService) SetProfileDistribution(value string) (string, error) {
	return c.setString("/user/profile-distribution", value)
}

func (c *SettingsService) GetConnectionTimeout() (int, error) {
	return c.getInt("/users/connection-timeout")
}

func (c *SettingsService) SetConnectionTimeout(value int) (int, error) {
	return c.setInt("/users/connection-timeout", value)
}

func (c *SettingsService) GetClientOptions() ([]string, error) {
	body, err := c.get("/wpc/client-options")
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

func (c *SettingsService) SetClientOptions(value []string) ([]string, error) {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	body, err := c.set("/wpc/client-options", jsonValue)
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

func (c *SettingsService) GetDefaultRegion() (string, error) {
	return c.getString("/wpc/default-region")
}

func (c *SettingsService) SetDefaultRegion(value string) (string, error) {
	return c.setString("/wpc/default-region", value)
}

func (c *SettingsService) GetDomainRoutingSubnet() (*DomainRoutingSubnet, error) {
	body, err := c.get("/wpc/domain-routing-subnet")
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

func (c *SettingsService) SetDomainRoutingSubnet(value DomainRoutingSubnet) (*DomainRoutingSubnet, error) {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	body, err := c.set("/wpc/domain-routing-subnet", jsonValue)
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

func (c *SettingsService) GetSnatEnabled() (bool, error) {
	return c.getBool("/wpc/snat")
}

func (c *SettingsService) SetSnatEnabled(value bool) (bool, error) {
	return c.setBool("/wpc/snat", value)
}

func (c *SettingsService) GetSubnet() (*Subnet, error) {
	body, err := c.get("/wpc/subnet")
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

func (c *SettingsService) SetSubnet(value Subnet) (*Subnet, error) {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	body, err := c.set("/wpc/subnet", jsonValue)
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

func (c *SettingsService) GetTopology() (string, error) {
	return c.getString("/wpc/topology")
}

func (c *SettingsService) SetTopology(value string) (string, error) {
	return c.setString("/wpc/topology", value)
}

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

func (c *SettingsService) getString(path string) (string, error) {
	endpoint := fmt.Sprintf("%s/settings"+path, c.client.GetV1Url())
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

func (c *SettingsService) setString(path string, value string) (string, error) {
	endpoint := fmt.Sprintf("%s/settings"+path, c.client.GetV1Url())
	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer([]byte(value)))
	req.Header.Set("Content-Type", "text/plain")
	if err != nil {
		return "", err
	}
	body, err := c.client.DoRequest(req)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func (c *SettingsService) getInt(path string) (int, error) {
	body, err := c.get(path)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(body))
}

func (c *SettingsService) setInt(path string, value int) (int, error) {
	body, err := c.set(path, []byte(strconv.Itoa(value)))
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(body))
}

func (c *SettingsService) get(path string) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/settings"+path, c.client.GetV1Url())
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

func (c *SettingsService) set(path string, value []byte) ([]byte, error) {
	endpoint := fmt.Sprintf("%s/settings"+path, c.client.GetV1Url())
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
