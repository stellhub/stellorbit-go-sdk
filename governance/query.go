package governance

import (
	"sort"
	"strings"
)

type RequestContext struct {
	ClientID      string
	TraceID       string
	SpanID        string
	TenantID      string
	QuotaKey      string
	AuthContextID string
	TrafficClass  string
	TrafficTag    string
	Attributes    map[string]string
}

func EmptyRequestContext() RequestContext {
	return RequestContext{}
}

func (c RequestContext) AsAttributes() map[string]string {
	attributes := copyStringMap(c.Attributes)
	putIfPresent(attributes, "clientId", c.ClientID)
	putIfPresent(attributes, "traceId", c.TraceID)
	putIfPresent(attributes, "spanId", c.SpanID)
	putIfPresent(attributes, "tenantId", c.TenantID)
	putIfPresent(attributes, "quotaKey", c.QuotaKey)
	putIfPresent(attributes, "authContextId", c.AuthContextID)
	putIfPresent(attributes, "trafficClass", c.TrafficClass)
	putIfPresent(attributes, "trafficTag", c.TrafficTag)
	return attributes
}

type RouteRuleQuery struct {
	ServiceName string
	RouteKey    string
	Attributes  map[string]string
	Context     RequestContext
}

func (q RouteRuleQuery) AttributesForMatch() map[string]string {
	attributes := copyStringMap(q.Attributes)
	mergeAttributes(attributes, commonRuleAttributes(q.ServiceName, q.Context))
	putIfPresent(attributes, "routeKey", q.RouteKey)
	return attributes
}

type CircuitBreakerRuleQuery struct {
	ServiceName string
	Operation   string
	Context     RequestContext
}

func (q CircuitBreakerRuleQuery) AttributesForMatch() map[string]string {
	attributes := commonRuleAttributes(q.ServiceName, q.Context)
	putIfPresent(attributes, "operation", q.Operation)
	return attributes
}

type AuthorizationRuleQuery struct {
	ServiceName string
	Principal   string
	TenantID    string
	Roles       []string
	Token       string
	Context     RequestContext
}

func (q AuthorizationRuleQuery) AttributesForMatch() map[string]string {
	attributes := commonRuleAttributes(q.ServiceName, q.Context)
	putIfPresent(attributes, "principal", q.Principal)
	tenantID := q.TenantID
	if strings.TrimSpace(tenantID) == "" {
		tenantID = q.Context.TenantID
	}
	putIfPresent(attributes, "tenantId", tenantID)
	putIfPresent(attributes, "token", q.Token)
	if len(q.Roles) > 0 {
		roles := append([]string(nil), q.Roles...)
		sort.Strings(roles)
		attributes["roles"] = strings.Join(roles, ",")
	}
	return attributes
}

type RateLimitRuleQuery struct {
	ServiceName string
	QuotaKey    string
	Context     RequestContext
}

func (q RateLimitRuleQuery) AttributesForMatch() map[string]string {
	attributes := commonRuleAttributes(q.ServiceName, q.Context)
	quotaKey := q.QuotaKey
	if strings.TrimSpace(quotaKey) == "" {
		quotaKey = q.Context.QuotaKey
	}
	if strings.TrimSpace(quotaKey) == "" {
		quotaKey = q.Context.TenantID
	}
	putIfPresent(attributes, "quotaKey", quotaKey)
	return attributes
}

func commonRuleAttributes(serviceName string, context RequestContext) map[string]string {
	attributes := context.AsAttributes()
	putIfPresent(attributes, "serviceName", serviceName)
	return attributes
}

func mergeAttributes(target map[string]string, source map[string]string) {
	for key, value := range source {
		target[key] = value
	}
}

func putIfPresent(values map[string]string, key string, value string) {
	if strings.TrimSpace(value) != "" {
		values[key] = value
	}
}

func copyStringMap(values map[string]string) map[string]string {
	if len(values) == 0 {
		return map[string]string{}
	}
	copied := make(map[string]string, len(values))
	for key, value := range values {
		copied[key] = value
	}
	return copied
}
