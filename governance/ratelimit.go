package governance

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
)

type LimitMode string

const (
	LimitModeQPS         LimitMode = "QPS"
	LimitModeConcurrency LimitMode = "CONCURRENCY"
	LimitModeHeader      LimitMode = "HEADER"
	LimitModeHotKey      LimitMode = "HOT_KEY"
	LimitModeCustom      LimitMode = "CUSTOM"
	LimitModeQuota       LimitMode = "QUOTA"
	LimitModeBandwidth   LimitMode = "BANDWIDTH"
	LimitModeConnection  LimitMode = "CONNECTION"
	LimitModeModel       LimitMode = "MODEL"
)

type LimitType string

const (
	LimitTypeRequest          LimitType = "REQUEST"
	LimitTypeConnection       LimitType = "CONNECTION"
	LimitTypeByte             LimitType = "BYTE"
	LimitTypeTenant           LimitType = "TENANT"
	LimitTypeUser             LimitType = "USER"
	LimitTypeCaller           LimitType = "CALLER"
	LimitTypeAPIKey           LimitType = "API_KEY"
	LimitTypeResource         LimitType = "RESOURCE"
	LimitTypeHeader           LimitType = "HEADER"
	LimitTypeGRPCMetadata     LimitType = "GRPC_METADATA"
	LimitTypeIP               LimitType = "IP"
	LimitTypeEndpoint         LimitType = "ENDPOINT"
	LimitTypeMethod           LimitType = "METHOD"
	LimitTypeTopic            LimitType = "TOPIC"
	LimitTypeModelRequest     LimitType = "MODEL_REQUEST"
	LimitTypeModelToken       LimitType = "MODEL_TOKEN"
	LimitTypeModelCost        LimitType = "MODEL_COST"
	LimitTypeModelConcurrency LimitType = "MODEL_CONCURRENCY"
	LimitTypeCustomKey        LimitType = "CUSTOM_KEY"
)

type LimitAlgorithm string

const (
	LimitAlgorithmTokenBucket      LimitAlgorithm = "TOKEN_BUCKET"
	LimitAlgorithmLeakyBucket      LimitAlgorithm = "LEAKY_BUCKET"
	LimitAlgorithmFixedWindow      LimitAlgorithm = "FIXED_WINDOW"
	LimitAlgorithmSlidingWindow    LimitAlgorithm = "SLIDING_WINDOW"
	LimitAlgorithmQuotaLease       LimitAlgorithm = "QUOTA_LEASE"
	LimitAlgorithmConcurrencyLimit LimitAlgorithm = "CONCURRENCY_LIMIT"
	LimitAlgorithmHotKey           LimitAlgorithm = "HOT_KEY"
	LimitAlgorithmCustom           LimitAlgorithm = "CUSTOM"
	LimitAlgorithmAdaptive         LimitAlgorithm = "ADAPTIVE"
)

type TrafficProtocol string

const (
	TrafficProtocolHTTP    TrafficProtocol = "HTTP"
	TrafficProtocolGRPC    TrafficProtocol = "GRPC"
	TrafficProtocolTCP     TrafficProtocol = "TCP"
	TrafficProtocolMessage TrafficProtocol = "MESSAGE"
	TrafficProtocolModel   TrafficProtocol = "MODEL"
	TrafficProtocolAny     TrafficProtocol = "ANY"
)

type ExecutionLocation string

const (
	ExecutionLocationApplication ExecutionLocation = "APPLICATION"
	ExecutionLocationSidecar     ExecutionLocation = "SIDECAR"
	ExecutionLocationGateway     ExecutionLocation = "GATEWAY"
	ExecutionLocationEdge        ExecutionLocation = "EDGE"
)

type CoordinationMode string

const (
	CoordinationModeLocalOnly   CoordinationMode = "LOCAL_ONLY"
	CoordinationModeGlobalSync  CoordinationMode = "GLOBAL_SYNC"
	CoordinationModeGlobalQuota CoordinationMode = "GLOBAL_QUOTA"
)

type EnforcementMode string

const (
	EnforcementModeLocal       EnforcementMode = "LOCAL"
	EnforcementModeGlobalSync  EnforcementMode = "GLOBAL_SYNC"
	EnforcementModeGlobalQuota EnforcementMode = "GLOBAL_QUOTA"
	EnforcementModeEdge        EnforcementMode = "EDGE"
)

