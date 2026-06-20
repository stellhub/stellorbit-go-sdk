package stellorbit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	stellnula "github.com/stellhub/stellnula-go-sdk"
	"github.com/stellhub/stellorbit-go-sdk/governance"
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
	if !errors.Is(err, ErrEndpointRequired) {
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

func TestProvidersReturnMatchedGovernanceRules(t *testing.T) {
	rules := []GovernanceRule{
		mustParseRule(t, "route-payment", `{
			"ruleType": "ROUTE",
			"targetService": "payment-service",
			"status": "ACTIVE",
			"priority": 10,
			"conditions": {"env": "prod"},
			"routes": [{"id": "payment-v2", "targetService": "payment-service-v2", "weight": 100}]
		}`),
		mustParseRule(t, "auth-payment", `{
			"ruleType": "AUTH",
			"targetService": "payment-service",
			"status": "ACTIVE",
			"priority": 0,
			"conditions": {"roles": ["payment-admin"]},
			"auth": {"requiredRoles": ["payment-admin"]}
		}`),
		mustParseRule(t, "rate-payment", `{
			"ruleType": "RATE_LIMIT",
			"targetService": "payment-service",
			"status": "ACTIVE",
			"priority": 0,
			"limitMode": "QPS",
			"limitType": "TENANT",
			"limitAlgorithm": "TOKEN_BUCKET",
			"trafficProtocol": "HTTP",
			"executionLocation": "APPLICATION",
			"coordinationMode": "LOCAL_ONLY",
			"limit": {"quota": 100, "windowSeconds": 60, "keyAttribute": "tenantId"}
		}`),
		mustParseRule(t, "breaker-payment", `{
			"ruleType": "CIRCUIT_BREAKER",
			"targetService": "payment-service",
			"status": "ACTIVE",
			"priority": 0,
			"breaker": {"failureRateThreshold": 50, "slidingWindowSize": 100}
		}`),
	}
	client, err := NewClient(
		Options{},
		WithRuleSource(NewInMemoryGovernanceRuleSource(NewGovernanceRuleRegistry(1, "test", rules))),
	)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if err := client.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}

	routeRules, err := client.Routes().Find(context.Background(), RouteRuleQuery{
		ServiceName: "payment-service",
		RouteKey:    "tenant-a",
		Attributes:  map[string]string{"env": "prod"},
	})
	if err != nil {
		t.Fatalf("find route rules: %v", err)
	}
	if len(routeRules) != 1 || routeRules[0].RuleID != "route-payment" {
		t.Fatalf("unexpected route rules: %#v", routeRules)
	}

	spoofedRules, err := client.Routes().Find(context.Background(), RouteRuleQuery{
		ServiceName: "payment-service",
		Attributes:  map[string]string{"serviceName": "spoofed-service"},
	})
	if err != nil {
		t.Fatalf("find spoofed rules: %v", err)
	}
	if len(spoofedRules) != 0 {
		t.Fatalf("route attributes must not override service name: %#v", spoofedRules)
	}

	authRules, err := client.Authorizations().Find(context.Background(), AuthorizationRuleQuery{
		ServiceName: "payment-service",
		Principal:   "alice",
		TenantID:    "tenant-a",
		Roles:       []string{"viewer", "payment-admin"},
	})
	if err != nil {
		t.Fatalf("find auth rules: %v", err)
	}
	if len(authRules) != 1 || authRules[0].RuleID != "auth-payment" {
		t.Fatalf("unexpected auth rules: %#v", authRules)
	}

	rateRules, err := client.RateLimits().Find(context.Background(), RateLimitRuleQuery{
		ServiceName: "payment-service",
		Context:     RequestContext{TenantID: "tenant-a"},
	})
	if err != nil {
		t.Fatalf("find rate limit rules: %v", err)
	}
	if len(rateRules) != 1 || rateRules[0].RuleID != "rate-payment" {
		t.Fatalf("unexpected rate rules: %#v", rateRules)
	}

	breakerRule, ok, err := client.CircuitBreakers().First(context.Background(), CircuitBreakerRuleQuery{
		ServiceName: "payment-service",
		Operation:   "POST /pay",
	})
	if err != nil {
		t.Fatalf("first breaker rule: %v", err)
	}
	if !ok || breakerRule.RuleID != "breaker-payment" {
		t.Fatalf("unexpected breaker rule: %#v %v", breakerRule, ok)
	}
}

