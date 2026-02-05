// Package rules provides best practice linting rules.
package rules

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/n2jsoft/liquibase-linter/internal/config"
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

// SprintFolderStructureRule enforces sprint-based folder organization.
type SprintFolderStructureRule struct {
	config *config.FileStructureConfig
}

// NewSprintFolderStructureRule creates a new sprint folder structure rule with configuration.
func NewSprintFolderStructureRule(cfg *config.FileStructureConfig) *SprintFolderStructureRule {
	if cfg == nil {
		// Use default configuration
		cfg = &config.FileStructureConfig{
			Enabled:          true,
			SprintPattern:    `^v\d+$`,
			StructurePattern: `(?i)^\d+\s*-\s*structure$`,
			DataPattern:      `(?i)^\d+\s*-\s*data$`,
			ExcludePatterns:  []string{"**/init/**"},
			SprintBasePath:   "",
		}
	}
	return &SprintFolderStructureRule{config: cfg}
}

// ID returns the rule identifier.
func (r *SprintFolderStructureRule) ID() string {
	return "sprint-folder-structure"
}

// Name returns the rule name.
func (r *SprintFolderStructureRule) Name() string {
	return "Sprint Folder Structure"
}

// Description returns the rule description.
func (r *SprintFolderStructureRule) Description() string {
	return "Enforces sprint-based folder organization with proper structure and data subdirectories"
}

// Severity returns the rule severity.
func (r *SprintFolderStructureRule) Severity() Severity {
	return SeverityWarning
}

// Check analyzes the changelog for sprint folder structure violations.
func (r *SprintFolderStructureRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	if !r.config.Enabled {
		return violations
	}

	// Compile regex patterns
	sprintRe, err := regexp.Compile(r.config.SprintPattern)
	if err != nil {
		return violations // Skip if pattern is invalid
	}

	structureRe, err := regexp.Compile(r.config.StructurePattern)
	if err != nil {
		return violations
	}

	dataRe, err := regexp.Compile(r.config.DataPattern)
	if err != nil {
		return violations
	}

	for _, cs := range changelog.ChangeSets {
		// Check if file should be excluded
		if r.shouldExclude(cs.FilePath) {
			continue
		}

		// Check if file is within a sprint structure
		if !r.isInSprintStructure(cs.FilePath, sprintRe, structureRe, dataRe) {
			violations = append(violations, Violation{
				Rule:        r.ID(),
				Severity:    r.Severity(),
				Message:     "Changeset is not in a proper sprint folder structure (expected: sprint/[structure|data]/file)",
				FilePath:    cs.FilePath,
				ChangeSetID: cs.ID,
				Author:      cs.Author,
			})
		}
	}

	return violations
}