type KeyExtractorSource string

const (
	KeyExtractorSourceHeader           KeyExtractorSource = "HEADER"
	KeyExtractorSourceGRPCMetadata     KeyExtractorSource = "GRPC_METADATA"
	KeyExtractorSourceHTTPPath         KeyExtractorSource = "HTTP_PATH"
	KeyExtractorSourceGRPCMethod       KeyExtractorSource = "GRPC_METHOD"
	KeyExtractorSourceRemoteIP         KeyExtractorSource = "REMOTE_IP"
	KeyExtractorSourceTenant           KeyExtractorSource = "TENANT"
	KeyExtractorSourceUser             KeyExtractorSource = "USER"
	KeyExtractorSourceCaller           KeyExtractorSource = "CALLER"
	KeyExtractorSourceAPIKey           KeyExtractorSource = "API_KEY"
	KeyExtractorSourceEndpoint         KeyExtractorSource = "ENDPOINT"
	KeyExtractorSourceMethod           KeyExtractorSource = "METHOD"
	KeyExtractorSourceTopic            KeyExtractorSource = "TOPIC"
	KeyExtractorSourceModel            KeyExtractorSource = "MODEL"
	KeyExtractorSourceModelRequest     KeyExtractorSource = "MODEL_REQUEST"
	KeyExtractorSourceModelToken       KeyExtractorSource = "MODEL_TOKEN"
	KeyExtractorSourceModelCost        KeyExtractorSource = "MODEL_COST"
	KeyExtractorSourceModelConcurrency KeyExtractorSource = "MODEL_CONCURRENCY"
	KeyExtractorSourceCustomExpression KeyExtractorSource = "CUSTOM_EXPRESSION"
	KeyExtractorSourceCustomKey        KeyExtractorSource = "CUSTOM_KEY"
)

type RateLimitRule struct {
	LimitMode           LimitMode         `json:"limitMode,omitempty"`
	LimitType           LimitType         `json:"limitType,omitempty"`
	LimitAlgorithm      LimitAlgorithm    `json:"limitAlgorithm,omitempty"`
	TrafficProtocol     TrafficProtocol   `json:"trafficProtocol,omitempty"`
	ExecutionLocation   ExecutionLocation `json:"executionLocation,omitempty"`
	CoordinationMode    CoordinationMode  `json:"coordinationMode,omitempty"`
	EnforcementMode     EnforcementMode   `json:"enforcementMode,omitempty"`
	TargetSelector      map[string]any    `json:"targetSelector,omitempty"`
	RequestMatcher      map[string]any    `json:"requestMatcher,omitempty"`
	KeyExtractor        KeyExtractor      `json:"keyExtractor,omitempty"`
	Dimensions          []any             `json:"dimensions,omitempty"`
	QuotaConfig         map[string]any    `json:"quotaConfig,omitempty"`
	WindowConfig        map[string]any    `json:"windowConfig,omitempty"`
	BurstConfig         map[string]any    `json:"burstConfig,omitempty"`
	ConcurrencyConfig   map[string]any    `json:"concurrencyConfig,omitempty"`
	HotspotConfig       map[string]any    `json:"hotspotConfig,omitempty"`
	CustomPolicy        map[string]any    `json:"customPolicy,omitempty"`
	ModelLimitConfig    map[string]any    `json:"modelLimitConfig,omitempty"`
	FallbackPolicy      map[string]any    `json:"fallbackPolicy,omitempty"`
	ResponsePolicy      map[string]any    `json:"responsePolicy,omitempty"`
	ObservabilityConfig map[string]any    `json:"observabilityConfig,omitempty"`
	ShadowConfig        map[string]any    `json:"shadowConfig,omitempty"`
	Content             map[string]any    `json:"-"`
}

type KeyExtractor struct {
	Keys []KeyExtractorKey `json:"keys,omitempty"`
	Raw  map[string]any    `json:"-"`
}

type KeyExtractorKey struct {
	Name        string             `json:"name,omitempty"`
	Source      KeyExtractorSource `json:"source,omitempty"`
	Key         string             `json:"key,omitempty"`
	Required    bool               `json:"required,omitempty"`
	Normalize   any                `json:"normalize,omitempty"`
	Unsupported bool               `json:"unsupported,omitempty"`
	Raw         map[string]any     `json:"-"`
}

