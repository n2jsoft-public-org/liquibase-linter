// Package rules provides the rules engine for analyzing Liquibase changelogs
// and identifying security vulnerabilities, anti-patterns, and best practice violations.
package rules

import (
	"bufio"
	"os"
	"strings"

	"github.com/n2jsoft/liquibase-linter/internal/parser"
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

// Violation represents a single rule violation found in a changelog.
type Violation struct {
	Rule        string
	Severity    Severity
	Message     string
	FilePath    string
	LineNumber  int
	Line        string // The actual line content that triggered the violation
	ChangeSetID string
	Author      string
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
	rules   map[string]Rule
	enabled map[string]bool
}

// NewRuleRegistry creates a new rule registry.
func NewRuleRegistry() *RuleRegistry {
	return &RuleRegistry{
		rules:   make(map[string]Rule),
		enabled: make(map[string]bool),
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

	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()

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
