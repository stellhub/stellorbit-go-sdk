package httpapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	apiKeyHeader      = "X-Stellorbit-Api-Key"
	contentTypeHeader = "Content-Type"
	acceptHeader      = "Accept"
	userAgentHeader   = "User-Agent"
	applicationJSON   = "application/json"
	defaultUserAgent  = "stellorbit-go-sdk"
)

type Client struct {
	endpoint   string
	apiKey     string
	httpClient *http.Client
}

type Options struct {
	Endpoint   string
	APIKey     string
	HTTPClient *http.Client
}

type RouteRequest struct {
	ServiceName string            `json:"serviceName"`
	RouteKey    string            `json:"routeKey"`
	Attributes  map[string]string `json:"attributes"`
}

type APIResponse struct {
	StatusCode int
	Body       string
}

func (r APIResponse) Successful() bool {
	return r.StatusCode >= 200 && r.StatusCode < 300
}

type HTTPError struct {
	StatusCode int
	Body       string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("stellorbit: http status %d: %s", e.StatusCode, e.Body)
}

func NewClient(options Options) *Client {
	return &Client{
		endpoint:   strings.TrimRight(options.Endpoint, "/"),
		apiKey:     options.APIKey,
		httpClient: options.HTTPClient,
	}
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
	request, err := http.NewRequestWithContext(ctx, method, c.endpoint+path, body)
	if err != nil {
		return nil, fmt.Errorf("create stellorbit request: %w", err)
	}
	request.Header.Set(acceptHeader, applicationJSON)
	request.Header.Set(userAgentHeader, defaultUserAgent)
	if c.apiKey != "" {
		request.Header.Set(apiKeyHeader, c.apiKey)
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