type RateLimitRuleQueryOption func(*RateLimitRuleQuery)

func NewRateLimitRuleQuery(serviceName string, options ...RateLimitRuleQueryOption) RateLimitRuleQuery {
	query := RateLimitRuleQuery{ServiceName: serviceName}
	for _, option := range options {
		if option != nil {
			option(&query)
		}
	}
	return query
}

func ByLimitMode(mode LimitMode) RateLimitRuleQueryOption {
	return func(query *RateLimitRuleQuery) {
		query.LimitMode = normalizedLimitMode(mode)
	}
}

func ByProtocol(protocol TrafficProtocol) RateLimitRuleQueryOption {
	return func(query *RateLimitRuleQuery) {
		query.TrafficProtocol = normalizedTrafficProtocol(protocol)
	}
}

func ByExecutionLocation(location ExecutionLocation) RateLimitRuleQueryOption {
	return func(query *RateLimitRuleQuery) {
		query.ExecutionLocation = normalizedExecutionLocation(location)
	}
}

func ByCoordinationMode(mode CoordinationMode) RateLimitRuleQueryOption {
	return func(query *RateLimitRuleQuery) {
		query.CoordinationMode = normalizedCoordinationMode(mode)
	}
}

func ByKeyExtractorSource(source KeyExtractorSource) RateLimitRuleQueryOption {
	return func(query *RateLimitRuleQuery) {
		query.KeyExtractorSource = normalizedKeyExtractorSource(source)
	}
}

func DistributedOnly() RateLimitRuleQueryOption {
	return func(query *RateLimitRuleQuery) {
		query.DistributedOnly = true
	}
}

func LocalOnly() RateLimitRuleQueryOption {
	return func(query *RateLimitRuleQuery) {
		query.LocalOnly = true
	}
}

func (r *RateLimitRule) UnmarshalJSON(raw []byte) error {
	var content map[string]any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&content); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return fmt.Errorf("rate limit rule must contain a single JSON document")
	}
	rule := newRateLimitRule(content)
	if err := validateRateLimitRule(rule); err != nil {
		return err
	}
	*r = rule
	return nil
}

func (r Rule) RateLimit() (RateLimitRule, bool) {
	if r.RuleType != RuleTypeRateLimit {
		return RateLimitRule{}, false
	}
	return newRateLimitRule(r.Content), true
}

func (r RateLimitRule) IsDistributedRule() bool {
	mode := r.EffectiveCoordinationMode()
	return mode == CoordinationModeGlobalSync || mode == CoordinationModeGlobalQuota
}

func (r RateLimitRule) IsLocalRuntimeRule() bool {
	return r.EffectiveCoordinationMode() == CoordinationModeLocalOnly
}

func (r RateLimitRule) EffectiveCoordinationMode() CoordinationMode {
	if r.CoordinationMode != "" {
		return r.CoordinationMode
	}
	switch r.EnforcementMode {
	case EnforcementModeGlobalSync:
		return CoordinationModeGlobalSync
	case EnforcementModeGlobalQuota:
		return CoordinationModeGlobalQuota
	case EnforcementModeLocal:
		return CoordinationModeLocalOnly
	default:
		return ""
	}
}

func (r RateLimitRule) HasKeyExtractorSource(source KeyExtractorSource) bool {
	source = normalizedKeyExtractorSource(source)
	if source == "" {
		return false
	}
	for _, key := range r.KeyExtractor.Keys {
		if normalizedKeyExtractorSource(key.Source) == source {
			return true
		}
	}
	return false
}

func (r RateLimitRule) UnsupportedKeyExtractorSources() []KeyExtractorKey {
	unsupported := make([]KeyExtractorKey, 0)
	for _, key := range r.KeyExtractor.Keys {
		if key.Unsupported {
			unsupported = append(unsupported, key)
		}
	}
	return unsupported
}

