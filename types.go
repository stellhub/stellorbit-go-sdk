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
type RateLimitRule = governance.RateLimitRule
type KeyExtractor = governance.KeyExtractor
type KeyExtractorKey = governance.KeyExtractorKey
type LimitMode = governance.LimitMode
type LimitType = governance.LimitType
type LimitAlgorithm = governance.LimitAlgorithm
type TrafficProtocol = governance.TrafficProtocol
type ExecutionLocation = governance.ExecutionLocation
type CoordinationMode = governance.CoordinationMode
type EnforcementMode = governance.EnforcementMode
type KeyExtractorSource = governance.KeyExtractorSource
type RateLimitRuleQueryOption = governance.RateLimitRuleQueryOption

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

	LimitModeQPS         LimitMode = governance.LimitModeQPS
	LimitModeConcurrency LimitMode = governance.LimitModeConcurrency
	LimitModeHeader      LimitMode = governance.LimitModeHeader
	LimitModeHotKey      LimitMode = governance.LimitModeHotKey
	LimitModeCustom      LimitMode = governance.LimitModeCustom
	LimitModeQuota       LimitMode = governance.LimitModeQuota
	LimitModeBandwidth   LimitMode = governance.LimitModeBandwidth
	LimitModeConnection  LimitMode = governance.LimitModeConnection
	LimitModeModel       LimitMode = governance.LimitModeModel

	LimitTypeRequest          LimitType = governance.LimitTypeRequest
	LimitTypeConnection       LimitType = governance.LimitTypeConnection
	LimitTypeByte             LimitType = governance.LimitTypeByte
	LimitTypeTenant           LimitType = governance.LimitTypeTenant
	LimitTypeUser             LimitType = governance.LimitTypeUser
	LimitTypeCaller           LimitType = governance.LimitTypeCaller
	LimitTypeAPIKey           LimitType = governance.LimitTypeAPIKey
	LimitTypeResource         LimitType = governance.LimitTypeResource
	LimitTypeHeader           LimitType = governance.LimitTypeHeader
	LimitTypeGRPCMetadata     LimitType = governance.LimitTypeGRPCMetadata
	LimitTypeIP               LimitType = governance.LimitTypeIP
	LimitTypeEndpoint         LimitType = governance.LimitTypeEndpoint
	LimitTypeMethod           LimitType = governance.LimitTypeMethod
	LimitTypeTopic            LimitType = governance.LimitTypeTopic
	LimitTypeModelRequest     LimitType = governance.LimitTypeModelRequest
	LimitTypeModelToken       LimitType = governance.LimitTypeModelToken
	LimitTypeModelCost        LimitType = governance.LimitTypeModelCost
	LimitTypeModelConcurrency LimitType = governance.LimitTypeModelConcurrency
	LimitTypeCustomKey        LimitType = governance.LimitTypeCustomKey

	LimitAlgorithmTokenBucket      LimitAlgorithm = governance.LimitAlgorithmTokenBucket
	LimitAlgorithmLeakyBucket      LimitAlgorithm = governance.LimitAlgorithmLeakyBucket
	LimitAlgorithmFixedWindow      LimitAlgorithm = governance.LimitAlgorithmFixedWindow
	LimitAlgorithmSlidingWindow    LimitAlgorithm = governance.LimitAlgorithmSlidingWindow
	LimitAlgorithmQuotaLease       LimitAlgorithm = governance.LimitAlgorithmQuotaLease
	LimitAlgorithmConcurrencyLimit LimitAlgorithm = governance.LimitAlgorithmConcurrencyLimit
	LimitAlgorithmHotKey           LimitAlgorithm = governance.LimitAlgorithmHotKey
	LimitAlgorithmCustom           LimitAlgorithm = governance.LimitAlgorithmCustom
	LimitAlgorithmAdaptive         LimitAlgorithm = governance.LimitAlgorithmAdaptive

	TrafficProtocolHTTP    TrafficProtocol = governance.TrafficProtocolHTTP
	TrafficProtocolGRPC    TrafficProtocol = governance.TrafficProtocolGRPC
	TrafficProtocolTCP     TrafficProtocol = governance.TrafficProtocolTCP
	TrafficProtocolMessage TrafficProtocol = governance.TrafficProtocolMessage
	TrafficProtocolModel   TrafficProtocol = governance.TrafficProtocolModel
	TrafficProtocolAny     TrafficProtocol = governance.TrafficProtocolAny

	ExecutionLocationApplication ExecutionLocation = governance.ExecutionLocationApplication
	ExecutionLocationSidecar     ExecutionLocation = governance.ExecutionLocationSidecar
	ExecutionLocationGateway     ExecutionLocation = governance.ExecutionLocationGateway
	ExecutionLocationEdge        ExecutionLocation = governance.ExecutionLocationEdge

	CoordinationModeLocalOnly   CoordinationMode = governance.CoordinationModeLocalOnly
	CoordinationModeGlobalSync  CoordinationMode = governance.CoordinationModeGlobalSync
	CoordinationModeGlobalQuota CoordinationMode = governance.CoordinationModeGlobalQuota

	EnforcementModeLocal       EnforcementMode = governance.EnforcementModeLocal
	EnforcementModeGlobalSync  EnforcementMode = governance.EnforcementModeGlobalSync
	EnforcementModeGlobalQuota EnforcementMode = governance.EnforcementModeGlobalQuota
	EnforcementModeEdge        EnforcementMode = governance.EnforcementModeEdge

	KeyExtractorSourceHeader           KeyExtractorSource = governance.KeyExtractorSourceHeader
	KeyExtractorSourceGRPCMetadata     KeyExtractorSource = governance.KeyExtractorSourceGRPCMetadata
	KeyExtractorSourceHTTPPath         KeyExtractorSource = governance.KeyExtractorSourceHTTPPath
	KeyExtractorSourceGRPCMethod       KeyExtractorSource = governance.KeyExtractorSourceGRPCMethod
	KeyExtractorSourceRemoteIP         KeyExtractorSource = governance.KeyExtractorSourceRemoteIP
	KeyExtractorSourceTenant           KeyExtractorSource = governance.KeyExtractorSourceTenant
	KeyExtractorSourceUser             KeyExtractorSource = governance.KeyExtractorSourceUser
	KeyExtractorSourceCaller           KeyExtractorSource = governance.KeyExtractorSourceCaller
	KeyExtractorSourceAPIKey           KeyExtractorSource = governance.KeyExtractorSourceAPIKey
	KeyExtractorSourceEndpoint         KeyExtractorSource = governance.KeyExtractorSourceEndpoint
	KeyExtractorSourceMethod           KeyExtractorSource = governance.KeyExtractorSourceMethod
	KeyExtractorSourceTopic            KeyExtractorSource = governance.KeyExtractorSourceTopic
	KeyExtractorSourceModel            KeyExtractorSource = governance.KeyExtractorSourceModel
	KeyExtractorSourceModelRequest     KeyExtractorSource = governance.KeyExtractorSourceModelRequest
	KeyExtractorSourceModelToken       KeyExtractorSource = governance.KeyExtractorSourceModelToken
	KeyExtractorSourceModelCost        KeyExtractorSource = governance.KeyExtractorSourceModelCost
	KeyExtractorSourceModelConcurrency KeyExtractorSource = governance.KeyExtractorSourceModelConcurrency
	KeyExtractorSourceCustomExpression KeyExtractorSource = governance.KeyExtractorSourceCustomExpression
	KeyExtractorSourceCustomKey        KeyExtractorSource = governance.KeyExtractorSourceCustomKey
)

var (
	NewRateLimitRuleQuery = governance.NewRateLimitRuleQuery
	ByLimitMode           = governance.ByLimitMode
	ByProtocol            = governance.ByProtocol
	ByExecutionLocation   = governance.ByExecutionLocation
	ByCoordinationMode    = governance.ByCoordinationMode
	ByKeyExtractorSource  = governance.ByKeyExtractorSource
	DistributedOnly       = governance.DistributedOnly
	LocalOnly             = governance.LocalOnly
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
