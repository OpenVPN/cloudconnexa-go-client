# Cloud Connexa Go Client

[![GoDoc](https://img.shields.io/static/v1?label=godoc&message=reference&color=blue)](https://pkg.go.dev/github.com/openvpn/cloudconnexa-go-client/v2/cloudconnexa)
[![Go Report Card](https://goreportcard.com/badge/github.com/openvpn/cloudconnexa-go-client/v2)](https://goreportcard.com/report/github.com/openvpn/cloudconnexa-go-client/v2)
[![codecov](https://codecov.io/gh/openvpn/cloudconnexa-go-client/branch/main/graph/badge.svg)](https://codecov.io/gh/openvpn/cloudconnexa-go-client)
[![Build Status](https://github.com/openvpn/cloudconnexa-go-client/workflows/Go%20build/badge.svg)](https://github.com/openvpn/cloudconnexa-go-client/actions)

This Go library enables access to the Cloud Connexa API, as detailed in the [Cloud Connexa API Documentation](https://openvpn.net/cloud-docs/developer/cloudconnexa-api.html).

## Installation Instructions

To install the cloudconnexa-go-client, ensure you are using a modern Go release that supports module mode. With Go set up, execute the following command:

```sh
go get github.com/openvpn/cloudconnexa-go-client/v2/cloudconnexa
```

## Features

- Complete Cloud Connexa API coverage
- Pagination support
- Rate limiting
- Automatic token management
- Concurrent safe

## How to Use

In your Go project, you can use the library by importing it as follows:

```go
import "github.com/openvpn/cloudconnexa-go-client/v2/cloudconnexa"
```

Instantiate a new CloudConnexa client. Subsequently, utilize the diverse services provided by the client to interact with distinct segments of the CloudConnexa API. For instance:

```go
client := cloudconnexa.NewClient("api_url", "client_id", "client_secret")

// List connectors
connectors, _, err := client.Connectors.List()
```

## Authentication

For auth need to pass three parameters:

1. client_id
2. client_secret
3. api_url (example: `https://myorg.api.openvpn.com`)

```go
package main

import (
    "fmt"
    "log"

    "github.com/openvpn/cloudconnexa-go-client/v2/cloudconnexa"
)

func main() {
    client, err := cloudconnexa.NewClient("api_url", "client_id", "client_secret")
    if err != nil {
        log.Fatalf("error creating client: %v", err)
    }

    networkID := "your_network_id"
    routes, err := client.Routes.List(networkID)
    if err != nil {
        log.Fatalf("error getting routes: %v", err)
    }

    fmt.Println("Received routes:", routes)
}
```

## Examples

### Creating a Network

```go
network := cloudconnexa.Network{
    Name:           "test-network",
    Description:    "Test network created via API",
    InternetAccess: cloudconnexa.InternetAccessSplitTunnelOn,
    Egress:        false,
}

createdNetwork, err := client.Networks.Create(network)
```

### Managing Users

```go
// List all users
users, err := client.Users.List("", "")

// Create a new user
user := cloudconnexa.User{
    Username: "testuser",
    Email:    "test@example.com",
    GroupId:  "group-id",
}

createdUser, err := client.Users.Create(user)
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

Please make sure to update tests as appropriate.

## License

This project is licensed under the Apache License - see the [LICENSE](LICENSE) file for details.

## Security

For security issues, please email [security@openvpn.net](mailto:security@openvpn.net?subject=Security%20Issue%20in%20cloudconnexa-go-client%20github%20repository) instead of posting a public issue on GitHub.
