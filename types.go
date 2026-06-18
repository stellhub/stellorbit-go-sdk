package stellorbit

import (
	"github.com/stellhub/stellorbit-go-sdk/governance"
	"github.com/stellhub/stellorbit-go-sdk/internal/httpapi"
)

type RouteRequest = httpapi.RouteRequest
type APIResponse = httpapi.APIResponse
type HTTPError = httpapi.HTTPError

type RequestContext = governance.RequestContext
type RouteRuleQuery = governance.RouteRuleQuery
type CircuitBreakerRuleQuery = governance.CircuitBreakerRuleQuery
type AuthorizationRuleQuery = governance.AuthorizationRuleQuery
type RateLimitRuleQuery = governance.RateLimitRuleQuery

type GovernanceRule = governance.Rule
type GovernanceRuleType = governance.RuleType
type GovernanceRuleStatus = governance.RuleStatus
type GovernanceRuleRegistry = governance.Registry
type GovernanceRuleSource = governance.Source

type RouteRuleProvider = governance.RouteRuleProvider
type CircuitBreakerRuleProvider = governance.CircuitBreakerRuleProvider
type AuthorizationRuleProvider = governance.AuthorizationRuleProvider
type RateLimitRuleProvider = governance.RateLimitRuleProvider

const (
	GovernanceRuleTypeCircuitBreaker GovernanceRuleType = governance.RuleTypeCircuitBreaker
	GovernanceRuleTypeRateLimit      GovernanceRuleType = governance.RuleTypeRateLimit
	GovernanceRuleTypeAuth           GovernanceRuleType = governance.RuleTypeAuth
	GovernanceRuleTypeRoute          GovernanceRuleType = governance.RuleTypeRoute
	GovernanceRuleTypeDegrade        GovernanceRuleType = governance.RuleTypeDegrade

	GovernanceRuleStatusDraft    GovernanceRuleStatus = governance.RuleStatusDraft
	GovernanceRuleStatusActive   GovernanceRuleStatus = governance.RuleStatusActive
	GovernanceRuleStatusDisabled GovernanceRuleStatus = governance.RuleStatusDisabled
	GovernanceRuleStatusDeleted  GovernanceRuleStatus = governance.RuleStatusDeleted
)

func EmptyRequestContext() RequestContext {
	return governance.EmptyRequestContext()
}

func NewGovernanceRule(rule GovernanceRule) (GovernanceRule, error) {
	return governance.NewRule(rule)
}

func ParseGovernanceRuleType(value string) (GovernanceRuleType, error) {
	return governance.ParseRuleType(value)
}

func ParseGovernanceRuleStatus(value string) (GovernanceRuleStatus, error) {
	return governance.ParseRuleStatus(value)
}

func EmptyGovernanceRuleRegistry() GovernanceRuleRegistry {
	return governance.EmptyRegistry()
}

func NewGovernanceRuleRegistry(revision int64, checksum string, rules []GovernanceRule) GovernanceRuleRegistry {
	return governance.NewRegistry(revision, checksum, rules)
}

func NewInMemoryGovernanceRuleSource(registry GovernanceRuleRegistry) *governance.InMemorySource {
	return governance.NewInMemorySource(registry)
}

type GovernanceRuleParser = governance.Parser
type GovernanceRuleMatcher = governance.Matcher
type GovernanceRuleSnapshotParser = governance.SnapshotParser

func NewGovernanceRuleParser() GovernanceRuleParser {
	return governance.NewParser()
}

func NewGovernanceRuleSnapshotParser(parser GovernanceRuleParser, logger Logger) GovernanceRuleSnapshotParser {
	return governance.NewSnapshotParser(parser, logger)
}