func newRateLimitRule(content map[string]any) RateLimitRule {
	payload := rateLimitPayload(content)
	rule := RateLimitRule{
		LimitMode:           normalizedLimitMode(LimitMode(firstNonBlank(textField(content, "limitMode"), textField(payload, "limitMode")))),
		LimitType:           normalizedLimitType(LimitType(firstNonBlank(textField(content, "limitType"), textField(payload, "limitType")))),
		LimitAlgorithm:      normalizedLimitAlgorithm(LimitAlgorithm(firstNonBlank(textField(content, "limitAlgorithm"), textField(payload, "limitAlgorithm")))),
		TrafficProtocol:     normalizedTrafficProtocol(TrafficProtocol(firstNonBlank(textField(content, "trafficProtocol"), textField(payload, "trafficProtocol")))),
		ExecutionLocation:   normalizedExecutionLocation(ExecutionLocation(firstNonBlank(textField(content, "executionLocation"), textField(payload, "executionLocation")))),
		CoordinationMode:    normalizedCoordinationMode(CoordinationMode(firstNonBlank(textField(content, "coordinationMode"), textField(payload, "coordinationMode")))),
		EnforcementMode:     normalizedEnforcementMode(EnforcementMode(firstNonBlank(textField(content, "enforcementMode"), textField(payload, "enforcementMode")))),
		TargetSelector:      mapField(content, payload, "targetSelector"),
		RequestMatcher:      mapField(content, payload, "requestMatcher"),
		KeyExtractor:        keyExtractorField(content, payload),
		Dimensions:          sliceField(content, payload, "dimensions"),
		QuotaConfig:         mapField(content, payload, "quotaConfig"),
		WindowConfig:        mapField(content, payload, "windowConfig"),
		BurstConfig:         mapField(content, payload, "burstConfig"),
		ConcurrencyConfig:   mapField(content, payload, "concurrencyConfig"),
		HotspotConfig:       mapField(content, payload, "hotspotConfig"),
		CustomPolicy:        mapField(content, payload, "customPolicy"),
		ModelLimitConfig:    mapField(content, payload, "modelLimitConfig"),
		FallbackPolicy:      mapField(content, payload, "fallbackPolicy"),
		ResponsePolicy:      mapField(content, payload, "responsePolicy"),
		ObservabilityConfig: mapField(content, payload, "observabilityConfig"),
		ShadowConfig:        mapField(content, payload, "shadowConfig"),
		Content:             copyAnyMap(content),
	}
	return rule
}

func validateRateLimitContent(content map[string]any) error {
	return validateRateLimitRule(newRateLimitRule(content))
}

func validateRateLimitRule(rule RateLimitRule) error {
	problems := make([]string, 0)
	requireRateLimitEnum(&problems, "limitMode", string(rule.LimitMode), validLimitMode)
	requireRateLimitEnum(&problems, "limitType", string(rule.LimitType), validLimitType)
	requireRateLimitEnum(&problems, "limitAlgorithm", string(rule.LimitAlgorithm), validLimitAlgorithm)
	requireRateLimitEnum(&problems, "trafficProtocol", string(rule.TrafficProtocol), validTrafficProtocol)
	requireRateLimitEnum(&problems, "executionLocation", string(rule.ExecutionLocation), validExecutionLocation)
	requireRateLimitEnum(&problems, "coordinationMode", string(rule.CoordinationMode), validCoordinationMode)
	validateOptionalRateLimitEnum(&problems, "enforcementMode", string(rule.EnforcementMode), validEnforcementMode)
	if len(problems) > 0 {
		return fmt.Errorf("invalid rate limit rule: %s", strings.Join(problems, "; "))
	}
	return nil
}

func requireRateLimitEnum(problems *[]string, field string, value string, valid func(string) bool) {
	if strings.TrimSpace(value) == "" {
		*problems = append(*problems, fmt.Sprintf("%s is required", field))
		return
	}
	validateOptionalRateLimitEnum(problems, field, value, valid)
}

func validateOptionalRateLimitEnum(problems *[]string, field string, value string, valid func(string) bool) {
	if strings.TrimSpace(value) == "" {
		return
	}
	if !valid(value) {
		*problems = append(*problems, fmt.Sprintf("%s %q is unsupported", field, value))
	}
}

