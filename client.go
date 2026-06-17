package stellorbit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultTimeout    = 10 * time.Second
	apiKeyHeader      = "X-Stellorbit-Api-Key"
	contentTypeHeader = "Content-Type"
	acceptHeader      = "Accept"
	userAgentHeader   = "User-Agent"
	applicationJSON   = "application/json"
	defaultUserAgent  = "stellorbit-go-sdk"
)

var ErrEndpointRequired = errors.New("stellorbit: endpoint is required")

type Client struct {
	options    Options
	httpClient *http.Client
}

type Options struct {
	Endpoint   string
	APIKey     string
	Timeout    time.Duration
	HTTPClient *http.Client
}

func NewClient(options Options) (*Client, error) {
	normalized, err := options.normalize()
	if err != nil {
		return nil, err
	}
	return &Client{
		options:    normalized,
		httpClient: normalized.HTTPClient,
	}, nil
}

func (c *Client) Route(ctx context.Context, request RouteRequest) (*APIResponse, error) {
	payload, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("marshal route request: %w", err)
	}
	httpRequest, err := c.newRequest(ctx, http.MethodPost, "/api/stellorbit/v1/routes/decide", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	httpRequest.Header.Set(contentTypeHeader, applicationJSON)
	return c.send(httpRequest)
}

func (c *Client) LifecyclePolicy(ctx context.Context, serviceName string) (*APIResponse, error) {
	path := "/api/stellorbit/v1/services/" + url.PathEscape(serviceName) + "/lifecycle-policy"
	request, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return c.send(request)
}

func (c *Client) TrafficPolicy(ctx context.Context, serviceName string) (*APIResponse, error) {
	path := "/api/stellorbit/v1/services/" + url.PathEscape(serviceName) + "/traffic-policy"
	request, err := c.newRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return c.send(request)
}

func (c *Client) newRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Request, error) {
	request, err := http.NewRequestWithContext(ctx, method, c.resolve(path), body)
	if err != nil {
		return nil, fmt.Errorf("create stellorbit request: %w", err)
	}
	request.Header.Set(acceptHeader, applicationJSON)
	request.Header.Set(userAgentHeader, defaultUserAgent)
	if c.options.APIKey != "" {
		request.Header.Set(apiKeyHeader, c.options.APIKey)
	}
	return request, nil
}

func (c *Client) send(request *http.Request) (*APIResponse, error) {
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("call stellorbit service: %w", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("read stellorbit response: %w", err)
	}

	apiResponse := &APIResponse{
		StatusCode: response.StatusCode,
		Body:       string(body),
	}
	if !apiResponse.Successful() {
		return apiResponse, &HTTPError{StatusCode: response.StatusCode, Body: string(body)}
	}
	return apiResponse, nil
}

func (c *Client) resolve(path string) string {
	base := strings.TrimRight(c.options.Endpoint, "/")
	return base + path
}

func (o Options) normalize() (Options, error) {
	if o.Endpoint == "" {
		return Options{}, ErrEndpointRequired
	}
	if _, err := url.ParseRequestURI(o.Endpoint); err != nil {
		return Options{}, fmt.Errorf("stellorbit: invalid endpoint: %w", err)
	}
	if o.Timeout == 0 {
		o.Timeout = defaultTimeout
	}
	if o.HTTPClient == nil {
		o.HTTPClient = &http.Client{Timeout: o.Timeout}
	}
	return o, nil
}