func TestRateLimitProviderFiltersEnterpriseLimiterModes(t *testing.T) {
	rules := []GovernanceRule{
		mustParseRule(t, "rate-local-qps", `{
			"ruleType": "RATE_LIMIT",
			"targetService": "payment-service",
			"status": "ACTIVE",
			"priority": 0,
			"limitMode": "QPS",
			"limitType": "REQUEST",
			"limitAlgorithm": "TOKEN_BUCKET",
			"trafficProtocol": "HTTP",
			"executionLocation": "APPLICATION",
			"coordinationMode": "LOCAL_ONLY",
			"keyExtractor": {
				"keys": [{"name": "tenant", "source": "TENANT", "key": "tenantId", "required": true}]
			},
			"limit": {"quota": 100, "windowSeconds": 60}
		}`),
		mustParseRule(t, "rate-grpc-header-global", `{
			"ruleType": "RATE_LIMIT",
			"targetService": "payment-service",
			"status": "ACTIVE",
			"priority": 1,
			"limitMode": "HEADER",
			"limitType": "HEADER",
			"limitAlgorithm": "QUOTA_LEASE",
			"trafficProtocol": "GRPC",
			"executionLocation": "APPLICATION",
			"coordinationMode": "GLOBAL_QUOTA",
			"requestMatcher": {"grpcService": "payment.PaymentService"},
			"keyExtractor": {
				"keys": [{"name": "region", "source": "GRPC_METADATA", "key": "x-region", "required": true, "normalize": "LOWERCASE"}]
			},
			"quotaConfig": {"quota": 500},
			"observabilityConfig": {"metrics": true},
			"shadowConfig": {"enabled": false},
			"limit": {"quota": 500, "windowSeconds": 60}
		}`),
		mustParseRule(t, "rate-global-qps", `{
			"ruleType": "RATE_LIMIT",
			"targetService": "payment-service",
			"status": "ACTIVE",
			"priority": 2,
			"limitMode": "QPS",
			"limitType": "REQUEST",
			"limitAlgorithm": "QUOTA_LEASE",
			"trafficProtocol": "HTTP",
			"executionLocation": "APPLICATION",
			"coordinationMode": "GLOBAL_QUOTA",
			"limit": {"quota": 200, "windowSeconds": 60}
		}`),
	}
	client, err := NewClient(
		Options{},
		WithRuleSource(NewInMemoryGovernanceRuleSource(NewGovernanceRuleRegistry(1, "test", rules))),
	)
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	if err := client.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}

	headerRules, err := client.RateLimits().Find(context.Background(), NewRateLimitRuleQuery(
		"payment-service",
		ByLimitMode(LimitModeHeader),
		ByProtocol(TrafficProtocolGRPC),
		ByKeyExtractorSource(KeyExtractorSourceGRPCMetadata),
		DistributedOnly(),
	))
	if err != nil {
		t.Fatalf("find header rules: %v", err)
	}
	if len(headerRules) != 1 || headerRules[0].RuleID != "rate-grpc-header-global" {
		t.Fatalf("unexpected header rules: %#v", headerRules)
	}

	localRules, err := client.RateLimits().Find(context.Background(), NewRateLimitRuleQuery("payment-service", LocalOnly()))
	if err != nil {
		t.Fatalf("find local rules: %v", err)
	}
	if len(localRules) != 1 || localRules[0].RuleID != "rate-local-qps" {
		t.Fatalf("unexpected local rules: %#v", localRules)
	}

	globalRules, err := client.RateLimits().Find(context.Background(), NewRateLimitRuleQuery(
		"payment-service",
		ByLimitMode(LimitModeQPS),
		DistributedOnly(),
	))
	if err != nil {
		t.Fatalf("find global rules: %v", err)
	}
	if len(globalRules) != 1 || globalRules[0].RuleID != "rate-global-qps" {
		t.Fatalf("unexpected global distributed rules: %#v", globalRules)
	}
	globalView, ok := globalRules[0].RateLimit()
	if !ok {
		t.Fatal("expected global rule to expose rate limit view")
	}
	if globalView.LimitMode != LimitModeQPS || !globalView.IsDistributedRule() {
		t.Fatalf("unexpected global rate limit view: %#v", globalView)
	}
}

