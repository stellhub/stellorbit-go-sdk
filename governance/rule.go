package governance

import (
	"fmt"
	"strings"
)

type RuleType string

const (
	RuleTypeCircuitBreaker RuleType = "CIRCUIT_BREAKER"
	RuleTypeRateLimit      RuleType = "RATE_LIMIT"
	RuleTypeAuth           RuleType = "AUTH"
	RuleTypeRoute          RuleType = "ROUTE"
	RuleTypeDegrade        RuleType = "DEGRADE"
)

type RuleStatus string

const (
	RuleStatusDraft    RuleStatus = "DRAFT"
	RuleStatusActive   RuleStatus = "ACTIVE"
	RuleStatusDisabled RuleStatus = "DISABLED"
	RuleStatusDeleted  RuleStatus = "DELETED"
)

type Rule struct {
	RuleID        string
	RuleName      string
	ConfigKey     string
	RuleType      RuleType
	TargetService string
	Status        RuleStatus
	Priority      int
	Revision      int64
	Checksum      string
	RawContent    string
	Content       map[string]any
}

func NewRule(rule Rule) (Rule, error) {
	rule.RuleID = strings.TrimSpace(rule.RuleID)
	if rule.RuleID == "" {
		return Rule{}, fmt.Errorf("governance rule id must not be blank")
	}
	rule.RuleName = defaultText(rule.RuleName, rule.RuleID)
	rule.ConfigKey = defaultText(rule.ConfigKey, rule.RuleID)
	if rule.RuleType == "" {
		return Rule{}, fmt.Errorf("governance rule type must not be blank")
	}
	rule.TargetService = strings.TrimSpace(rule.TargetService)
	if rule.TargetService == "" {
		return Rule{}, fmt.Errorf("governance rule target service must not be blank")
	}
	if rule.Status == "" {
		rule.Status = RuleStatusDraft
	}
	if !validRuleStatus(rule.Status) {
		return Rule{}, fmt.Errorf("unsupported governance rule status %q", rule.Status)
	}
	if rule.Priority < 0 {
		return Rule{}, fmt.Errorf("governance rule priority must be greater than or equal to 0")
	}
	rule.Content = copyAnyMap(rule.Content)
	if rule.RuleType == RuleTypeRateLimit {
		if err := requireRulePayload(rule.RuleType, rule.Content); err != nil {
			return Rule{}, err
		}
		if err := validateRateLimitContent(rule.Content); err != nil {
			return Rule{}, err
		}
	}
	return rule, nil
}

func (r Rule) Active() bool {
	return r.Status == RuleStatusActive
}

func (r Rule) MatchesService(serviceName string) bool {
	return r.TargetService == "*" || r.TargetService == serviceName
}

func (r Rule) FromConfig(configID string) bool {
	configID = strings.TrimSpace(configID)
	return configID != "" && (r.RuleID == configID || r.ConfigKey == configID)
}

func (r Rule) Clone() Rule {
	r.Content = copyAnyMap(r.Content)
	return r
}

func ParseRuleType(value string) (RuleType, error) {
	normalized := strings.ToUpper(strings.TrimSpace(strings.ReplaceAll(value, "-", "_")))
	switch normalized {
	case string(RuleTypeCircuitBreaker), "BREAKER", "CIRCUIT":
		return RuleTypeCircuitBreaker, nil
	case string(RuleTypeRateLimit), "RATELIMIT", "RATE_LIMITER":
		return RuleTypeRateLimit, nil
	case string(RuleTypeAuth), "AUTHORIZATION", "AUTH_POLICY":
		return RuleTypeAuth, nil
	case string(RuleTypeRoute), "ROUTING":
		return RuleTypeRoute, nil
	case string(RuleTypeDegrade):
		return RuleTypeDegrade, nil
	default:
		return "", fmt.Errorf("unsupported governance rule type %q", value)
	}
}

func ParseRuleStatus(value string) (RuleStatus, error) {
	normalized := strings.ToUpper(strings.TrimSpace(strings.ReplaceAll(value, "-", "_")))
	switch normalized {
	case "", string(RuleStatusDraft):
		return RuleStatusDraft, nil
	case string(RuleStatusActive), "ENABLED":
		return RuleStatusActive, nil
	case string(RuleStatusDisabled), "INACTIVE":
		return RuleStatusDisabled, nil
	case string(RuleStatusDeleted):
		return RuleStatusDeleted, nil
	default:
		return "", fmt.Errorf("unsupported governance rule status %q", value)
	}
}

func validRuleStatus(status RuleStatus) bool {
	switch status {
	case RuleStatusDraft, RuleStatusActive, RuleStatusDisabled, RuleStatusDeleted:
		return true
	default:
		return false
	}
}

func defaultText(value, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func copyAnyMap(values map[string]any) map[string]any {
	if len(values) == 0 {
		return map[string]any{}
	}
	copied := make(map[string]any, len(values))
	for key, value := range values {
		copied[key] = deepCopyAny(value)
	}
	return copied
}

func deepCopyAny(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		return copyAnyMap(typed)
	case []any:
		copied := make([]any, 0, len(typed))
		for _, item := range typed {
			copied = append(copied, deepCopyAny(item))
		}
		return copied
	default:
		return typed
	}
}
