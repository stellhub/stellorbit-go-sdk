# StellOrbit Go SDK

`stellorbit-go-sdk` is the Go client SDK for [`stellhub/stellorbit-service`](https://github.com/stellhub/stellorbit-service), the Stell service governance engine responsible for routing, load balancing, retries, traffic shifting, and service lifecycle policies.

## Positioning

This repository provides the Go client implementation for applications, platform services, and middleware components that need to consume StellOrbit governance capabilities.

It does not implement governance rules locally. The SDK delegates policy decisions to `stellorbit-service` and keeps Go applications aligned with the central Stell service governance control plane.

## Capabilities

- Route decision requests for service-to-service traffic.
- Service lifecycle policy lookup.
- Traffic governance policy lookup.
- API key based authentication header support.
- Standard Go `net/http` transport without third-party runtime dependencies.
- Timeout configuration for request execution.
- Context propagation for all client calls.

## Current Status

| Item | Value |
| --- | --- |
| Stability | Early development |
| Language | Go |
| Transport | `net/http` |
| Target service | `stellorbit-service` |
| Maintainer | StellHub |

## Installation

```bash
go get github.com/stellhub/stellorbit-go-sdk
```

## Quick Start

```go
package main

import (
	"context"
	"fmt"
	"log"

	stellorbit "github.com/stellhub/stellorbit-go-sdk"
)

func main() {
	client, err := stellorbit.NewClient(stellorbit.Options{
		Endpoint: "http://localhost:8080",
		APIKey:   "local-dev-api-key",
	})
	if err != nil {
		log.Fatal(err)
	}

	response, err := client.Route(context.Background(), stellorbit.RouteRequest{
		ServiceName: "payment-service",
		RouteKey:    "tenant-a",
		Attributes: map[string]string{
			"env":    "dev",
			"region": "local",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(response.Body)
}
```

## API Surface

| Method | Responsibility |
| --- | --- |
| `Route(ctx, request)` | Request a route decision from StellOrbit |
| `LifecyclePolicy(ctx, serviceName)` | Fetch service lifecycle governance policy |
| `TrafficPolicy(ctx, serviceName)` | Fetch routing, retry, and traffic shifting policy |

## Development

Run verification:

```bash
go test ./...
```

## Repository Scope

This SDK intentionally keeps the first version small. Future releases can add:

- Strongly typed policy models.
- Retry and failover helpers.
- OpenAPI generated DTOs when the service contract is stable.
- Integration tests against local `stellorbit-service`.
- Observability hooks for client-side metrics and tracing.

## License

The license will be defined before the first stable release.
