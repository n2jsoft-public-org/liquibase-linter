// Package rules provides best practice linting rules.
package rules

import (
	"regexp"
	"strings"

	"github.com/n2jsoft/liquibase-linter/internal/parser"
)

// MissingRollbackRule detects changesets without rollback instructions.
type MissingRollbackRule struct{}

// ID returns the rule identifier.
func (r *MissingRollbackRule) ID() string {
	return "missing-rollback"
}

// Name returns the rule name.
func (r *MissingRollbackRule) Name() string {
	return "Missing Rollback Detection"
}

// Description returns the rule description.
func (r *MissingRollbackRule) Description() string {
	return "Detects changesets that lack rollback instructions, making it difficult to revert changes"
}

// Severity returns the rule severity.
func (r *MissingRollbackRule) Severity() Severity {
	return SeverityWarning
}

// Check analyzes the changelog for missing rollbacks.
func (r *MissingRollbackRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	for _, cs := range changelog.ChangeSets {
		if !cs.HasRollback() {
			violations = append(violations, Violation{
				Rule:        r.ID(),
				Severity:    r.Severity(),
				Message:     "Changeset lacks rollback instructions",
				FilePath:    cs.FilePath,
				ChangeSetID: cs.ID,
				Author:      cs.Author,
			})
		}
	}

	return violations
}

// NonIdempotentChangesRule detects changes that may fail on re-run.
type NonIdempotentChangesRule struct{}

// ID returns the rule identifier.
func (r *NonIdempotentChangesRule) ID() string {
	return "non-idempotent"
}

// Name returns the rule name.
func (r *NonIdempotentChangesRule) Name() string {
	return "Non-Idempotent Changes Detection"
}

// Description returns the rule description.
func (r *NonIdempotentChangesRule) Description() string {
	return "Detects changes that may fail when re-run without proper preconditions or idempotent patterns"
}

// Severity returns the rule severity.
func (r *NonIdempotentChangesRule) Severity() Severity {
	return SeverityWarning
}

// Check analyzes the changelog for non-idempotent changes.
func (r *NonIdempotentChangesRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	// Operations that typically need preconditions to be idempotent
	riskyOperations := map[string]bool{
		"createtable":         true,
		"createindex":         true,
		"addforeignkey":       true,
		"addprimarykey":       true,
		"addcolumn":           true,
		"addnotnull":          true,
		"adduniqueconstraint": true,
	}

	for _, cs := range changelog.ChangeSets {
		// Skip if runAlways is set (indicates intentional re-run)
		if cs.RunAlways {
			continue
		}

		hasPreconditions := cs.HasPreconditions()

		for _, change := range cs.Changes {
			changeType := strings.ToLower(strings.ReplaceAll(change.Type, " ", ""))

			if riskyOperations[changeType] && !hasPreconditions {
				violations = append(violations, Violation{
					Rule:        r.ID(),
					Severity:    r.Severity(),
					Message:     "Change type '" + change.Type + "' may fail on re-run without preconditions",
					FilePath:    cs.FilePath,
					ChangeSetID: cs.ID,
					Author:      cs.Author,
				})
				break // Only report once per changeset
			}
		}
	}

	return violations
}

// NamingConventionRule enforces naming conventions for database objects.
type NamingConventionRule struct{}

// ID returns the rule identifier.
func (r *NamingConventionRule) ID() string {
	return "naming-convention"
}

// Name returns the rule name.
func (r *NamingConventionRule) Name() string {
	return "Naming Convention Enforcement"
}

// Description returns the rule description.
func (r *NamingConventionRule) Description() string {
	return "Enforces naming conventions for tables, columns, and indexes (snake_case, lowercase, no special characters)"
}

// Severity returns the rule severity.
func (r *NamingConventionRule) Severity() Severity {
	return SeverityInfo
}

// Check analyzes the changelog for naming convention violations.
func (r *NamingConventionRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	// Valid naming pattern: lowercase letters, numbers, underscores
	validPattern := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

	for _, cs := range changelog.ChangeSets {
		for _, change := range cs.Changes {
			var name string
			var objectType string

			// Check table names
			if change.TableName != "" {
				name = change.TableName
				objectType = "table"
			} else if change.IndexName != "" {
				name = change.IndexName
				objectType = "index"
			} else if change.ColumnName != "" {
				name = change.ColumnName
				objectType = "column"
			}

			if name != "" && !validPattern.MatchString(name) {
				violations = append(violations, Violation{
					Rule:        r.ID(),
					Severity:    r.Severity(),
					Message:     "Invalid naming convention for " + objectType + " '" + name + "' (should be lowercase snake_case)",
					FilePath:    cs.FilePath,
					ChangeSetID: cs.ID,
					Author:      cs.Author,
				})
			}
		}
	}

	return violations
}

// ChangesetDocumentationRule enforces changeset documentation.
type ChangesetDocumentationRule struct{}

// ID returns the rule identifier.
func (r *ChangesetDocumentationRule) ID() string {
	return "changeset-documentation"
}

// Name returns the rule name.
func (r *ChangesetDocumentationRule) Name() string {
	return "Changeset Documentation"
}

// Description returns the rule description.
func (r *ChangesetDocumentationRule) Description() string {
	return "Ensures changesets have proper documentation via comments"
}

// Severity returns the rule severity.
func (r *ChangesetDocumentationRule) Severity() Severity {
	return SeverityInfo
}

// Check analyzes the changelog for missing documentation.
func (r *ChangesetDocumentationRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	for _, cs := range changelog.ChangeSets {
		if strings.TrimSpace(cs.Comment) == "" {
			violations = append(violations, Violation{
				Rule:        r.ID(),
				Severity:    r.Severity(),
				Message:     "Changeset lacks documentation comment",
				FilePath:    cs.FilePath,
				ChangeSetID: cs.ID,
				Author:      cs.Author,
			})
		}
	}

	return violations
}

// ContextMisuseRule detects risky context usage.
type ContextMisuseRule struct{}

// ID returns the rule identifier.
func (r *ContextMisuseRule) ID() string {
	return "context-misuse"
}

// Name returns the rule name.
func (r *ContextMisuseRule) Name() string {
	return "Context Misuse Detection"
}

// Description returns the rule description.
func (r *ContextMisuseRule) Description() string {
	return "Detects dangerous operations without proper context restrictions (e.g., production changes without context)"
}

// Severity returns the rule severity.
func (r *ContextMisuseRule) Severity() Severity {
	return SeverityWarning
}

// Check analyzes the changelog for context misuse.
func (r *ContextMisuseRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	dangerousOps := []string{"DROP", "TRUNCATE", "DELETE"}

	for _, cs := range changelog.ChangeSets {
		// If no context is specified, these operations could run in production
		if cs.Context == "" {
			for _, change := range cs.Changes {
				sqlUpper := strings.ToUpper(change.SQL)
				changeTypeUpper := strings.ToUpper(change.Type)

				for _, op := range dangerousOps {
					if strings.Contains(sqlUpper, op) || strings.Contains(changeTypeUpper, op) {
						violations = append(violations, Violation{
							Rule:        r.ID(),
							Severity:    r.Severity(),
							Message:     "Dangerous operation without context restriction (may run in production)",
							FilePath:    cs.FilePath,
							ChangeSetID: cs.ID,
							Author:      cs.Author,
						})
						break
					}
				}
			}
		}
	}

	return violations
}
