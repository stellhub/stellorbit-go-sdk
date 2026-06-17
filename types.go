package stellorbit

import "fmt"

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