func TestRateLimitRuleUnmarshalEnterpriseFields(t *testing.T) {
	var rule RateLimitRule
	if err := json.Unmarshal([]byte(`{
		"limitMode": "MODEL",
		"limitType": "MODEL_TOKEN",
		"limitAlgorithm": "ADAPTIVE",
		"trafficProtocol": "MODEL",
		"executionLocation": "GATEWAY",
		"coordinationMode": "GLOBAL_SYNC",
		"targetSelector": {"service": "chat-service"},
		"requestMatcher": {"path": "/v1/chat/completions"},
		"keyExtractor": {
			"keys": [
				{"name": "api-key", "source": "API_KEY", "key": "Authorization", "required": true, "normalize": "HASH"},
				{"name": "custom", "source": "PLUGIN_CONTEXT", "key": "pluginKey"}
			]
		},
		"dimensions": [{"name": "tenant"}],
		"quotaConfig": {"quota": 1000},
		"windowConfig": {"seconds": 60},
		"burstConfig": {"capacity": 200},
		"concurrencyConfig": {"max": 10},
		"hotspotConfig": {"topN": 20},
		"customPolicy": {"type": "EXPRESSION"},
		"modelLimitConfig": {"unit": "token"},
		"fallbackPolicy": {"strategy": "FAIL_OPEN"},
		"responsePolicy": {"status": 429},
		"observabilityConfig": {"metrics": true},
		"shadowConfig": {"enabled": true}
	}`), &rule); err != nil {
		t.Fatalf("unmarshal enterprise rate limit rule: %v", err)
	}
	if rule.LimitMode != LimitModeModel || rule.LimitType != LimitTypeModelToken || rule.LimitAlgorithm != LimitAlgorithmAdaptive {
		t.Fatalf("unexpected model limit fields: %#v", rule)
	}
	if rule.TrafficProtocol != TrafficProtocolModel || rule.ExecutionLocation != ExecutionLocationGateway {
		t.Fatalf("unexpected protocol or execution location: %#v", rule)
	}
	if !rule.IsDistributedRule() {
		t.Fatalf("expected global sync rule to be distributed: %#v", rule)
	}
	if len(rule.RequestMatcher) == 0 || len(rule.QuotaConfig) == 0 || len(rule.ModelLimitConfig) == 0 {
		t.Fatalf("expected structured config maps: %#v", rule)
	}
	if len(rule.KeyExtractor.Keys) != 2 {
		t.Fatalf("expected key extractor keys, got %#v", rule.KeyExtractor.Keys)
	}
	unsupported := rule.UnsupportedKeyExtractorSources()
	if len(unsupported) != 1 || unsupported[0].Source != KeyExtractorSource("PLUGIN_CONTEXT") {
		t.Fatalf("expected unsupported key source to be marked, got %#v", unsupported)
	}
}

