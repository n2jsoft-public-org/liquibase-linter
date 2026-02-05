// Package rules provides the rules engine for analyzing Liquibase changelogs
// and identifying security vulnerabilities, anti-patterns, and best practice violations.
package rules

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/n2jsoft-public-org/liquibase-linter/internal/parser"
)

// Severity represents the severity level of a rule violation.
type Severity int

const (
	// SeverityInfo represents informational messages.
	SeverityInfo Severity = iota
	// SeverityWarning represents potential issues that should be reviewed.
	SeverityWarning
	// SeverityCritical represents critical issues that must be fixed.
	SeverityCritical
)

// String returns the string representation of the severity.
func (s Severity) String() string {
	switch s {
	case SeverityInfo:
		return "info"
	case SeverityWarning:
		return "warning"
	case SeverityCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// ParseSeverity converts a severity string to a Severity constant.
func ParseSeverity(s string) (Severity, error) {
	switch strings.ToLower(s) {
	case "info", "informational":
		return SeverityInfo, nil
	case "warning", "warn":
		return SeverityWarning, nil
	case "critical", "error":
		return SeverityCritical, nil
	default:
		return SeverityWarning, fmt.Errorf("invalid severity: %s", s)
	}
}

// Violation represents a single rule violation found in a changelog.
type Violation struct {
	Rule        string
	Message     string
	FilePath    string
	Line        string // The actual line content that triggered the violation
	ChangeSetID string
	Author      string
	Severity    Severity
	LineNumber  int
}

// Rule is the interface that all linting rules must implement.
type Rule interface {
	ID() string
	Name() string
	Description() string
	Severity() Severity
	Check(changelog *parser.Changelog) []Violation
}

// RuleRegistry manages all available rules.
type RuleRegistry struct {
	rules            map[string]Rule
	enabled          map[string]bool
	severityOverride map[string]Severity
}

// NewRuleRegistry creates a new rule registry.
func NewRuleRegistry() *RuleRegistry {
	return &RuleRegistry{
		rules:            make(map[string]Rule),
		enabled:          make(map[string]bool),
		severityOverride: make(map[string]Severity),
	}
}

// Register adds a rule to the registry.
func (r *RuleRegistry) Register(rule Rule) {
	r.rules[rule.ID()] = rule
	r.enabled[rule.ID()] = true
}

// Enable enables a rule by ID.
func (r *RuleRegistry) Enable(ruleID string) {
	if _, exists := r.rules[ruleID]; exists {
		r.enabled[ruleID] = true
	}
}

// Disable disables a rule by ID.
func (r *RuleRegistry) Disable(ruleID string) {
	if _, exists := r.rules[ruleID]; exists {
		r.enabled[ruleID] = false
	}
}

// IsEnabled checks if a rule is enabled.
func (r *RuleRegistry) IsEnabled(ruleID string) bool {
	return r.enabled[ruleID]
}

// SetSeverity sets a severity override for a specific rule.
func (r *RuleRegistry) SetSeverity(ruleID string, severity Severity) {
	if _, exists := r.rules[ruleID]; exists {
		r.severityOverride[ruleID] = severity
	}
}

// GetSeverity returns the effective severity for a rule (with override if present).
func (r *RuleRegistry) GetSeverity(ruleID string) Severity {
	if severity, hasOverride := r.severityOverride[ruleID]; hasOverride {
		return severity
	}
	if rule, exists := r.rules[ruleID]; exists {
		return rule.Severity()
	}
	return SeverityWarning // Default fallback
}

// GetRule returns a rule by ID.
func (r *RuleRegistry) GetRule(ruleID string) (Rule, bool) {
	rule, exists := r.rules[ruleID]
	return rule, exists
}

// GetAllRules returns all registered rules.
func (r *RuleRegistry) GetAllRules() []Rule {
	rules := make([]Rule, 0, len(r.rules))
	for _, rule := range r.rules {
		rules = append(rules, rule)
	}
	return rules
}

// GetEnabledRules returns all enabled rules.
func (r *RuleRegistry) GetEnabledRules() []Rule {
	rules := make([]Rule, 0)
	for id, rule := range r.rules {
		if r.enabled[id] {
			rules = append(rules, rule)
		}
	}
	return rules
}

// CheckChangelog runs all enabled rules against a changelog.
func (r *RuleRegistry) CheckChangelog(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	for _, rule := range r.GetEnabledRules() {
		ruleViolations := rule.Check(changelog)
		
		// Apply severity overrides if present
		ruleID := rule.ID()
		if severity, hasOverride := r.severityOverride[ruleID]; hasOverride {
			for i := range ruleViolations {
				ruleViolations[i].Severity = severity
			}
		}
		
		violations = append(violations, ruleViolations...)
	}

	return violations
}

// ReadLineFromFile reads a specific line from a file (1-indexed).
// Returns the line content with whitespace trimmed, or empty string if line not found.
func ReadLineFromFile(filePath string, lineNumber int) string {
	if lineNumber <= 0 {
		return ""
	}
	//nolint:gosec // G304: File path is provided by user for linting
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			// Log but don't fail - this is a best effort to read
			fmt.Fprintf(os.Stderr, "warning: failed to close file: %v\n", closeErr)
		}
	}()

	scanner := bufio.NewScanner(file)
	currentLine := 0

	for scanner.Scan() {
		currentLine++
		if currentLine == lineNumber {
			return strings.TrimSpace(scanner.Text())
		}
	}

	return ""
}

// FilterSuppressedViolations filters out violations for changesets that have suppressed
// the violated rule via inline comments (e.g., liquibase-linter:disable rule-id).
func FilterSuppressedViolations(violations []Violation, changelog *parser.Changelog) []Violation {
	if len(violations) == 0 {
		return violations
	}

	// Build a map of changeset ID -> changeset for quick lookup
	// Use FilePath+ID+Author as composite key to handle duplicate IDs across files
	changesetMap := make(map[string]*parser.ChangeSet)
	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
		key := cs.FilePath + ":" + cs.ID + ":" + cs.Author
		changesetMap[key] = cs
	}

	// Filter violations
	filtered := make([]Violation, 0, len(violations))
	for _, violation := range violations {
		key := violation.FilePath + ":" + violation.ChangeSetID + ":" + violation.Author
		if cs, exists := changesetMap[key]; exists {
			// Check if this rule is suppressed for this changeset
			if cs.IsSuppressed(violation.Rule) {
				continue // Skip this violation
			}
		}
		filtered = append(filtered, violation)
	}

	return filtered
}

// SuppressionWarning represents a warning about an invalid suppression directive.
type SuppressionWarning struct {
	ChangeSetID string
	Author      string
	FilePath    string
	RuleID      string // Invalid rule ID
	Message     string
}

// ValidateSuppressions checks all suppression directives in a changelog and returns
// warnings for any invalid rule IDs.
func ValidateSuppressions(changelog *parser.Changelog, registry *RuleRegistry) []SuppressionWarning {
	warnings := make([]SuppressionWarning, 0)

	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
		for _, suppressedRule := range cs.SuppressedRules {
			// Check if the rule exists in the registry
			if _, exists := registry.GetRule(suppressedRule); !exists {
				warnings = append(warnings, SuppressionWarning{
					ChangeSetID: cs.ID,
					Author:      cs.Author,
					FilePath:    cs.FilePath,
					RuleID:      suppressedRule,
					Message:     fmt.Sprintf("Unknown rule '%s' in suppression directive", suppressedRule),
				})
			}
		}
	}

	return warnings
}
