package rules

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/n2jsoft-public-org/liquibase-linter/internal/config"
	"github.com/n2jsoft-public-org/liquibase-linter/internal/parser"
)

// NoManualTransactionsRule detects manual transaction control in SQL.
type NoManualTransactionsRule struct {
	patterns           []*regexp.Regexp
	caseInsensitive    bool
	excludeChangeTypes map[string]bool
	excludePatterns    []string
}

// NewNoManualTransactionsRule creates a new rule with configuration.
func NewNoManualTransactionsRule(cfg *config.NoManualTransactionsConfig) *NoManualTransactionsRule {
	if cfg == nil {
		cfg = &config.NoManualTransactionsConfig{
			Patterns: []string{
				`\bBEGIN\s+(TRANSACTION|TRAN|WORK)?\b`,
				`\bSTART\s+TRANSACTION\b`,
				`\bCOMMIT\s+(TRANSACTION|TRAN|WORK)?\b`,
				`\bROLLBACK(\s+(TRANSACTION|TRAN|WORK))?\b`,
				`\bSAVE(POINT)?\s+TRANSACTION\b`,
			},
			CaseInsensitive: true,
			ExcludeChangeTypes: []string{
				"createProcedure",
				"createFunction",
				"createTrigger",
			},
		}
	}

	// Compile patterns with optional case insensitivity
	var patterns []*regexp.Regexp
	for _, p := range cfg.Patterns {
		flags := ""
		if cfg.CaseInsensitive {
			flags = "(?i)"
		}
		if re, err := regexp.Compile(flags + p); err == nil {
			patterns = append(patterns, re)
		}
	}

	// Build exclude map for fast lookup
	excludeTypes := make(map[string]bool)
	for _, t := range cfg.ExcludeChangeTypes {
		excludeTypes[strings.ToLower(t)] = true
	}

	return &NoManualTransactionsRule{
		patterns:           patterns,
		caseInsensitive:    cfg.CaseInsensitive,
		excludeChangeTypes: excludeTypes,
		excludePatterns:    cfg.ExcludePatterns,
	}
}

// ID returns the unique identifier for this rule.
func (r *NoManualTransactionsRule) ID() string {
	return "no-manual-transactions"
}

// Name returns the human-readable name of this rule.
func (r *NoManualTransactionsRule) Name() string {
	return "No Manual Transaction Control"
}

// Description returns a detailed description of what this rule checks.
func (r *NoManualTransactionsRule) Description() string {
	return "Detects manual transaction control statements (BEGIN, COMMIT, ROLLBACK) that interfere with Liquibase's transaction management"
}

// Severity returns the severity level of violations found by this rule.
func (r *NoManualTransactionsRule) Severity() Severity {
	return SeverityWarning
}

// Check examines a changelog for violations of this rule.
func (r *NoManualTransactionsRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	for _, cs := range changelog.ChangeSets {
		// Check if file should be excluded
		if r.shouldExcludeFile(cs.FilePath) {
			continue
		}

		for _, change := range cs.Changes {
			// Check if change type should be excluded
			changeType := strings.ToLower(strings.ReplaceAll(change.Type, " ", ""))
			if r.excludeChangeTypes[changeType] {
				continue
			}

			// Extract SQL
			sql := change.SQL
			if sql == "" {
				continue
			}

			// Preprocess SQL to remove comments and string literals
			cleanSQL := r.preprocessSQL(sql)

			// Check for transaction keywords
			for _, pattern := range r.patterns {
				if match := pattern.FindString(cleanSQL); match != "" {
					violations = append(violations, Violation{
						Rule:        r.ID(),
						Severity:    r.Severity(),
						Message:     fmt.Sprintf("Manual transaction control detected: '%s' - let Liquibase manage transactions", strings.TrimSpace(match)),
						FilePath:    cs.FilePath,
						ChangeSetID: cs.ID,
						Author:      cs.Author,
					})
					break // Only report once per change
				}
			}
		}
	}

	return violations
}

// preprocessSQL removes comments and string literals to avoid false positives
func (r *NoManualTransactionsRule) preprocessSQL(sql string) string {
	// Remove single-line comments (-- comment)
	sql = regexp.MustCompile(`--[^\n]*`).ReplaceAllString(sql, "")

	// Remove multi-line comments (/* comment */)
	sql = regexp.MustCompile(`/\*[\s\S]*?\*/`).ReplaceAllString(sql, "")

	// Remove single-quoted strings ('string')
	// Use non-greedy matching and handle escaped quotes
	sql = regexp.MustCompile(`'(?:[^']|'')*'`).ReplaceAllString(sql, "")

	// Remove double-quoted strings ("string")
	sql = regexp.MustCompile(`"(?:[^"]|"")*"`).ReplaceAllString(sql, "")

	return sql
}

// shouldExcludeFile checks if a file path matches any exclude patterns
func (r *NoManualTransactionsRule) shouldExcludeFile(filePath string) bool {
	if len(r.excludePatterns) == 0 {
		return false
	}

	for _, pattern := range r.excludePatterns {
		matched, err := parser.MatchesResourceFilter(filePath, pattern)
		if err == nil && matched {
			return true
		}
	}

	return false
}