func TestParserRejectsUnknownRateLimitEnum(t *testing.T) {
	configID, payload := mustAggregatePayload(t, "rate-unknown", `{
		"ruleType": "RATE_LIMIT",
		"targetService": "payment-service",
		"status": "ACTIVE",
		"priority": 0,
		"limitMode": "QPSS",
		"limitType": "REQUEST",
		"limitAlgorithm": "TOKEN_BUCKET",
		"trafficProtocol": "HTTP",
		"executionLocation": "APPLICATION",
		"coordinationMode": "LOCAL_ONLY",
		"limit": {"quota": 100, "windowSeconds": 60}
	}`)
	_, err := NewGovernanceRuleParser().Parse(governance.Entry{
		ConfigID:  configID,
		ConfigKey: configID,
		Value:     payload,
		Revision:  1,
	}, "test")
	if err == nil {
		t.Fatal("expected unknown rate limit enum to be rejected")
	}
}

func TestStellnulaGovernanceRuleSourceSubscribesGovernanceChannel(t *testing.T) {
	var captured struct {
		AppID     string `json:"appId"`
		ClientID  string `json:"clientId"`
		Namespace string `json:"namespace"`
		Group     string `json:"group"`
	}
	configID, payload := mustAggregatePayload(t, "route-payment", `{
		"ruleType": "ROUTE",
		"targetService": "payment-service",
		"status": "ACTIVE",
		"priority": 0,
		"routes": [{"id": "payment-v2", "targetService": "payment-service-v2", "weight": 100}]
	}`)
	entries := []stellnula.ConfigEntry{{
		ConfigID:    configID,
		ConfigKey:   configID,
		ContentType: "FILE",
		Value:       payload,
		Version:     1,
		Revision:    7,
	}}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/client/bootstrap":
			if err := json.NewDecoder(r.Body).Decode(&captured); err != nil {
				t.Fatalf("decode bootstrap request: %v", err)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"revision": int64(7),
				"configs":  entries,
			})
		case "/api/v1/client/heartbeat":
			_, _ = w.Write([]byte(`{"accepted":true,"serverRevision":7}`))
		default:
			t.Fatalf("unexpected stellnula path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client, err := NewClient(Options{
		StellnulaEndpoint: server.URL,
		AppID:             "demo",
		ClientID:          "client-1",
		WatchEnabled:      Bool(false),
		SnapshotDirectory: t.TempDir(),
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	t.Cleanup(func() {
		_ = client.Close()
	})
	if err := client.Start(context.Background()); err != nil {
		t.Fatalf("start: %v", err)
	}
	if captured.AppID != "demo" || captured.ClientID != "client-1" {
		t.Fatalf("unexpected identity: %#v", captured)
	}
	if captured.Namespace != defaultRuleNamespace || captured.Group != defaultRuleGroup {
		t.Fatalf("unexpected governance channel: %#v", captured)
	}
	rules, err := client.Routes().Find(context.Background(), RouteRuleQuery{ServiceName: "payment-service"})
	if err != nil {
		t.Fatalf("find route rules: %v", err)
	}
	if len(rules) != 1 || rules[0].RuleID != "route-payment" {
		t.Fatalf("unexpected route rules from stellnula snapshot: %#v", rules)
	}
}

func TestSnapshotParserKeepsFallbackAndRemovesDeletedRule(t *testing.T) {
	parser := NewGovernanceRuleSnapshotParser(NewGovernanceRuleParser(), nil)
	previousRule := mustParseRule(t, "rate-payment", `{
		"ruleType": "RATE_LIMIT",
		"targetService": "payment-service",
		"status": "ACTIVE",
		"priority": 0,
		"limitMode": "QPS",
		"limitType": "REQUEST",
		"limitAlgorithm": "TOKEN_BUCKET",
		"trafficProtocol": "HTTP",
		"executionLocation": "APPLICATION",
		"coordinationMode": "LOCAL_ONLY",
		"limit": {"quota": 100}
	}`)
	aggregateConfigID := previousRule.ConfigKey
	previous := NewGovernanceRuleRegistry(1, "previous", []GovernanceRule{previousRule})

	invalid := parser.Parse(governance.Snapshot{
		Revision: 2,
		Checksum: "next",
		Entries: []governance.Entry{{
			ConfigID:  aggregateConfigID,
			ConfigKey: aggregateConfigID,
			Value:     "{",
			Revision:  2,
		}},
	}, previous)
	if _, ok := invalid.FindByID("rate-payment"); !ok {
		t.Fatalf("expected previous rule to be retained after invalid update")
	}

	deleted := parser.Parse(governance.Snapshot{
		Revision: 3,
		Checksum: "deleted",
		Entries: []governance.Entry{{
			ConfigID:  aggregateConfigID,
			ConfigKey: aggregateConfigID,
			Deleted:   true,
			Revision:  3,
		}},
	}, previous)
	if len(deleted.Rules) != 0 {
		t.Fatalf("expected deleted rule to be removed, got %#v", deleted.Rules)
	}
}

