# Cloud Connexa Go Client

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/openvpn/cloudconnexa-go-client/v2/cloudconnexa)
[![codecov](https://codecov.io/gh/openvpn/cloudconnexa-go-client/branch/main/graph/badge.svg)](https://codecov.io/gh/openvpn/cloudconnexa-go-client)
[![Build Status](https://github.com/openvpn/cloudconnexa-go-client/workflows/Go%20build/badge.svg)](https://github.com/openvpn/cloudconnexa-go-client/actions)

The official Go client library for the Cloud Connexa API provides programmatic access to OpenVPN Cloud Connexa services.

**Full CloudConnexa API v1.1.0 Support** - Complete coverage of all public API endpoints with modern Go patterns.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [Authentication](#authentication)
- [Usage Examples](#usage-examples)
- [API Coverage](#api-coverage)
- [Configuration](#configuration)
- [Testing](#testing)
- [Contributing](#contributing)
- [Versioning](#versioning)
- [License](#license)
- [Support](#support)

## Installation

Requires Go 1.23 or later.

```bash
go get github.com/openvpn/cloudconnexa-go-client/v2/cloudconnexa
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/openvpn/cloudconnexa-go-client/v2/cloudconnexa"
)

func main() {
    client, err := cloudconnexa.NewClient("https://myorg.api.openvpn.com", "client_id", "client_secret")
    if err != nil {
        log.Fatal(err)
    }

    // List networks
    networks, err := client.Networks.List()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d networks\n", len(networks))
}
```

## Authentication

The client requires three parameters for authentication:

- `api_url`: Your organisation's API endpoint (e.g., `https://myorg.api.openvpn.com`)
- `client_id`: OAuth2 client ID
- `client_secret`: OAuth2 client secret

```go
client, err := cloudconnexa.NewClient(apiURL, clientID, clientSecret)
if err != nil {
    return err
}
```

## Usage Examples

### Network Management

```go
// Create a network
network := cloudconnexa.Network{
    Name:           "production-network",
    Description:    "Production environment network",
    InternetAccess: cloudconnexa.InternetAccessSplitTunnelOn,
    Egress:         true,
}

createdNetwork, err := client.Networks.Create(network)
if err != nil {
    log.Fatal(err)
}

// List networks with pagination
networks, pagination, err := client.Networks.GetByPage(1, 10)
if err != nil {
    log.Fatal(err)
}

// Update a network
updatedNetwork, err := client.Networks.Update(networkID, network)
if err != nil {
    log.Fatal(err)
}

// Delete a network
err = client.Networks.Delete(networkID)
if err != nil {
    log.Fatal(err)
}
```

### User Management

```go
// Create a user
user := cloudconnexa.User{
    Username:  "john.doe",
    Email:     "john.doe@company.com",
    FirstName: "John",
    LastName:  "Doe",
    GroupID:   "group-123",
}

createdUser, err := client.Users.Create(user)
if err != nil {
    log.Fatal(err)
}

// List users with filtering
users, err := client.Users.List("", "active")
if err != nil {
    log.Fatal(err)
}

// Get user by ID
user, err := client.Users.GetByID(userID)
if err != nil {
    log.Fatal(err)
}
```

### Connector Management

```go
// List connectors
connectors, err := client.Connectors.List()
if err != nil {
    log.Fatal(err)
}

// Create a connector
connector := cloudconnexa.Connector{
    Name:        "office-connector",
    Description: "Main office connector",
    NetworkID:   networkID,
}

createdConnector, err := client.Connectors.Create(connector)
if err != nil {
    log.Fatal(err)
}
```

### Host Management

```go
// List hosts
hosts, err := client.Hosts.List()
if err != nil {
    log.Fatal(err)
}

// Get host by ID
host, err := client.Hosts.GetByID(hostID)
if err != nil {
    log.Fatal(err)
}
```

### DNS Records

```go
// List DNS records
dnsRecords, err := client.DNSRecords.List()
if err != nil {
    log.Fatal(err)
}

// Create DNS record
record := cloudconnexa.DNSRecord{
    Domain:      "api.internal.company.com",
    Description: "Internal API endpoint",
    IPAddress:   "10.0.1.100",
}

createdRecord, err := client.DNSRecords.Create(record)
if err != nil {
    log.Fatal(err)
}
```

## API Coverage

The client provides **100% coverage** of the CloudConnexa API v1.1.0 with all public endpoints:

### **Core Resources**

- **Networks** - Complete network lifecycle management (CRUD operations)
- **Users** - User management, authentication, and device associations
- **User Groups** - Group policies, permissions, and access control
- **VPN Regions** - Available VPN server regions and capabilities

### **Connectivity & Infrastructure**

- **Network Connectors** - Site-to-site connectivity with IPsec tunnel support
- **Host Connectors** - Host-based connectivity and routing
- **Hosts** - Host configuration, monitoring, and IP services
- **Routes** - Network routing configuration and management

### **Services & Monitoring**

- **DNS Records** - Private DNS management with direct endpoint access
- **Host IP Services** - Service definitions and port configurations
- **Sessions** - OpenVPN session monitoring and analytics
- **Devices** - Device lifecycle management and security controls

### **Security & Access Control**

- **Access Groups** - Fine-grained access policies and rules
- **Location Contexts** - Location-based access controls
- **Settings** - System-wide configuration and preferences

### **API v1.1.0 Features**

- **Direct Endpoints**: Optimised single-call access for DNS Records and User Groups
- **Enhanced Sessions API**: Complete OpenVPN session monitoring with cursor-based pagination
- **Comprehensive Devices API**: Full device management with filtering and bulk operations
- **IPsec Support**: Start/stop IPsec tunnels for Network Connectors
- **Updated DTOs**: Simplified data structures aligned with API v1.1.0

### **All Endpoints Support**

- **Pagination** - Both cursor-based (Sessions) and page-based (legacy) pagination
- **Error Handling** - Structured error types with detailed messages
- **Rate Limiting** - Automatic rate limiting with configurable limits
- **Type Safety** - Strong typing with comprehensive validation
- **Concurrent Safety** - Thread-safe operations for production use
- **Performance Optimized** - Direct API calls where available

## Configuration

### Rate Limiting

The client includes built-in rate limiting to respect API limits:

```go
// Rate limiting is automatic, no configuration needed
client, err := cloudconnexa.NewClient(apiURL, clientID, clientSecret)
```

### Custom HTTP Client

```go
import (
    "net/http"
    "time"
)

// Use custom HTTP client with timeout
httpClient := &http.Client{
    Timeout: 30 * time.Second,
}

client, err := cloudconnexa.NewClient(apiURL, clientID, clientSecret)
// Client uses default HTTP client with sensible timeouts
```

### Error Handling

The client provides structured error types:

```go
networks, err := client.Networks.List()
if err != nil {
    if clientErr, ok := err.(*cloudconnexa.ErrClientResponse); ok {
        fmt.Printf("API Error: %d - %s\n", clientErr.StatusCode, clientErr.Message)
    } else {
        fmt.Printf("Network Error: %v\n", err)
    }
}
```

## Testing

### Unit Tests

```bash
# Run unit tests
make test

# Run tests with coverage
go test -v -race -coverprofile=coverage.txt ./cloudconnexa/...
```

### End-to-End Tests

```bash
# Run e2e tests (requires API credentials)
export CLOUDCONNEXA_BASE_URL="https://your-org.api.openvpn.com"
export CLOUDCONNEXA_CLIENT_ID="your-client-id"
export CLOUDCONNEXA_CLIENT_SECRET="your-client-secret"

make e2e
```

### Linting

```bash
# Run linters
make lint

# Install golangci-lint if needed
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### Development Setup

1. Fork the repository
2. Clone your fork
3. Install dependencies: `make deps`
4. Run tests: `make test`
5. Run linters: `make lint`
6. Submit a Pull Request

### Code Standards

- Follow Go conventions and best practices
- Write comprehensive tests for new features
- Update documentation for API changes
- Use meaningful commit messages

## Versioning

This project follows [Semantic Versioning](https://semver.org/):

- **Major version**: Breaking API changes
- **Minor version**: New features, backward compatible
- **Patch version**: Bug fixes, backward compatible

Current version: `v2.x.x`

### Changelog

See [Releases](https://github.com/openvpn/cloudconnexa-go-client/releases) for the detailed changelog.

## License

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) file for details.

## Support

### Documentation

- [Cloud Connexa API Documentation](https://openvpn.net/cloud-docs/developer/cloudconnexa-api-v1-1-0.html)
- [Go Package Documentation](https://pkg.go.dev/github.com/openvpn/cloudconnexa-go-client/v2/cloudconnexa)

### Issues and Questions

- **Bug reports**: [GitHub Issues](https://github.com/openvpn/cloudconnexa-go-client/issues)
- **Feature requests**: [GitHub Issues](https://github.com/openvpn/cloudconnexa-go-client/issues)
- **Security issues**: Email [security@openvpn.net](mailto:security@openvpn.net?subject=Security%20Issue%20in%20cloudconnexa-go-client)

### Requirements

- Go 1.23 or later
- Valid Cloud Connexa API credentials
- Network access to Cloud Connexa API endpoints
