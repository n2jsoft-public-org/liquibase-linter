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
		want     string
		severity Severity
	}{
		{want: "info", severity: SeverityInfo},
		{want: "warning", severity: SeverityWarning},
		{want: "critical", severity: SeverityCritical},
		{want: "unknown", severity: Severity(99)},
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

func TestFilterSuppressedViolations(t *testing.T) {
	//nolint:govet // Test struct field order optimized for readability over memory alignment
	tests := []struct {
		violations      []Violation
		expectedRuleIDs []string // Rule IDs of violations that should remain
		name            string
		changelog       *parser.Changelog
		expectedCount   int
	}{
		{
			name: "no suppressions - all violations remain",
			violations: []Violation{
				{Rule: "sql-injection", ChangeSetID: "1", Author: "john"},
				{Rule: "missing-rollback", ChangeSetID: "1", Author: "john"},
			},
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:              "1",
						Author:          "john",
						SuppressedRules: []string{},
					},
				},
			},
			expectedCount:   2,
			expectedRuleIDs: []string{"sql-injection", "missing-rollback"},
		},
		{
			name: "one rule suppressed",
			violations: []Violation{
				{Rule: "sql-injection", ChangeSetID: "1", Author: "john"},
				{Rule: "missing-rollback", ChangeSetID: "1", Author: "john"},
			},
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:              "1",
						Author:          "john",
						SuppressedRules: []string{"sql-injection"},
					},
				},
			},
			expectedCount:   1,
			expectedRuleIDs: []string{"missing-rollback"},
		},
		{
			name: "all rules suppressed",
			violations: []Violation{
				{Rule: "sql-injection", ChangeSetID: "1", Author: "john"},
				{Rule: "missing-rollback", ChangeSetID: "1", Author: "john"},
			},
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:              "1",
						Author:          "john",
						SuppressedRules: []string{"sql-injection", "missing-rollback"},
					},
				},
			},
			expectedCount:   0,
			expectedRuleIDs: []string{},
		},
		{
			name: "multiple changesets with different suppressions",
			violations: []Violation{
				{Rule: "sql-injection", ChangeSetID: "1", Author: "john"},
				{Rule: "missing-rollback", ChangeSetID: "1", Author: "john"},
				{Rule: "sql-injection", ChangeSetID: "2", Author: "jane"},
				{Rule: "missing-rollback", ChangeSetID: "2", Author: "jane"},
			},
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:              "1",
						Author:          "john",
						SuppressedRules: []string{"sql-injection"},
					},
					{
						ID:              "2",
						Author:          "jane",
						SuppressedRules: []string{"missing-rollback"},
					},
				},
			},
			expectedCount:   2,
			expectedRuleIDs: []string{"missing-rollback", "sql-injection"},
		},
		{
			name: "case insensitive suppression matching",
			violations: []Violation{
				{Rule: "sql-injection", ChangeSetID: "1", Author: "john"},
			},
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:              "1",
						Author:          "john",
						SuppressedRules: []string{"SQL-INJECTION"},
					},
				},
			},
			expectedCount:   0,
			expectedRuleIDs: []string{},
		},
		{
			name: "duplicate changeset IDs across different files - suppressions should work independently",
			violations: []Violation{
				{Rule: "sql-injection", ChangeSetID: "1", Author: "john", FilePath: "/path/to/file1.sql"},
				{Rule: "sql-injection", ChangeSetID: "1", Author: "john", FilePath: "/path/to/file2.sql"},
			},
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:              "1",
						Author:          "john",
						FilePath:        "/path/to/file1.sql",
						SuppressedRules: []string{"sql-injection"}, // This file suppresses the rule
					},
					{
						ID:              "1",
						Author:          "john",
						FilePath:        "/path/to/file2.sql",
						SuppressedRules: []string{}, // This file does NOT suppress the rule
					},
				},
			},
			expectedCount:   1, // Only the violation from file2 should remain
			expectedRuleIDs: []string{"sql-injection"},
		},
		{
			name:            "empty violations list",
			violations:      []Violation{},
			changelog:       &parser.Changelog{ChangeSets: []parser.ChangeSet{}},
			expectedCount:   0,
			expectedRuleIDs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := FilterSuppressedViolations(tt.violations, tt.changelog)
			if len(filtered) != tt.expectedCount {
				t.Errorf("FilterSuppressedViolations() returned %d violations, want %d",
					len(filtered), tt.expectedCount)
				return
			}

			// Check that the correct violations remain
			for i, violation := range filtered {
				found := false
				for _, expectedID := range tt.expectedRuleIDs {
					if violation.Rule == expectedID {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Unexpected violation[%d] with rule %q remained after filtering",
						i, violation.Rule)
				}
			}
		})
	}
}

