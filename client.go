package stellorbit

import (
	"context"
	"fmt"
	"sync/atomic"

	stellnula "github.com/stellhub/stellnula-go-sdk"
	"github.com/stellhub/stellorbit-go-sdk/governance"
	"github.com/stellhub/stellorbit-go-sdk/internal/httpapi"
	stellnulasource "github.com/stellhub/stellorbit-go-sdk/internal/source/stellnula"
)

type Client struct {
	options                Options
	httpClient             *httpapi.Client
	ruleSource             governance.Source
	routeProvider          governance.RouteRuleProvider
	circuitBreakerProvider governance.CircuitBreakerRuleProvider
	authorizationProvider  governance.AuthorizationRuleProvider
	rateLimitProvider      governance.RateLimitRuleProvider
	started                atomic.Bool
}

func NewClient(options Options, opts ...Option) (*Client, error) {
	settings := clientSettings{}
	for _, opt := range opts {
		if opt != nil {
			opt(&settings)
		}
	}
	normalized, err := options.normalize(settings.ruleSource != nil)
	if err != nil {
		return nil, err
	}
	source := settings.ruleSource
	if source == nil {
		if normalized.StellnulaEndpoint != "" {
			source, err = newStellnulaSource(normalized, settings.stellnulaClientOptions...)
			if err != nil {
				return nil, err
			}
		} else {
			source = governance.NewInMemorySource(governance.EmptyRegistry())
		}
	}
	client := &Client{
		options:    normalized,
		ruleSource: source,
	}
	if normalized.Endpoint != "" {
		client.httpClient = httpapi.NewClient(httpapi.Options{
			Endpoint:   normalized.Endpoint,
			APIKey:     normalized.APIKey,
			HTTPClient: normalized.HTTPClient,
		})
	}
	client.routeProvider = governance.NewRouteRuleProvider(client.Rules)
	client.circuitBreakerProvider = governance.NewCircuitBreakerRuleProvider(client.Rules)
	client.authorizationProvider = governance.NewAuthorizationRuleProvider(client.Rules)
	client.rateLimitProvider = governance.NewRateLimitRuleProvider(client.Rules)
	return client, nil
}

func (c *Client) Start(ctx context.Context) error {
	if !c.started.CompareAndSwap(false, true) {
		return nil
	}
	if err := c.ruleSource.Start(ctx); err != nil {
		c.started.Store(false)
		return err
	}
	return nil
}

func (c *Client) Close() error {
	return c.ruleSource.Close()
}

func (c *Client) Rules() governance.Registry {
	return c.ruleSource.Registry()
}

func (c *Client) Routes() governance.RouteRuleProvider {
	return c.routeProvider
}

func (c *Client) CircuitBreakers() governance.CircuitBreakerRuleProvider {
	return c.circuitBreakerProvider
}

func (c *Client) Authorizations() governance.AuthorizationRuleProvider {
	return c.authorizationProvider
}

func (c *Client) RateLimits() governance.RateLimitRuleProvider {
	return c.rateLimitProvider
}

func (c *Client) Route(ctx context.Context, request RouteRequest) (*APIResponse, error) {
	httpClient, err := c.compatHTTPClient()
	if err != nil {
		return nil, err
	}
	return httpClient.Route(ctx, request)
}

func (c *Client) LifecyclePolicy(ctx context.Context, serviceName string) (*APIResponse, error) {
	httpClient, err := c.compatHTTPClient()
	if err != nil {
		return nil, err
	}
	return httpClient.LifecyclePolicy(ctx, serviceName)
}

func (c *Client) TrafficPolicy(ctx context.Context, serviceName string) (*APIResponse, error) {
	httpClient, err := c.compatHTTPClient()
	if err != nil {
		return nil, err
	}
	return httpClient.TrafficPolicy(ctx, serviceName)
}

func (c *Client) compatHTTPClient() (*httpapi.Client, error) {
	if c.httpClient == nil {
		return nil, ErrEndpointRequired
	}
	return c.httpClient, nil
}

func newStellnulaSource(options Options, clientOptions ...stellnula.ClientOption) (governance.Source, error) {
	if options.StellnulaEndpoint == "" {
		return nil, fmt.Errorf("stellorbit: stellnula endpoint is required for governance rule source")
	}
	return stellnulasource.New(stellnulasource.Options{
		Endpoint:                 options.StellnulaEndpoint,
		GRPCEndpoint:             options.StellnulaGRPCEndpoint,
		GRPCPlaintext:            options.StellnulaGRPCPlaintext,
		APIToken:                 options.StellnulaAPIToken,
		AppID:                    options.AppID,
		ClientID:                 options.ClientID,
		Env:                      options.Env,
		Region:                   options.Region,
		Zone:                     options.Zone,
		Cluster:                  options.Cluster,
		Namespace:                options.RuleNamespace,
		Group:                    options.RuleGroup,
		WatchEnabled:             options.WatchEnabled,
		FailFastOnBootstrap:      options.FailFastOnBootstrap,
		SnapshotDirectory:        options.SnapshotDirectory,
		Labels:                   options.Labels,
		AcceptLargeFileReference: options.AcceptLargeFileReference,
		Logger:                   options.Logger,
		HTTPClient:               options.HTTPClient,
	}, clientOptions...)
}
