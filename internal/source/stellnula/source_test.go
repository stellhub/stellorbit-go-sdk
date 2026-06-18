package stellnula

import (
	"testing"

	stn "github.com/stellhub/stellnula-go-sdk"
	"github.com/stellhub/stellorbit-go-sdk/governance"
)

func TestReplaceRegistryIgnoresStaleSnapshot(t *testing.T) {
	rule, err := governance.NewRule(governance.Rule{
		RuleID:        "route-payment",
		RuleName:      "Route payment",
		ConfigKey:     "stellorbit.payment-service.route",
		RuleType:      governance.RuleTypeRoute,
		TargetService: "payment-service",
		Status:        governance.RuleStatusActive,
		Priority:      0,
		Revision:      10,
		Content: map[string]any{
			"ruleType":      "ROUTE",
			"targetService": "payment-service",
			"status":        "ACTIVE",
			"priority":      0,
			"routes":        []any{map[string]any{"id": "payment-v2"}},
		},
	})
	if err != nil {
		t.Fatalf("new rule: %v", err)
	}
	source := &Source{parser: governance.NewSnapshotParser(governance.NewParser(), nil)}
	source.registry.Store(governance.NewRegistry(10, "current", []governance.Rule{rule}))

	source.replaceRegistry(stn.Snapshot{
		Revision: 9,
		Checksum: "stale",
		Entries:  []stn.ConfigEntry{},
	})

	registry := governance.LoadRegistry(&source.registry)
	if registry.Revision != 10 {
		t.Fatalf("expected current revision to remain 10, got %d", registry.Revision)
	}
	if _, ok := registry.FindByID("route-payment"); !ok {
		t.Fatalf("expected current rule to remain after stale snapshot")
	}
}
