package stellorbit

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoute(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/stellorbit/v1/routes/decide" {
			t.Fatalf("expected route path, got %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.Header.Get(apiKeyHeader) != "test-key" {
			t.Fatal("expected api key header")
		}

		var request RouteRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if request.ServiceName != "payment-service" {
			t.Fatalf("expected payment-service, got %s", request.ServiceName)
		}

		w.Header().Set(contentTypeHeader, applicationJSON)
		_, _ = w.Write([]byte(`{"target":"payment-service-v1","retry":2}`))
	}))
	defer server.Close()

	client, err := NewClient(Options{Endpoint: server.URL, APIKey: "test-key"})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	response, err := client.Route(context.Background(), RouteRequest{
		ServiceName: "payment-service",
		RouteKey:    "tenant-a",
		Attributes:  map[string]string{"env": "dev"},
	})
	if err != nil {
		t.Fatalf("route: %v", err)
	}
	if !response.Successful() {
		t.Fatalf("expected successful response")
	}
	if response.Body == "" {
		t.Fatal("expected response body")
	}
}

func TestPolicyLookups(t *testing.T) {
	paths := map[string]bool{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths[r.URL.Path] = true
		_, _ = w.Write([]byte(`{"policy":"ok"}`))
	}))
	defer server.Close()

	client, err := NewClient(Options{Endpoint: server.URL})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if _, err := client.LifecyclePolicy(context.Background(), "checkout-service"); err != nil {
		t.Fatalf("lifecycle policy: %v", err)
	}
	if _, err := client.TrafficPolicy(context.Background(), "checkout-service"); err != nil {
		t.Fatalf("traffic policy: %v", err)
	}

	if !paths["/api/stellorbit/v1/services/checkout-service/lifecycle-policy"] {
		t.Fatal("expected lifecycle policy path")
	}
	if !paths["/api/stellorbit/v1/services/checkout-service/traffic-policy"] {
		t.Fatal("expected traffic policy path")
	}
}

func TestNewClientRequiresEndpoint(t *testing.T) {
	_, err := NewClient(Options{})
	if err != ErrEndpointRequired {
		t.Fatalf("expected ErrEndpointRequired, got %v", err)
	}
}

func TestHTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "denied", http.StatusForbidden)
	}))
	defer server.Close()

	client, err := NewClient(Options{Endpoint: server.URL})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}

	response, err := client.TrafficPolicy(context.Background(), "checkout-service")
	if err == nil {
		t.Fatal("expected error")
	}
	if response == nil || response.StatusCode != http.StatusForbidden {
		t.Fatalf("expected forbidden response, got %#v", response)
	}
}