func TestSnapshotParserKeepsClearedAggregateWhenOtherEntryIsInvalid(t *testing.T) {
	parser := NewGovernanceRuleSnapshotParser(NewGovernanceRuleParser(), nil)
	previousRule := mustParseRule(t, "route-payment", `{
		"ruleType": "ROUTE",
		"targetService": "payment-service",
		"status": "ACTIVE",
		"priority": 0,
		"routes": [{"id": "payment-v2"}]
	}`)
	previous := NewGovernanceRuleRegistry(1, "previous", []GovernanceRule{previousRule})
	emptyConfigID, emptyPayload := mustEmptyAggregatePayload(t, "payment-service", "ROUTE")
	invalidConfigID := aggregateConfigID("payment-service", "AUTH")

	next := parser.Parse(governance.Snapshot{
		Revision: 2,
		Checksum: "next",
		Entries: []governance.Entry{
			{
				ConfigID:  emptyConfigID,
				ConfigKey: emptyConfigID,
				Value:     emptyPayload,
				Revision:  2,
			},
			{
				ConfigID:  invalidConfigID,
				ConfigKey: invalidConfigID,
				Value:     "{",
				Revision:  2,
			},
		},
	}, previous)
	if len(next.Rules) != 0 {
		t.Fatalf("expected cleared aggregate to stay empty, got %#v", next.Rules)
	}
}

func TestParserUsesAggregatedRuleIDContract(t *testing.T) {
	configID, payload := mustAggregatePayload(t, "json-route-id", `{
		"ruleType": "ROUTE",
		"targetService": "payment-service",
		"status": "ACTIVE",
		"priority": 0,
		"routes": [{"id": "payment-v2"}]
	}`)
	rule, err := NewGovernanceRuleParser().Parse(governance.Entry{
		ConfigID:  configID,
		ConfigKey: configID,
		Value:     payload,
		Revision:  1,
	}, "test")
	if err != nil {
		t.Fatalf("parse rule: %v", err)
	}
	if rule.RuleID != "json-route-id" {
		t.Fatalf("expected rules[].ruleId to be authoritative rule id, got %s", rule.RuleID)
	}
	if rule.ConfigKey != configID {
		t.Fatalf("expected aggregate config id as config key, got %s", rule.ConfigKey)
	}
	if rule.Content["aggregateConfigId"] != configID {
		t.Fatalf("expected aggregate config id in content, got %#v", rule.Content)
	}
}

func TestParserRejectsUnknownRuleStatus(t *testing.T) {
	configID, payload := mustAggregatePayload(t, "route-payment", `{
		"ruleType": "ROUTE",
		"targetService": "payment-service",
		"status": "ACTVE",
		"priority": 0,
		"routes": [{"id": "payment-v2"}]
	}`)
	_, err := NewGovernanceRuleParser().Parse(governance.Entry{
		ConfigID:  configID,
		ConfigKey: configID,
		Value:     payload,
		Revision:  1,
	}, "test")
	if err == nil {
		t.Fatal("expected unknown rule status to be rejected")
	}
}

