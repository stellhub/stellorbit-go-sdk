package governance

import "strings"

type Matcher struct{}

func (m Matcher) Matches(rule Rule, attributes map[string]string) bool {
	if rule.RuleID == "" {
		return false
	}
	return m.ConditionsMatch(rule.Content, attributes)
}

func (m Matcher) ConditionsMatch(content map[string]any, attributes map[string]string) bool {
	conditions := conditionObject(content)
	if len(conditions) == 0 {
		return true
	}
	resolved := copyStringMap(attributes)
	for key, expected := range conditions {
		if !conditionMatches(resolved[key], expected) {
			return false
		}
	}
	return true
}

func conditionObject(content map[string]any) map[string]any {
	for _, key := range []string{"conditions", "match", "when"} {
		if value, ok := content[key]; ok {
			if object, ok := toStringAnyMap(value); ok {
				return object
			}
		}
	}
	return map[string]any{}
}

func conditionMatches(actual string, expected any) bool {
	if object, ok := toStringAnyMap(expected); ok {
		return mapConditionMatches(actual, object)
	}
	if values, ok := toAnySlice(expected); ok {
		actualValues := splitActualValues(actual)
		if len(actualValues) == 0 {
			return false
		}
		for _, expectedValue := range values {
			if actualValues[stringValue(expectedValue)] {
				return true
			}
		}
		return false
	}
	return splitActualValues(actual)[stringValue(expected)]
}

func mapConditionMatches(actual string, condition map[string]any) bool {
	actualValues := splitActualValues(actual)
	if value, ok := condition["exists"]; ok {
		exists := strings.EqualFold(stringValue(value), "true")
		return exists == (len(actualValues) > 0)
	}
	if value, ok := condition["equals"]; ok {
		return actualValues[stringValue(value)]
	}
	if value, ok := condition["notEquals"]; ok {
		return !actualValues[stringValue(value)]
	}
	if value, ok := condition["in"]; ok {
		return anyExpectedValueMatches(actualValues, value)
	}
	if value, ok := condition["notIn"]; ok {
		return !anyExpectedValueMatches(actualValues, value)
	}
	return false
}

func anyExpectedValueMatches(actualValues map[string]bool, value any) bool {
	values, ok := toAnySlice(value)
	if !ok {
		return actualValues[stringValue(value)]
	}
	if len(actualValues) == 0 {
		return false
	}
	for _, expectedValue := range values {
		if actualValues[stringValue(expectedValue)] {
			return true
		}
	}
	return false
}

func splitActualValues(actual string) map[string]bool {
	values := map[string]bool{}
	for _, part := range strings.Split(actual, ",") {
		value := strings.TrimSpace(part)
		if value != "" {
			values[value] = true
		}
	}
	return values
}

func toStringAnyMap(value any) (map[string]any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		return typed, true
	case map[any]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			result[stringValue(key)] = item
		}
		return result, true
	default:
		return nil, false
	}
}

func toAnySlice(value any) ([]any, bool) {
	switch typed := value.(type) {
	case []any:
		return typed, true
	case []string:
		values := make([]any, 0, len(typed))
		for _, item := range typed {
			values = append(values, item)
		}
		return values, true
	default:
		return nil, false
	}
}