func rateLimitRuleMatchesQuery(rule Rule, query RateLimitRuleQuery) bool {
	rateLimit, ok := rule.RateLimit()
	if !ok {
		return false
	}
	if query.LimitMode != "" && rateLimit.LimitMode != normalizedLimitMode(query.LimitMode) {
		return false
	}
	if query.TrafficProtocol != "" && rateLimit.TrafficProtocol != normalizedTrafficProtocol(query.TrafficProtocol) {
		return false
	}
	if query.ExecutionLocation != "" && rateLimit.ExecutionLocation != normalizedExecutionLocation(query.ExecutionLocation) {
		return false
	}
	if query.CoordinationMode != "" && rateLimit.EffectiveCoordinationMode() != normalizedCoordinationMode(query.CoordinationMode) {
		return false
	}
	if query.KeyExtractorSource != "" && !rateLimit.HasKeyExtractorSource(query.KeyExtractorSource) {
		return false
	}
	if query.DistributedOnly && !rateLimit.IsDistributedRule() {
		return false
	}
	if query.LocalOnly && !rateLimit.IsLocalRuntimeRule() {
		return false
	}
	return true
}

func rateLimitPayload(content map[string]any) map[string]any {
	if value, ok := content["limit"]; ok {
		if object, ok := toStringAnyMap(value); ok {
			return object
		}
	}
	return map[string]any{}
}

func keyExtractorField(content, payload map[string]any) KeyExtractor {
	extractor := KeyExtractor{Raw: mapField(content, payload, "keyExtractor")}
	keys := make([]KeyExtractorKey, 0)
	if values, ok := toAnySlice(extractor.Raw["keys"]); ok {
		for _, value := range values {
			object, ok := toStringAnyMap(value)
			if !ok {
				continue
			}
			keys = append(keys, keyExtractorKeyField(object))
		}
	}
	extractor.Keys = keys
	return extractor
}

func keyExtractorKeyField(content map[string]any) KeyExtractorKey {
	source, unsupported := parseKeyExtractorSource(textField(content, "source"))
	return KeyExtractorKey{
		Name:        textField(content, "name"),
		Source:      source,
		Key:         textField(content, "key"),
		Required:    boolField(content, "required"),
		Normalize:   deepCopyAny(content["normalize"]),
		Unsupported: unsupported,
		Raw:         copyAnyMap(content),
	}
}

func mapField(content, payload map[string]any, key string) map[string]any {
	if value, ok := content[key]; ok {
		if object, ok := toStringAnyMap(value); ok {
			return copyAnyMap(object)
		}
	}
	if value, ok := payload[key]; ok {
		if object, ok := toStringAnyMap(value); ok {
			return copyAnyMap(object)
		}
	}
	return map[string]any{}
}

func sliceField(content, payload map[string]any, key string) []any {
	if value, ok := content[key]; ok {
		if values, ok := toAnySlice(value); ok {
			return copyAnySlice(values)
		}
	}
	if value, ok := payload[key]; ok {
		if values, ok := toAnySlice(value); ok {
			return copyAnySlice(values)
		}
	}
	return []any{}
}

func copyAnySlice(values []any) []any {
	if len(values) == 0 {
		return []any{}
	}
	copied := make([]any, 0, len(values))
	for _, value := range values {
		copied = append(copied, deepCopyAny(value))
	}
	return copied
}

func boolField(content map[string]any, key string) bool {
	value, ok := content[key]
	if !ok || value == nil {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true")
	default:
		return strings.EqualFold(stringValue(typed), "true")
	}
}

func parseKeyExtractorSource(value string) (KeyExtractorSource, bool) {
	source := normalizedKeyExtractorSource(KeyExtractorSource(value))
	switch source {
	case "",
		KeyExtractorSourceHeader,
		KeyExtractorSourceGRPCMetadata,
		KeyExtractorSourceHTTPPath,
		KeyExtractorSourceGRPCMethod,
		KeyExtractorSourceRemoteIP,
		KeyExtractorSourceTenant,
		KeyExtractorSourceUser,
		KeyExtractorSourceCaller,
		KeyExtractorSourceAPIKey,
		KeyExtractorSourceEndpoint,
		KeyExtractorSourceMethod,
		KeyExtractorSourceTopic,
		KeyExtractorSourceModel,
		KeyExtractorSourceModelRequest,
		KeyExtractorSourceModelToken,
		KeyExtractorSourceModelCost,
		KeyExtractorSourceModelConcurrency,
		KeyExtractorSourceCustomExpression,
		KeyExtractorSourceCustomKey:
		return source, false
	default:
		return source, true
	}
}