func TestParserRejectsAggregateWithoutValidatorPayload(t *testing.T) {
	configID, payload := mustAggregatePayload(t, "route-payment", `{
		"ruleType": "ROUTE",
		"targetService": "payment-service",
		"status": "ACTIVE",
		"priority": 0,
		"routes": [{"id": "payment-v2"}]
	}`)
	var aggregate map[string]any
	if err := json.Unmarshal([]byte(payload), &aggregate); err != nil {
		t.Fatalf("decode aggregate payload: %v", err)
	}
	delete(aggregate, "routes")
	raw, err := json.Marshal(aggregate)
	if err != nil {
		t.Fatalf("encode aggregate payload: %v", err)
	}
	_, err = NewGovernanceRuleParser().Parse(governance.Entry{
		ConfigID:  configID,
		ConfigKey: configID,
		Value:     string(raw),
		Revision:  1,
	}, "test")
	if err == nil {
		t.Fatal("expected aggregate payload without validator routes to be rejected")
	}
}

func TestParserRejectsTrailingJSON(t *testing.T) {
	configID, payload := mustAggregatePayload(t, "route-payment", `{
		"ruleType": "ROUTE",
		"targetService": "payment-service",
		"status": "ACTIVE",
		"priority": 0,
		"routes": [{"id": "payment-v2"}]
	}`)
	_, err := NewGovernanceRuleParser().Parse(governance.Entry{
		ConfigID:  configID,
		ConfigKey: configID,
		Value:     payload + ` {"extra": true}`,
		Revision:  1,
	}, "test")
	if err == nil {
		t.Fatal("expected trailing JSON to be rejected")
	}
}

func TestNewStellnulaGovernanceRuleSourceWithClientRejectsNilClient(t *testing.T) {
	defer func() {
		if recovered := recover(); recovered == nil {
			t.Fatal("expected nil stellnula client to panic")
		}
	}()
	_ = NewStellnulaGovernanceRuleSourceWithClient(nil, false, nil)
}

func TestGovernanceRegistryCloneProtectsNestedContent(t *testing.T) {
	rule := mustParseRule(t, "route-payment", `{
		"ruleType": "ROUTE",
		"targetService": "payment-service",
		"status": "ACTIVE",
		"priority": 0,
		"routes": [{"id": "payment-v2", "weight": 100}]
	}`)
	registry := NewGovernanceRuleRegistry(1, "test", []GovernanceRule{rule})
	cloned := registry.Clone()

	route := cloned.Rules[0].Content["routes"].([]any)[0].(map[string]any)
	route["weight"] = json.Number("0")

	originalRoute := registry.Rules[0].Content["routes"].([]any)[0].(map[string]any)
	if originalRoute["weight"] == json.Number("0") {
		t.Fatalf("expected nested content mutation to stay outside original registry")
	}
}

func mustParseRule(t *testing.T, ruleID string, content string) GovernanceRule {
	t.Helper()
	configID, payload := mustAggregatePayload(t, ruleID, content)
	rule, err := NewGovernanceRuleParser().Parse(governance.Entry{
		ConfigID:    configID,
		ConfigKey:   configID,
		ContentType: "FILE",
		Value:       payload,
		Revision:    1,
	}, "test")
	if err != nil {
		t.Fatalf("parse rule %s: %v", ruleID, err)
	}
	return rule
}

