# StellOrbit Go SDK

`stellorbit-go-sdk` is the Go client SDK for [`stellhub/stellorbit-service`](https://github.com/stellhub/stellorbit-service), the Stell service governance engine responsible for routing, load balancing, retries, traffic shifting, and service lifecycle policies.

It consumes governance rules from [`stellnula-service`](https://github.com/stellhub/stellnula-service) through [`stellnula-go-sdk`](https://github.com/stellhub/stellnula-go-sdk), keeps a local immutable rule registry, and exposes rule providers for Go applications and middleware adapters.

## Positioning

This repository provides the Go client implementation for applications, platform services, and middleware components that need to consume StellOrbit governance capabilities.

It does not implement circuit breaker state machines, rate limit algorithms, authorization interceptors, or routing executors in the core package. Those actions belong to application middleware, HTTP handlers, gRPC interceptors, gateways, or framework adapters that consume the providers.

## Capabilities

- Bootstrap governance rules from StellNula.
- Subscribe to the `governance/service-governance` rule channel.
- Build a local immutable governance rule registry.
- Provide basic rule providers:
  - `RouteRuleProvider`
  - `CircuitBreakerRuleProvider`
  - `AuthorizationRuleProvider`
  - `RateLimitRuleProvider`
- Preserve last-known-good rules when an update is partially invalid.
- Route decision requests for service-to-service traffic through the compatibility HTTP API.
- Service lifecycle and traffic policy lookup through the compatibility HTTP API.
- API key based authentication header support.
- Standard Go `net/http` transport for compatibility calls.
- Timeout configuration for request execution.
- Context propagation for all client calls.

## Current Status

| Item | Value |
| --- | --- |
| Stability | Early development |
| Language | Go |
| Rule source | `stellnula-service` |
| Transport | `net/http`, StellNula data plane |
| Target service | `stellorbit-service` |
| Maintainer | StellHub |

## Installation

```bash
go get github.com/stellhub/stellorbit-go-sdk
```

## Quick Start

Consume governance rules from StellNula:

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
		StellnulaEndpoint: "http://localhost:8060",
		AppID:             "payment-service",
		ClientID:          "payment-service-local",
		Env:               "dev",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	if err := client.Start(context.Background()); err != nil {
		log.Fatal(err)
	}

	rules, err := client.Routes().Find(context.Background(), stellorbit.RouteRuleQuery{
		ServiceName: "payment-service",
		RouteKey:    "tenant-a",
		Attributes: map[string]string{
			"env": "dev",
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(rules)
}
```

The SDK subscribes to:

| Field | Value |
| --- | --- |
| `namespace` | `governance` |
| `group` | `service-governance` |

Governance configs are aggregate payloads published by `stellorbit-service`. Each StellNula entry uses a fixed `configId` in the form `stellorbit.<applicationCode>.<ruleType>` and must carry `schemaVersion=stellorbit.governance.aggregate.v1`. The SDK parses `rules[]` from that aggregate payload; `rules[].ruleId` is the rule identity, while the aggregate `configId` is used to replace, delete, and fall back an entire rule group. Aggregate root payloads must include the validator field for their type: `routes`, `limit`, `breaker`, or `auth`.

The original StellOrbit HTTP API remains available for compatibility:

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
| `Start(ctx)` | Start the StellNula-backed governance rule source |
| `Rules()` | Return the current immutable governance rule registry |
| `Routes()` | Return `RouteRuleProvider` |
| `CircuitBreakers()` | Return `CircuitBreakerRuleProvider` |
| `Authorizations()` | Return `AuthorizationRuleProvider` |
| `RateLimits()` | Return `RateLimitRuleProvider` |
| `Route(ctx, request)` | Request a route decision from StellOrbit |
| `LifecyclePolicy(ctx, serviceName)` | Fetch service lifecycle governance policy |
| `TrafficPolicy(ctx, serviceName)` | Fetch routing, retry, and traffic shifting policy |

## Package Layout

| Package | Responsibility |
| --- | --- |
| `github.com/stellhub/stellorbit-go-sdk` | Public SDK facade, client lifecycle, options, compatibility aliases |
| `github.com/stellhub/stellorbit-go-sdk/governance` | Rule model, registry, parser, matcher, queries, providers, source contracts |
| `internal/source/stellnula` | StellNula-backed governance rule source |
| `internal/httpapi` | Compatibility HTTP client for legacy StellOrbit API calls |

## Development

Run verification:

```bash
go test ./...
```

## Repository Scope

This SDK intentionally keeps the first version small. Future releases can add:

- Strongly typed policy models.
- HTTP middleware and gRPC interceptor adapters.
- Retry, failover, and load balancing helpers.
- OpenAPI generated DTOs when the service contract is stable.
- Integration tests against local `stellorbit-service`.
- Observability hooks for client-side metrics and tracing.

## License

The license will be defined before the first stable release.