func validLimitMode(value string) bool {
	switch LimitMode(normalizedEnumValue(value)) {
	case LimitModeQPS,
		LimitModeConcurrency,
		LimitModeHeader,
		LimitModeHotKey,
		LimitModeCustom,
		LimitModeQuota,
		LimitModeBandwidth,
		LimitModeConnection,
		LimitModeModel:
		return true
	default:
		return false
	}
}

func validLimitType(value string) bool {
	switch LimitType(normalizedEnumValue(value)) {
	case LimitTypeRequest,
		LimitTypeConnection,
		LimitTypeByte,
		LimitTypeTenant,
		LimitTypeUser,
		LimitTypeCaller,
		LimitTypeAPIKey,
		LimitTypeResource,
		LimitTypeHeader,
		LimitTypeGRPCMetadata,
		LimitTypeIP,
		LimitTypeEndpoint,
		LimitTypeMethod,
		LimitTypeTopic,
		LimitTypeModelRequest,
		LimitTypeModelToken,
		LimitTypeModelCost,
		LimitTypeModelConcurrency,
		LimitTypeCustomKey:
		return true
	default:
		return false
	}
}

func validLimitAlgorithm(value string) bool {
	switch LimitAlgorithm(normalizedEnumValue(value)) {
	case LimitAlgorithmTokenBucket,
		LimitAlgorithmLeakyBucket,
		LimitAlgorithmFixedWindow,
		LimitAlgorithmSlidingWindow,
		LimitAlgorithmQuotaLease,
		LimitAlgorithmConcurrencyLimit,
		LimitAlgorithmHotKey,
		LimitAlgorithmCustom,
		LimitAlgorithmAdaptive:
		return true
	default:
		return false
	}
}

func validTrafficProtocol(value string) bool {
	switch TrafficProtocol(normalizedEnumValue(value)) {
	case TrafficProtocolHTTP,
		TrafficProtocolGRPC,
		TrafficProtocolTCP,
		TrafficProtocolMessage,
		TrafficProtocolModel,
		TrafficProtocolAny:
		return true
	default:
		return false
	}
}

func validExecutionLocation(value string) bool {
	switch ExecutionLocation(normalizedEnumValue(value)) {
	case ExecutionLocationApplication,
		ExecutionLocationSidecar,
		ExecutionLocationGateway,
		ExecutionLocationEdge:
		return true
	default:
		return false
	}
}

func validCoordinationMode(value string) bool {
	switch CoordinationMode(normalizedEnumValue(value)) {
	case CoordinationModeLocalOnly,
		CoordinationModeGlobalSync,
		CoordinationModeGlobalQuota:
		return true
	default:
		return false
	}
}

func validEnforcementMode(value string) bool {
	switch EnforcementMode(normalizedEnumValue(value)) {
	case EnforcementModeLocal,
		EnforcementModeGlobalSync,
		EnforcementModeGlobalQuota,
		EnforcementModeEdge:
		return true
	default:
		return false
	}
}

func normalizedLimitMode(value LimitMode) LimitMode {
	return LimitMode(normalizedEnumValue(string(value)))
}

func normalizedLimitType(value LimitType) LimitType {
	return LimitType(normalizedEnumValue(string(value)))
}

func normalizedLimitAlgorithm(value LimitAlgorithm) LimitAlgorithm {
	return LimitAlgorithm(normalizedEnumValue(string(value)))
}

func normalizedTrafficProtocol(value TrafficProtocol) TrafficProtocol {
	return TrafficProtocol(normalizedEnumValue(string(value)))
}

func normalizedExecutionLocation(value ExecutionLocation) ExecutionLocation {
	return ExecutionLocation(normalizedEnumValue(string(value)))
}

func normalizedCoordinationMode(value CoordinationMode) CoordinationMode {
	return CoordinationMode(normalizedEnumValue(string(value)))
}

func normalizedEnforcementMode(value EnforcementMode) EnforcementMode {
	return EnforcementMode(normalizedEnumValue(string(value)))
}

func normalizedKeyExtractorSource(value KeyExtractorSource) KeyExtractorSource {
	return KeyExtractorSource(normalizedEnumValue(string(value)))
}

func normalizedEnumValue(value string) string {
	return strings.ToUpper(strings.TrimSpace(strings.ReplaceAll(value, "-", "_")))
}
