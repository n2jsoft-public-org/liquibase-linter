package rules

import (
	"testing"

	"github.com/n2jsoft-public-org/liquibase-linter/internal/parser"
)

func TestRuleRegistry_Register(t *testing.T) {
	registry := NewRuleRegistry()
	rule := &SQLInjectionRule{}

	registry.Register(rule)

	retrieved, exists := registry.GetRule(rule.ID())
	if !exists {
		t.Error("Rule should exist after registration")
	}
	if retrieved.ID() != rule.ID() {
		t.Errorf("Retrieved rule ID = %v, want %v", retrieved.ID(), rule.ID())
	}
}

func TestRuleRegistry_EnableDisable(t *testing.T) {
	registry := NewRuleRegistry()
	rule := &SQLInjectionRule{}
	registry.Register(rule)

	if !registry.IsEnabled(rule.ID()) {
		t.Error("Rule should be enabled by default")
	}

	registry.Disable(rule.ID())
	if registry.IsEnabled(rule.ID()) {
		t.Error("Rule should be disabled after Disable()")
	}

	registry.Enable(rule.ID())
	if !registry.IsEnabled(rule.ID()) {
		t.Error("Rule should be enabled after Enable()")
	}
}

func TestRuleRegistry_GetEnabledRules(t *testing.T) {
	registry := NewRuleRegistry()

	rule1 := &SQLInjectionRule{}
	rule2 := &MissingRollbackRule{}

	registry.Register(rule1)
	registry.Register(rule2)
	registry.Disable(rule2.ID())

	enabledRules := registry.GetEnabledRules()
	if len(enabledRules) != 1 {
		t.Errorf("Expected 1 enabled rule, got %d", len(enabledRules))
	}
	if enabledRules[0].ID() != rule1.ID() {
		t.Errorf("Enabled rule ID = %v, want %v", enabledRules[0].ID(), rule1.ID())
	}
}

func TestRuleRegistry_CheckChangelog(t *testing.T) {
	registry := NewRuleRegistry()
	registry.Register(&MissingRollbackRule{})

	changelog := &parser.Changelog{
		FilePath: "test.xml",
		ChangeSets: []parser.ChangeSet{
			{
				ID:       "1",
				Author:   "test",
				Rollback: nil,
			},
		},
	}

	violations := registry.CheckChangelog(changelog)
	if len(violations) == 0 {
		t.Error("Expected violations for changeset without rollback")
	}
	if violations[0].Rule != "missing-rollback" {
		t.Errorf("Violation rule = %v, want missing-rollback", violations[0].Rule)
	}
}

func TestSeverity_String(t *testing.T) {
	tests := []struct {
		severity Severity
		want     string
	}{
		{SeverityInfo, "info"},
		{SeverityWarning, "warning"},
		{SeverityCritical, "critical"},
		{Severity(99), "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.severity.String()
			if got != tt.want {
				t.Errorf("Severity.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRuleRegistry_GetAllRules(t *testing.T) {
	registry := NewRuleRegistry()

	rule1 := &SQLInjectionRule{}
	rule2 := &MissingRollbackRule{}
	rule3 := &MissingIndexRule{}

	registry.Register(rule1)
	registry.Register(rule2)
	registry.Register(rule3)

	allRules := registry.GetAllRules()
	if len(allRules) != 3 {
		t.Errorf("Expected 3 rules, got %d", len(allRules))
	}
}
