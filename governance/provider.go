package governance

import (
	"context"
	"errors"
	"strings"
)

var ErrServiceNameRequired = errors.New("stellorbit/governance: service name is required")

type RouteRuleProvider interface {
	Find(ctx context.Context, query RouteRuleQuery) ([]Rule, error)
	First(ctx context.Context, query RouteRuleQuery) (Rule, bool, error)
}

type CircuitBreakerRuleProvider interface {
	Find(ctx context.Context, query CircuitBreakerRuleQuery) ([]Rule, error)
	First(ctx context.Context, query CircuitBreakerRuleQuery) (Rule, bool, error)
}

type AuthorizationRuleProvider interface {
	Find(ctx context.Context, query AuthorizationRuleQuery) ([]Rule, error)
	First(ctx context.Context, query AuthorizationRuleQuery) (Rule, bool, error)
}

type RateLimitRuleProvider interface {
	Find(ctx context.Context, query RateLimitRuleQuery) ([]Rule, error)
	First(ctx context.Context, query RateLimitRuleQuery) (Rule, bool, error)
}

type RegistryFunc func() Registry

func NewRouteRuleProvider(registry RegistryFunc) RouteRuleProvider {
	return &defaultRouteRuleProvider{support: newRuleProviderSupport(registry)}
}

func NewCircuitBreakerRuleProvider(registry RegistryFunc) CircuitBreakerRuleProvider {
	return &defaultCircuitBreakerRuleProvider{support: newRuleProviderSupport(registry)}
}

func NewAuthorizationRuleProvider(registry RegistryFunc) AuthorizationRuleProvider {
	return &defaultAuthorizationRuleProvider{support: newRuleProviderSupport(registry)}
}

func NewRateLimitRuleProvider(registry RegistryFunc) RateLimitRuleProvider {
	return &defaultRateLimitRuleProvider{support: newRuleProviderSupport(registry)}
}

type ruleProviderSupport struct {
	registry RegistryFunc
	matcher  Matcher
}

func newRuleProviderSupport(registry RegistryFunc) ruleProviderSupport {
	return ruleProviderSupport{registry: registry}
}

func (s ruleProviderSupport) find(ctx context.Context, ruleTypes []RuleType, serviceName string, attributes map[string]string) ([]Rule, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	serviceName = strings.TrimSpace(serviceName)
	if serviceName == "" {
		return nil, ErrServiceNameRequired
	}
	rules := s.registry().ActiveRulesByTypes(ruleTypes, serviceName)
	matched := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		if s.matcher.Matches(rule, attributes) {
			matched = append(matched, rule)
		}
	}
	return matched, nil
}

type defaultRouteRuleProvider struct {
	support ruleProviderSupport
}

func (p *defaultRouteRuleProvider) Find(ctx context.Context, query RouteRuleQuery) ([]Rule, error) {
	return p.support.find(ctx, []RuleType{RuleTypeRoute}, query.ServiceName, query.AttributesForMatch())
}

func (p *defaultRouteRuleProvider) First(ctx context.Context, query RouteRuleQuery) (Rule, bool, error) {
	rules, err := p.Find(ctx, query)
	return firstRule(rules, err)
}

type defaultCircuitBreakerRuleProvider struct {
	support ruleProviderSupport
}

func (p *defaultCircuitBreakerRuleProvider) Find(ctx context.Context, query CircuitBreakerRuleQuery) ([]Rule, error) {
	return p.support.find(ctx, []RuleType{RuleTypeCircuitBreaker, RuleTypeDegrade}, query.ServiceName, query.AttributesForMatch())
}

func (p *defaultCircuitBreakerRuleProvider) First(ctx context.Context, query CircuitBreakerRuleQuery) (Rule, bool, error) {
	rules, err := p.Find(ctx, query)
	return firstRule(rules, err)
}

type defaultAuthorizationRuleProvider struct {
	support ruleProviderSupport
}

func (p *defaultAuthorizationRuleProvider) Find(ctx context.Context, query AuthorizationRuleQuery) ([]Rule, error) {
	return p.support.find(ctx, []RuleType{RuleTypeAuth}, query.ServiceName, query.AttributesForMatch())
}

func (p *defaultAuthorizationRuleProvider) First(ctx context.Context, query AuthorizationRuleQuery) (Rule, bool, error) {
	rules, err := p.Find(ctx, query)
	return firstRule(rules, err)
}

type defaultRateLimitRuleProvider struct {
	support ruleProviderSupport
}

func (p *defaultRateLimitRuleProvider) Find(ctx context.Context, query RateLimitRuleQuery) ([]Rule, error) {
	return p.support.find(ctx, []RuleType{RuleTypeRateLimit}, query.ServiceName, query.AttributesForMatch())
}

func (p *defaultRateLimitRuleProvider) First(ctx context.Context, query RateLimitRuleQuery) (Rule, bool, error) {
	rules, err := p.Find(ctx, query)
	return firstRule(rules, err)
}

func firstRule(rules []Rule, err error) (Rule, bool, error) {
	if err != nil {
		return Rule{}, false, err
	}
	if len(rules) == 0 {
		return Rule{}, false, nil
	}
	return rules[0], true, nil
}
