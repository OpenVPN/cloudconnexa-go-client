# CloudConnexa Go Client

This Go library enables access to the CloudConnexa API, as detailed in the [CloudConnexa API Documentation](https://openvpn.net/cloud-docs/developer/cloudconnexa-api.html).

## Installation Instructions

To install the cloudconnexa-go-client, ensure you are using a modern Go release that supports module mode. With Go set up, execute the following command:

```sh
go get github.com/openvpn/cloudconnexa-go-client/v2/cloudconnexa
```

## How to Use

In your Go project, you can use the library by importing it as follows:

```go
import "github.com/openvpn/cloudconnexa-go-client/v2/cloudconnexa"
```

Instantiate a new CloudConnexa client. Subsequently, utilize the diverse services provided by the client to interact with distinct segments of the CloudConnexa API. For instance:

```go
client := cloudconnexa.NewClient("api_url", "client_id", "client_secret")

// List connectors
connectors, _, err := client.Connectors.List(ctx, nil)
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

	networkId := "your_network_id"
	routes, err := client.Routes.List(networkId)
	if err != nil {
		log.Fatalf("error getting routes: %v", err)
	}

	fmt.Println("Received routes:", routes)
}
```
