package governance

import "sort"

type Registry struct {
	Revision int64
	Checksum string
	Rules    []Rule
}

func EmptyRegistry() Registry {
	return Registry{Rules: []Rule{}}
}

func NewRegistry(revision int64, checksum string, rules []Rule) Registry {
	copied := make([]Rule, 0, len(rules))
	for _, rule := range rules {
		copied = append(copied, rule.Clone())
	}
	sort.SliceStable(copied, func(i, j int) bool {
		if copied[i].Priority != copied[j].Priority {
			return copied[i].Priority < copied[j].Priority
		}
		if copied[i].Revision != copied[j].Revision {
			return copied[i].Revision > copied[j].Revision
		}
		return copied[i].RuleID < copied[j].RuleID
	})
	return Registry{
		Revision: revision,
		Checksum: checksum,
		Rules:    copied,
	}
}

func (r Registry) Clone() Registry {
	return NewRegistry(r.Revision, r.Checksum, r.Rules)
}

func (r Registry) FindByID(ruleID string) (Rule, bool) {
	for _, rule := range r.Rules {
		if rule.RuleID == ruleID {
			return rule.Clone(), true
		}
	}
	return Rule{}, false
}

func (r Registry) ActiveRules(ruleType RuleType, serviceName string) []Rule {
	return r.ActiveRulesByTypes([]RuleType{ruleType}, serviceName)
}

func (r Registry) ActiveRulesByTypes(ruleTypes []RuleType, serviceName string) []Rule {
	acceptedTypes := make(map[RuleType]struct{}, len(ruleTypes))
	for _, ruleType := range ruleTypes {
		acceptedTypes[ruleType] = struct{}{}
	}
	matched := make([]Rule, 0)
	for _, rule := range r.Rules {
		if !rule.Active() || !rule.MatchesService(serviceName) {
			continue
		}
		if _, ok := acceptedTypes[rule.RuleType]; !ok {
			continue
		}
		matched = append(matched, rule.Clone())
	}
	return matched
}