func mustAggregatePayload(t *testing.T, ruleID string, content string) (string, string) {
	t.Helper()
	var model map[string]any
	decoder := json.NewDecoder(strings.NewReader(content))
	decoder.UseNumber()
	if err := decoder.Decode(&model); err != nil {
		t.Fatalf("decode rule content %s: %v", ruleID, err)
	}
	ruleType := stringField(model, "ruleType")
	targetService := stringField(model, "targetService")
	if ruleType == "" || targetService == "" {
		t.Fatalf("rule content %s must include ruleType and targetService", ruleID)
	}
	stellnulaRuleType := stellnulaRuleType(ruleType)
	validatorField := validatorField(stellnulaRuleType)
	validatorPayload, ok := model[validatorField]
	if !ok {
		t.Fatalf("rule content %s must include validator payload %s", ruleID, validatorField)
	}
	configID := aggregateConfigID(targetService, stellnulaRuleType)
	sourceRuleType := sourceRuleType(ruleType)
	payload := map[string]any{
		"schemaVersion":   "stellorbit.governance.aggregate.v1",
		"releaseVersion":  json.Number("1"),
		"generatedAt":     "2026-06-17T10:15:30+08:00",
		"applicationCode": targetService,
		"configId":        configID,
		"ruleType":        stellnulaRuleType,
		"sourceRuleType":  sourceRuleType,
		"targetService":   targetService,
		"status":          stringField(model, "status"),
		"priority":        model["priority"],
		"releaseName":     "test-release",
		"runtimeFormat":   "JSON",
		"ruleCount":       json.Number("1"),
		"rules": []any{
			map[string]any{
				"ruleId":            ruleID,
				"configId":          configID,
				"ruleType":          sourceRuleType,
				"stellnulaRuleType": stellnulaRuleType,
				"ruleCode":          ruleID,
				"ruleName":          "Rule " + ruleID,
				"targetService":     targetService,
				"status":            stringField(model, "status"),
				"priority":          model["priority"],
				"draftVersion":      json.Number("1"),
				"schemaVersion":     "stellorbit.governance.v1",
				"checksum":          "rule-checksum",
				"content":           model,
			},
		},
		validatorField: []any{validatorPayload},
		"checksum":     "aggregate-checksum",
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("encode aggregate payload %s: %v", ruleID, err)
	}
	return configID, string(raw)
}

func mustEmptyAggregatePayload(t *testing.T, targetService string, ruleType string) (string, string) {
	t.Helper()
	stellnulaRuleType := stellnulaRuleType(ruleType)
	validatorField := validatorField(stellnulaRuleType)
	configID := aggregateConfigID(targetService, stellnulaRuleType)
	payload := map[string]any{
		"schemaVersion":   "stellorbit.governance.aggregate.v1",
		"releaseVersion":  json.Number("1"),
		"generatedAt":     "2026-06-17T10:15:30+08:00",
		"applicationCode": targetService,
		"configId":        configID,
		"ruleType":        stellnulaRuleType,
		"sourceRuleType":  sourceRuleType(ruleType),
		"targetService":   targetService,
		"status":          "DISABLED",
		"priority":        json.Number("1000"),
		"releaseName":     "test-release",
		"runtimeFormat":   "JSON",
		"ruleCount":       json.Number("0"),
		"rules":           []any{},
		validatorField:    []any{},
		"checksum":        "aggregate-checksum",
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("encode empty aggregate payload: %v", err)
	}
	return configID, string(raw)
}

func stringField(values map[string]any, key string) string {
	value, ok := values[key]
	if !ok || value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func sourceRuleType(ruleType string) string {
	if strings.EqualFold(ruleType, "CIRCUIT_BREAKER") {
		return "BREAKER"
	}
	return strings.ToUpper(strings.TrimSpace(ruleType))
}

func stellnulaRuleType(ruleType string) string {
	if strings.EqualFold(ruleType, "BREAKER") {
		return "CIRCUIT_BREAKER"
	}
	return strings.ToUpper(strings.TrimSpace(ruleType))
}

func validatorField(ruleType string) string {
	switch strings.ToUpper(strings.TrimSpace(ruleType)) {
	case "ROUTE":
		return "routes"
	case "RATE_LIMIT":
		return "limit"
	case "CIRCUIT_BREAKER":
		return "breaker"
	case "AUTH":
		return "auth"
	default:
		return strings.ToLower(ruleType)
	}
}

func aggregateConfigID(applicationCode, ruleType string) string {
	return "stellorbit." + normalizeConfigSegment(applicationCode) + "." + normalizeConfigSegment(ruleType)
}

func normalizeConfigSegment(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	for _, r := range value {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' || r == '_' || r == '.' || r == '-' {
			builder.WriteRune(r)
			continue
		}
		builder.WriteByte('-')
	}
	return builder.String()
}
