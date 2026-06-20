package governance

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Entry struct {
	ConfigID    string
	ConfigKey   string
	ContentType string
	Value       string
	Revision    int64
	Deleted     bool
}

type Snapshot struct {
	Revision int64
	Checksum string
	Entries  []Entry
}

type Parser struct{}

const aggregateSchemaVersion = "stellorbit.governance.aggregate.v1"

func NewParser() Parser {
	return Parser{}
}

func (p Parser) Parse(entry Entry, checksum string) (Rule, error) {
	rules, err := p.ParseAll(entry, checksum)
	if err != nil {
		return Rule{}, err
	}
	if len(rules) != 1 {
		return Rule{}, fmt.Errorf("aggregate governance config %s must contain exactly one rule, got %d", entryIdentity(entry), len(rules))
	}
	return rules[0], nil
}

func (p Parser) ParseAll(entry Entry, checksum string) ([]Rule, error) {
	raw := entry.Value
	var root map[string]any
	decoder := json.NewDecoder(bytes.NewBufferString(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&root); err != nil {
		return nil, fmt.Errorf("governance rule content must be valid JSON: %s: %w", entry.ConfigID, err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("governance rule content must contain a single JSON document: %s", entry.ConfigID)
	}
	rootRuleType, err := validateAggregatePayload(entry, root)
	if err != nil {
		return nil, err
	}
	ruleValues, err := arrayField(root, "rules")
	if err != nil {
		return nil, err
	}

	rules := make([]Rule, 0, len(ruleValues))
	for index, value := range ruleValues {
		ruleNode, ok := value.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("aggregate governance config %s rule at index %d must be an object", entry.ConfigID, index)
		}
		rule, err := parseAggregatedRule(entry, root, ruleNode, rootRuleType, checksum)
		if err != nil {
			return nil, fmt.Errorf("parse aggregate governance rule %s[%d]: %w", entry.ConfigID, index, err)
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

func validateAggregatePayload(entry Entry, root map[string]any) (RuleType, error) {
	configID := aggregateConfigID(entry)
	if configID == "" {
		return "", fmt.Errorf("aggregate governance config id must not be blank")
	}
	schemaVersion := textField(root, "schemaVersion")
	if schemaVersion != aggregateSchemaVersion {
		return "", fmt.Errorf("aggregate governance config %s schemaVersion must be %s", configID, aggregateSchemaVersion)
	}
	payloadConfigID := textField(root, "configId")
	if payloadConfigID == "" {
		return "", fmt.Errorf("aggregate governance config %s payload configId must not be blank", configID)
	}
	if payloadConfigID != configID {
		return "", fmt.Errorf("aggregate governance config %s payload configId mismatch: %s", configID, payloadConfigID)
	}
	ruleType, err := ParseRuleType(textField(root, "ruleType"))
	if err != nil {
		return "", err
	}
	if err := requireAggregatePayload(ruleType, root); err != nil {
		return "", err
	}
	return ruleType, nil
}

func parseAggregatedRule(entry Entry, root, ruleNode map[string]any, rootRuleType RuleType, snapshotChecksum string) (Rule, error) {
	content, err := objectField(ruleNode, "content")
	if err != nil {
		return Rule{}, err
	}
	entryConfigID := aggregateConfigID(entry)
	ruleConfigID := textField(ruleNode, "configId")
	if ruleConfigID == "" {
		ruleConfigID = entryConfigID
	}
	if ruleConfigID != entryConfigID {
		return Rule{}, fmt.Errorf("rule configId %s does not match aggregate configId %s", ruleConfigID, entryConfigID)
	}

	ruleID := textField(ruleNode, "ruleId")
	if ruleID == "" {
		return Rule{}, fmt.Errorf("ruleId must not be blank")
	}

	ruleType := rootRuleType
	if value := textField(ruleNode, "stellnulaRuleType"); value != "" {
		ruleType, err = ParseRuleType(value)
		if err != nil {
			return Rule{}, err
		}
	}

	targetService := firstNonBlank(textField(ruleNode, "targetService"), textField(content, "targetService"), textField(root, "targetService"))
	status, err := ParseRuleStatus(firstNonBlank(textField(ruleNode, "status"), textField(content, "status"), textField(root, "status")))
	if err != nil {
		return Rule{}, err
	}
	priority, err := priorityField(ruleNode, content, root)
	if err != nil {
		return Rule{}, err
	}
	ruleCode := textField(ruleNode, "ruleCode")
	aggregateChecksum := firstNonBlank(textField(root, "checksum"), snapshotChecksum)

	mergedContent := copyAnyMap(content)
	mergedContent["ruleType"] = string(ruleType)
	if sourceRuleType := firstNonBlank(textField(ruleNode, "ruleType"), textField(root, "sourceRuleType")); sourceRuleType != "" {
		mergedContent["sourceRuleType"] = sourceRuleType
	}
	if targetService != "" {
		mergedContent["targetService"] = targetService
	}
	mergedContent["status"] = string(status)
	if priority >= 0 {
		mergedContent["priority"] = priority
	}
	if ruleCode != "" {
		mergedContent["ruleCode"] = ruleCode
	}
	if schemaVersion := textField(ruleNode, "schemaVersion"); schemaVersion != "" {
		mergedContent["schemaVersion"] = schemaVersion
	}
	mergedContent["aggregateConfigId"] = entryConfigID
	if aggregateChecksum != "" {
		mergedContent["aggregateChecksum"] = aggregateChecksum
	}
	if err := requireRulePayload(ruleType, mergedContent); err != nil {
		return Rule{}, err
	}
	if ruleType == RuleTypeRateLimit {
		if err := validateRateLimitContent(mergedContent); err != nil {
			return Rule{}, err
		}
	}
	rawContent, err := marshalJSON(mergedContent)
	if err != nil {
		return Rule{}, err
	}
	rule := Rule{
		RuleID:        ruleID,
		RuleName:      firstNonBlank(textField(ruleNode, "ruleName", "name"), ruleCode),
		ConfigKey:     entryConfigID,
		RuleType:      ruleType,
		TargetService: targetService,
		Status:        status,
		Priority:      priority,
		Revision:      entry.Revision,
		Checksum:      firstNonBlank(textField(ruleNode, "checksum"), aggregateChecksum),
		RawContent:    rawContent,
		Content:       mergedContent,
	}
	return NewRule(rule)
}

func priorityField(ruleNode, content, root map[string]any) (int, error) {
	priority, err := intField(ruleNode, -1, "priority")
	if err != nil {
		return 0, err
	}
	if priority >= 0 {
		return priority, nil
	}
	priority, err = intField(content, -1, "priority")
	if err != nil {
		return 0, err
	}
	if priority >= 0 {
		return priority, nil
	}
	return intField(root, -1, "priority")
}

func requireAggregatePayload(ruleType RuleType, content map[string]any) error {
	field, err := payloadField(ruleType)
	if err != nil {
		return err
	}
	return requireAnyNode(content, field)
}

func requireRulePayload(ruleType RuleType, content map[string]any) error {
	field, err := payloadField(ruleType)
	if err != nil {
		return err
	}
	return requireAnyNode(content, field)
}

func payloadField(ruleType RuleType) (string, error) {
	switch ruleType {
	case RuleTypeRoute:
		return "routes", nil
	case RuleTypeRateLimit:
		return "limit", nil
	case RuleTypeCircuitBreaker:
		return "breaker", nil
	case RuleTypeAuth:
		return "auth", nil
	case RuleTypeDegrade:
		return "degrade", nil
	default:
		return "", fmt.Errorf("unsupported governance rule type %q", ruleType)
	}
}

func requireAnyNode(content map[string]any, keys ...string) error {
	for _, key := range keys {
		if value, ok := content[key]; ok && value != nil {
			return nil
		}
	}
	return fmt.Errorf("governance rule payload must include one of %s", strings.Join(keys, ", "))
}

func objectField(content map[string]any, key string) (map[string]any, error) {
	value, ok := content[key]
	if !ok || value == nil {
		return nil, fmt.Errorf("governance rule %s must be an object", key)
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("governance rule %s must be an object", key)
	}
	return object, nil
}

func arrayField(content map[string]any, key string) ([]any, error) {
	value, ok := content[key]
	if !ok || value == nil {
		return nil, fmt.Errorf("aggregate governance payload %s must be an array", key)
	}
	array, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("aggregate governance payload %s must be an array", key)
	}
	return array, nil
}

func textField(content map[string]any, keys ...string) string {
	for _, key := range keys {
		value, ok := content[key]
		if !ok || value == nil {
			continue
		}
		text := stringValue(value)
		if strings.TrimSpace(text) != "" {
			return strings.TrimSpace(text)
		}
	}
	return ""
}

func intField(content map[string]any, fallback int, keys ...string) (int, error) {
	for _, key := range keys {
		value, ok := content[key]
		if !ok || value == nil {
			continue
		}
		switch typed := value.(type) {
		case json.Number:
			number, err := typed.Int64()
			if err != nil {
				return 0, fmt.Errorf("governance rule %s must be an integer: %w", key, err)
			}
			return int(number), nil
		case float64:
			return int(typed), nil
		case int:
			return typed, nil
		case string:
			number, err := strconv.Atoi(strings.TrimSpace(typed))
			if err != nil {
				return 0, fmt.Errorf("governance rule %s must be an integer: %w", key, err)
			}
			return number, nil
		default:
			return 0, fmt.Errorf("governance rule %s must be an integer", key)
		}
	}
	return fallback, nil
}

func firstNonBlank(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func aggregateConfigID(entry Entry) string {
	return strings.TrimSpace(entry.ConfigID)
}

func marshalJSON(content map[string]any) (string, error) {
	raw, err := json.Marshal(content)
	if err != nil {
		return "", fmt.Errorf("marshal governance rule content: %w", err)
	}
	return string(raw), nil
}

func stringValue(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	case fmt.Stringer:
		return typed.String()
	case bool:
		return strconv.FormatBool(typed)
	case float64:
		return strconv.FormatFloat(typed, 'f', -1, 64)
	default:
		return fmt.Sprintf("%v", typed)
	}
}