func TestValidateSuppressions(t *testing.T) {
	//nolint:govet // Test struct field order optimized for readability over memory alignment
	tests := []struct {
		expectedRules []string // Invalid rule IDs
		name          string
		changelog     *parser.Changelog
		registry      *RuleRegistry
		expectedCount int
	}{
		{
			name: "all valid suppressions",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:              "1",
						Author:          "john",
						FilePath:        "test.xml",
						SuppressedRules: []string{"sql-injection", "missing-rollback"},
					},
				},
			},
			registry: func() *RuleRegistry {
				r := NewRuleRegistry()
				r.Register(&SQLInjectionRule{})
				r.Register(&MissingRollbackRule{})
				return r
			}(),
			expectedCount: 0,
			expectedRules: []string{},
		},
		{
			name: "one invalid suppression",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:              "1",
						Author:          "john",
						FilePath:        "test.xml",
						SuppressedRules: []string{"sql-injection", "invalid-rule"},
					},
				},
			},
			registry: func() *RuleRegistry {
				r := NewRuleRegistry()
				r.Register(&SQLInjectionRule{})
				return r
			}(),
			expectedCount: 1,
			expectedRules: []string{"invalid-rule"},
		},
		{
			name: "all invalid suppressions",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:              "1",
						Author:          "john",
						FilePath:        "test.xml",
						SuppressedRules: []string{"invalid-rule-1", "invalid-rule-2"},
					},
				},
			},
			registry: func() *RuleRegistry {
				r := NewRuleRegistry()
				r.Register(&SQLInjectionRule{})
				return r
			}(),
			expectedCount: 2,
			expectedRules: []string{"invalid-rule-1", "invalid-rule-2"},
		},
		{
			name: "multiple changesets with mixed validity",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:              "1",
						Author:          "john",
						FilePath:        "test.xml",
						SuppressedRules: []string{"sql-injection"},
					},
					{
						ID:              "2",
						Author:          "jane",
						FilePath:        "test.xml",
						SuppressedRules: []string{"invalid-rule"},
					},
				},
			},
			registry: func() *RuleRegistry {
				r := NewRuleRegistry()
				r.Register(&SQLInjectionRule{})
				return r
			}(),
			expectedCount: 1,
			expectedRules: []string{"invalid-rule"},
		},
		{
			name: "no suppressions",
			changelog: &parser.Changelog{
				ChangeSets: []parser.ChangeSet{
					{
						ID:              "1",
						Author:          "john",
						FilePath:        "test.xml",
						SuppressedRules: []string{},
					},
				},
			},
			registry: func() *RuleRegistry {
				r := NewRuleRegistry()
				r.Register(&SQLInjectionRule{})
				return r
			}(),
			expectedCount: 0,
			expectedRules: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := ValidateSuppressions(tt.changelog, tt.registry)
			if len(warnings) != tt.expectedCount {
				t.Errorf("ValidateSuppressions() returned %d warnings, want %d",
					len(warnings), tt.expectedCount)
				return
			}

			// Check that the correct invalid rules are reported
			for _, warning := range warnings {
				found := false
				for _, expectedRule := range tt.expectedRules {
					if warning.RuleID == expectedRule {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Unexpected warning for rule %q", warning.RuleID)
				}
			}
		})
	}
}