// shouldExclude checks if a file path should be excluded from this rule.
func (r *SprintFolderStructureRule) shouldExclude(filePath string) bool {
	// Normalize path for consistent matching
	normPath := filepath.ToSlash(filePath)

	for _, pattern := range r.config.ExcludePatterns {
		// Special case for **/init/** pattern - check if path contains /init/
		if pattern == "**/init/**" {
			if strings.Contains(normPath, "/init/") || strings.Contains(normPath, "\\init\\") {
				return true
			}
		}

		// Use MatchesResourceFilter for proper glob matching including **
		matched, err := parser.MatchesResourceFilter(normPath, pattern)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// isInSprintStructure checks if a file path follows the sprint/structure|data organization.
func (r *SprintFolderStructureRule) isInSprintStructure(filePath string, sprintRe, structureRe, dataRe *regexp.Regexp) bool {
	// Normalize path separators
	normPath := filepath.ToSlash(filePath)
	parts := strings.Split(normPath, "/")

	// Need at least 3 parts: sprint/type/file
	if len(parts) < 3 {
		return false
	}

	// If SprintBasePath is specified, ensure the file is under that path
	if r.config.SprintBasePath != "" {
		normalizedBase := filepath.ToSlash(r.config.SprintBasePath)
		if !strings.Contains(normPath, normalizedBase) {
			return false
		}
	}

	// Find sprint folder in path
	sprintIdx := -1
	for i, part := range parts {
		if sprintRe.MatchString(part) {
			sprintIdx = i
			break
		}
	}

	if sprintIdx == -1 {
		return false
	}

	// Check if there's a structure or data folder after the sprint folder
	if sprintIdx+1 < len(parts) {
		nextPart := parts[sprintIdx+1]
		return structureRe.MatchString(nextPart) || dataRe.MatchString(nextPart)
	}

	return false
}

// DDLLocationRule ensures DDL changes are in structure directories.
type DDLLocationRule struct {
	config *config.FileStructureConfig
}

// NewDDLLocationRule creates a new DDL location rule with configuration.
func NewDDLLocationRule(cfg *config.FileStructureConfig) *DDLLocationRule {
	if cfg == nil {
		cfg = &config.FileStructureConfig{
			Enabled:          true,
			StructurePattern: `(?i)^\d+\s*-\s*structure$`,
			ExcludePatterns:  []string{"**/init/**"},
		}
	}
	return &DDLLocationRule{config: cfg}
}

// ID returns the rule identifier.
func (r *DDLLocationRule) ID() string {
	return "ddl-location"
}

// Name returns the rule name.
func (r *DDLLocationRule) Name() string {
	return "DDL Change Location"
}

// Description returns the rule description.
func (r *DDLLocationRule) Description() string {
	return "Ensures DDL (Data Definition Language) changes are located in structure directories"
}

// Severity returns the rule severity.
func (r *DDLLocationRule) Severity() Severity {
	return SeverityWarning
}

// Check analyzes the changelog for DDL changes in wrong locations.
func (r *DDLLocationRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	if !r.config.Enabled {
		return violations
	}

	structureRe, err := regexp.Compile(r.config.StructurePattern)
	if err != nil {
		return violations
	}

	for _, cs := range changelog.ChangeSets {
		// Check if file should be excluded
		if r.shouldExclude(cs.FilePath) {
			continue
		}

		// Check each change in the changeset
		for _, change := range cs.Changes {
			if change.IsDDLChange() && !r.isInStructureDir(cs.FilePath, structureRe) {
				violations = append(violations, Violation{
					Rule:        r.ID(),
					Severity:    r.Severity(),
					Message:     "DDL change found outside structure directory (change type: " + change.GetChangeType() + ")",
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

// shouldExclude checks if a file path should be excluded from this rule.
func (r *DDLLocationRule) shouldExclude(filePath string) bool {
	// Normalize path for consistent matching
	normPath := filepath.ToSlash(filePath)

	for _, pattern := range r.config.ExcludePatterns {
		// Special case for **/init/** pattern - check if path contains /init/
		if pattern == "**/init/**" {
			if strings.Contains(normPath, "/init/") || strings.Contains(normPath, "\\init\\") {
				return true
			}
		}

		// Use MatchesResourceFilter for proper glob matching including **
		matched, err := parser.MatchesResourceFilter(normPath, pattern)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// isInStructureDir checks if a file path is in a structure directory.
func (r *DDLLocationRule) isInStructureDir(filePath string, structureRe *regexp.Regexp) bool {
	normPath := filepath.ToSlash(filePath)
	parts := strings.Split(normPath, "/")

	for _, part := range parts {
		if structureRe.MatchString(part) {
			return true
		}
	}

	return false
}

// DMLLocationRule ensures DML changes are in data directories.
type DMLLocationRule struct {
	config *config.FileStructureConfig
}

// NewDMLLocationRule creates a new DML location rule with configuration.
func NewDMLLocationRule(cfg *config.FileStructureConfig) *DMLLocationRule {
	if cfg == nil {
		cfg = &config.FileStructureConfig{
			Enabled:         true,
			DataPattern:     `(?i)^\d+\s*-\s*data$`,
			ExcludePatterns: []string{"**/init/**"},
		}
	}
	return &DMLLocationRule{config: cfg}
}

// ID returns the rule identifier.
func (r *DMLLocationRule) ID() string {
	return "dml-location"
}

// Name returns the rule name.
func (r *DMLLocationRule) Name() string {
	return "DML Change Location"
}

// Description returns the rule description.
func (r *DMLLocationRule) Description() string {
	return "Ensures DML (Data Manipulation Language) changes are located in data directories"
}

// Severity returns the rule severity.
func (r *DMLLocationRule) Severity() Severity {
	return SeverityWarning
}

// Check analyzes the changelog for DML changes in wrong locations.
func (r *DMLLocationRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	if !r.config.Enabled {
		return violations
	}

	dataRe, err := regexp.Compile(r.config.DataPattern)
	if err != nil {
		return violations
	}

	for _, cs := range changelog.ChangeSets {
		// Check if file should be excluded
		if r.shouldExclude(cs.FilePath) {
			continue
		}

		// Check each change in the changeset
		for _, change := range cs.Changes {
			if change.IsDMLChange() && !r.isInDataDir(cs.FilePath, dataRe) {
				violations = append(violations, Violation{
					Rule:        r.ID(),
					Severity:    r.Severity(),
					Message:     "DML change found outside data directory (change type: " + change.GetChangeType() + ")",
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

// shouldExclude checks if a file path should be excluded from this rule.
func (r *DMLLocationRule) shouldExclude(filePath string) bool {
	// Normalize path for consistent matching
	normPath := filepath.ToSlash(filePath)

	for _, pattern := range r.config.ExcludePatterns {
		// Special case for **/init/** pattern - check if path contains /init/
		if pattern == "**/init/**" {
			if strings.Contains(normPath, "/init/") || strings.Contains(normPath, "\\init\\") {
				return true
			}
		}

		// Use MatchesResourceFilter for proper glob matching including **
		matched, err := parser.MatchesResourceFilter(normPath, pattern)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// isInDataDir checks if a file path is in a data directory.
func (r *DMLLocationRule) isInDataDir(filePath string, dataRe *regexp.Regexp) bool {
	normPath := filepath.ToSlash(filePath)
	parts := strings.Split(normPath, "/")

	for _, part := range parts {
		if dataRe.MatchString(part) {
			return true
		}
	}

	return false
}
