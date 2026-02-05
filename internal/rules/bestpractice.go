// Package rules provides best practice linting rules.
package rules

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/n2jsoft-public-org/liquibase-linter/internal/config"
	"github.com/n2jsoft-public-org/liquibase-linter/internal/parser"
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

	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
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
type NonIdempotentChangesRule struct {
	config config.RuleConfig
}

// NewNonIdempotentChangesRule creates a new NonIdempotentChangesRule with the given configuration.
func NewNonIdempotentChangesRule(cfg config.RuleConfig) *NonIdempotentChangesRule {
	// Default to risky-only mode if not specified
	if cfg.Mode == "" {
		cfg.Mode = config.ModeRiskyOnly
	}
	return &NonIdempotentChangesRule{config: cfg}
}

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

// shouldExcludeFile checks if a file should be excluded from the rule based on exclude patterns.
func (r *NonIdempotentChangesRule) shouldExcludeFile(filePath string) bool {
	normPath := filepath.ToSlash(filePath)

	for _, pattern := range r.config.ExcludePatterns {
		// Special case for **/init/** and **/seed/** patterns
		if pattern == "**/init/**" {
			if strings.Contains(normPath, "/init/") || strings.Contains(normPath, "\\init\\") {
				return true
			}
		}
		if pattern == "**/seed/**" {
			if strings.Contains(normPath, "/seed/") || strings.Contains(normPath, "\\seed\\") {
				return true
			}
		}

		// Use MatchesResourceFilter for proper glob matching
		matched, err := parser.MatchesResourceFilter(normPath, pattern)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// hasRiskyOperation checks if a changeset contains operations that typically need preconditions.
func (r *NonIdempotentChangesRule) hasRiskyOperation(cs *parser.ChangeSet) (hasRisky bool, operation string) {
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

	for _, change := range cs.Changes {
		changeType := strings.ToLower(strings.ReplaceAll(change.Type, " ", ""))
		if riskyOperations[changeType] {
			return true, change.Type
		}
	}
	return false, ""
}

// Check analyzes the changelog for non-idempotent changes.
func (r *NonIdempotentChangesRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
		// Skip if runAlways is set (indicates intentional re-run)
		if cs.RunAlways {
			continue
		}

		// Skip if file matches exclude patterns
		if r.shouldExcludeFile(cs.FilePath) {
			continue
		}

		hasPreconditions := cs.HasPreconditions()

		// Mode: all - require preconditions on ALL changesets
		if r.config.Mode == config.ModeAll {
			if !hasPreconditions {
				violations = append(violations, Violation{
					Rule:        r.ID(),
					Severity:    r.Severity(),
					Message:     "Changeset requires preconditions (mode: all)",
					FilePath:    cs.FilePath,
					ChangeSetID: cs.ID,
					Author:      cs.Author,
				})
			}
			continue
		}

		// Mode: risky-only (default) - require preconditions only for risky operations
		if r.config.Mode == config.ModeRiskyOnly {
			hasRisky, changeType := r.hasRiskyOperation(cs)
			if hasRisky && !hasPreconditions {
				violations = append(violations, Violation{
					Rule:        r.ID(),
					Severity:    r.Severity(),
					Message:     "Changeset with risky operation '" + changeType + "' requires preconditions (mode: risky-only)",
					FilePath:    cs.FilePath,
					ChangeSetID: cs.ID,
					Author:      cs.Author,
				})
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

	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
		for _, change := range cs.Changes {
			var name string
			var objectType string

			// Check table names
			switch {
			case change.TableName != "":
				name = change.TableName
				objectType = "table"
			case change.IndexName != "":
				name = change.IndexName
				objectType = "index"
			case change.ColumnName != "":
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

	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
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

	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
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

	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
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

	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
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

	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
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

// LabelPatternRule enforces label naming conventions.
type LabelPatternRule struct {
	patterns        []*regexp.Regexp
	excludePatterns []string
	requireLabel    bool
}

// NewLabelPatternRule creates a new label pattern rule with configuration.
func NewLabelPatternRule(cfg *config.LabelPatternConfig) *LabelPatternRule {
	if cfg == nil {
		cfg = &config.LabelPatternConfig{
			Pattern:         `^v\d+$`,
			RequireLabel:    true,
			ExcludePatterns: []string{"**/init/**"},
		}
	}

	// Compile patterns
	var patterns []*regexp.Regexp

	// Single pattern
	if cfg.Pattern != "" {
		if re, err := regexp.Compile(cfg.Pattern); err == nil {
			patterns = append(patterns, re)
		}
	}

	// Multiple patterns
	for _, p := range cfg.Patterns {
		if re, err := regexp.Compile(p); err == nil {
			patterns = append(patterns, re)
		}
	}

	return &LabelPatternRule{
		patterns:        patterns,
		requireLabel:    cfg.RequireLabel,
		excludePatterns: cfg.ExcludePatterns,
	}
}

// ID returns the rule identifier.
func (r *LabelPatternRule) ID() string {
	return "label-pattern"
}

// Name returns the rule name.
func (r *LabelPatternRule) Name() string {
	return "Label Pattern Enforcement"
}

// Description returns the rule description.
func (r *LabelPatternRule) Description() string {
	return "Ensures changeset labels follow configured naming patterns"
}

// Severity returns the rule severity.
func (r *LabelPatternRule) Severity() Severity {
	return SeverityWarning
}

// Check analyzes the changelog for label pattern violations.
func (r *LabelPatternRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
		// Check if file should be excluded
		if r.shouldExclude(cs.FilePath) {
			continue
		}

		// Check if labels are required but missing
		if r.requireLabel && len(cs.Labels) == 0 {
			violations = append(violations, Violation{
				Rule:        r.ID(),
				Severity:    r.Severity(),
				Message:     "Changeset lacks required label",
				FilePath:    cs.FilePath,
				ChangeSetID: cs.ID,
				Author:      cs.Author,
			})
			continue
		}

		// Check each label against patterns
		for _, label := range cs.Labels {
			label = strings.TrimSpace(label)

			// Check for empty labels
			if label == "" {
				violations = append(violations, Violation{
					Rule:        r.ID(),
					Severity:    r.Severity(),
					Message:     "Changeset has empty label",
					FilePath:    cs.FilePath,
					ChangeSetID: cs.ID,
					Author:      cs.Author,
				})
				continue
			}

			// Check against patterns
			if !r.matchesAnyPattern(label) {
				violations = append(violations, Violation{
					Rule:        r.ID(),
					Severity:    r.Severity(),
					Message:     "Label '" + label + "' does not match required pattern",
					FilePath:    cs.FilePath,
					ChangeSetID: cs.ID,
					Author:      cs.Author,
				})
			}
		}
	}

	return violations
}

// matchesAnyPattern checks if a label matches any configured pattern.
func (r *LabelPatternRule) matchesAnyPattern(label string) bool {
	// If no patterns configured, accept any label
	if len(r.patterns) == 0 {
		return true
	}

	for _, pattern := range r.patterns {
		if pattern.MatchString(label) {
			return true
		}
	}
	return false
}

// shouldExclude checks if a file path should be excluded from this rule.
func (r *LabelPatternRule) shouldExclude(filePath string) bool {
	if len(r.excludePatterns) == 0 {
		return false
	}

	normPath := filepath.ToSlash(filePath)

	for _, pattern := range r.excludePatterns {
		// Special case for **/init/** pattern
		if pattern == "**/init/**" {
			if strings.Contains(normPath, "/init/") || strings.Contains(normPath, "\\init\\") {
				return true
			}
		}

		// Use MatchesResourceFilter for proper glob matching
		matched, err := parser.MatchesResourceFilter(normPath, pattern)
		if err == nil && matched {
			return true
		}
	}

	return false
}

// RedundantOnErrorHaltRule detects redundant onError:HALT in preconditions.
type RedundantOnErrorHaltRule struct{}

// NewRedundantOnErrorHaltRule creates a new redundant onError:HALT rule.
func NewRedundantOnErrorHaltRule() *RedundantOnErrorHaltRule {
	return &RedundantOnErrorHaltRule{}
}

// ID returns the unique identifier for this rule.
func (r *RedundantOnErrorHaltRule) ID() string {
	return "redundant-onerror-halt"
}

// Name returns the human-readable name of this rule.
func (r *RedundantOnErrorHaltRule) Name() string {
	return "Redundant onError:HALT Detection"
}

// Description returns a detailed description of what this rule checks.
func (r *RedundantOnErrorHaltRule) Description() string {
	return "Detects redundant onError:HALT configuration - HALT is the default and doesn't need to be specified"
}

// Severity returns the severity level of violations found by this rule.
func (r *RedundantOnErrorHaltRule) Severity() Severity {
	return SeverityInfo
}

// Check examines a changelog for violations of this rule.
func (r *RedundantOnErrorHaltRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
		// Only check changesets that have preconditions
		if cs.Preconditions == nil {
			continue
		}

		// Check if onError is explicitly set to "HALT"
		if strings.EqualFold(cs.Preconditions.OnError, "HALT") {
			violations = append(violations, Violation{
				Rule:        r.ID(),
				Severity:    r.Severity(),
				Message:     "Redundant 'onError=\"HALT\"' - this is the default behavior and can be omitted",
				FilePath:    cs.FilePath,
				ChangeSetID: cs.ID,
				Author:      cs.Author,
			})
		}
	}

	return violations
}

// NoIfExistsRule detects database-specific IF EXISTS patterns in SQL scripts
// and recommends using Liquibase preconditions instead.
type NoIfExistsRule struct{}

// NewNoIfExistsRule creates a new no-if-exists rule.
func NewNoIfExistsRule() *NoIfExistsRule {
	return &NoIfExistsRule{}
}

// ID returns the unique identifier for this rule.
func (r *NoIfExistsRule) ID() string {
	return "no-if-exists"
}

// Name returns the human-readable name of this rule.
func (r *NoIfExistsRule) Name() string {
	return "No IF EXISTS"
}

// Description returns a detailed description of what this rule checks.
func (r *NoIfExistsRule) Description() string {
	return "Detects database-specific IF EXISTS patterns and recommends Liquibase preconditions for cross-database compatibility"
}

// Severity returns the severity level of violations found by this rule.
func (r *NoIfExistsRule) Severity() Severity {
	return SeverityWarning
}

// Patterns to detect various IF EXISTS syntaxes across different databases
var ifExistsPatterns = []*regexp.Regexp{
	// SQL Server: IF EXISTS (SELECT ...) or IF NOT EXISTS (SELECT ...)
	regexp.MustCompile(`(?i)\bIF\s+(NOT\s+)?EXISTS\s*\(`),

	// SQL Server: IF OBJECT_ID('name', 'type') IS [NOT] NULL
	regexp.MustCompile(`(?i)\bIF\s+OBJECT_ID\s*\(`),

	// PostgreSQL: DO $$ BEGIN IF [NOT] EXISTS (...) THEN ... END IF; END $$;
	regexp.MustCompile(`(?i)\bDO\s+\$\$\s*BEGIN\s+IF\s+(NOT\s+)?EXISTS`),

	// MySQL: DROP PROCEDURE IF EXISTS / CREATE ... IF NOT EXISTS
	regexp.MustCompile(`(?i)\b(DROP|CREATE)\s+(PROCEDURE|FUNCTION|TABLE|INDEX|DATABASE|SCHEMA)\s+IF\s+(NOT\s+)?EXISTS\b`),

	// Generic: IF [NOT] EXISTS in procedural blocks
	regexp.MustCompile(`(?i)\bBEGIN\s+IF\s+(NOT\s+)?EXISTS\s*\(`),
}

// Check examines a changelog for violations of this rule.
func (r *NoIfExistsRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
		for _, change := range cs.Changes {
			// Only check SQL changes
			if change.SQL == "" {
				continue
			}

			// Preprocess SQL to remove comments and string literals
			cleanSQL := r.preprocessSQL(change.SQL)

			// Check for IF EXISTS patterns
			for _, pattern := range ifExistsPatterns {
				if match := pattern.FindString(cleanSQL); match != "" {
					violations = append(violations, Violation{
						Rule:        r.ID(),
						Severity:    r.Severity(),
						Message:     fmt.Sprintf("Use Liquibase preconditions instead of database-specific IF EXISTS (found: '%s'). Preconditions are cross-database compatible.", strings.TrimSpace(match)),
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
func (r *NoIfExistsRule) preprocessSQL(sql string) string {
	// Remove single-line comments (-- comment)
	sql = regexp.MustCompile(`--[^\n]*`).ReplaceAllString(sql, "")

	// Remove multi-line comments (/* comment */)
	sql = regexp.MustCompile(`/\*[\s\S]*?\*/`).ReplaceAllString(sql, "")

	// Remove single-quoted strings ('string')
	sql = regexp.MustCompile(`'(?:[^']|'')*'`).ReplaceAllString(sql, "")

	// Remove double-quoted strings ("string")
	sql = regexp.MustCompile(`"(?:[^"]|"")*"`).ReplaceAllString(sql, "")

	return sql
}

// AtomicChangesetRule enforces one change per changeset for better atomicity.
type AtomicChangesetRule struct {
	config *config.AtomicChangesetConfig
}

// NewAtomicChangesetRule creates a new atomic changeset rule.
func NewAtomicChangesetRule(cfg *config.AtomicChangesetConfig) *AtomicChangesetRule {
	return &AtomicChangesetRule{config: cfg}
}

// ID returns the unique identifier for this rule.
func (r *AtomicChangesetRule) ID() string {
	return "atomic-changeset"
}

// Name returns the human-readable name of this rule.
func (r *AtomicChangesetRule) Name() string {
	return "Atomic Changeset"
}

// Description returns a detailed description of what this rule checks.
func (r *AtomicChangesetRule) Description() string {
	return "Enforces that each changeset contains only a single change operation for better atomicity and rollback clarity"
}

// Severity returns the severity level of violations found by this rule.
func (r *AtomicChangesetRule) Severity() Severity {
	return SeverityInfo
}

// Check examines a changelog for violations of this rule.
func (r *AtomicChangesetRule) Check(changelog *parser.Changelog) []Violation {
	if !r.config.Enabled {
		return nil
	}

	violations := make([]Violation, 0)

	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
		// Check if file should be excluded
		if r.shouldExcludeFile(cs.FilePath) {
			continue
		}

		// Count changes in this changeset
		changeCount := r.countChanges(cs)

		if changeCount > 1 {
			changeTypes := r.getChangeTypes(cs)
			violations = append(violations, Violation{
				Rule:     r.ID(),
				Severity: r.Severity(),
				Message: fmt.Sprintf("Changeset contains %d changes (%s). Consider splitting into separate changesets for better atomicity and rollback clarity.",
					changeCount, strings.Join(changeTypes, ", ")),
				FilePath:    cs.FilePath,
				ChangeSetID: cs.ID,
				Author:      cs.Author,
			})
		}
	}

	return violations
}

// shouldExcludeFile checks if the file path matches any exclude patterns
func (r *AtomicChangesetRule) shouldExcludeFile(filePath string) bool {
	if len(r.config.ExcludePatterns) == 0 {
		return false
	}

	// Normalize path for consistent matching
	normPath := parser.NormalizePath(filePath)

	for _, pattern := range r.config.ExcludePatterns {
		// Use MatchesResourceFilter for proper glob matching including **
		matched, err := parser.MatchesResourceFilter(normPath, pattern)
		if err == nil && matched {
			return true
		}
	}

	return false
}

// countChanges counts the number of distinct changes in a changeset
func (r *AtomicChangesetRule) countChanges(cs *parser.ChangeSet) int {
	count := 0
	hasCreateTable := false
	indexCount := 0

	for _, change := range cs.Changes {
		changeType := change.GetChangeType()

		// Handle SQL changes specially - count statements
		if change.SQL != "" {
			sqlCount := r.countSQLStatements(change.SQL)
			if sqlCount > r.config.MaxSQLStatements {
				count += sqlCount
			} else {
				count++
			}
			continue
		}

		// Track table creation
		if changeType == "createTable" {
			hasCreateTable = true
			count++
			continue
		}

		// Track indexes
		if changeType == "createIndex" {
			indexCount++
			count++
			continue
		}

		// Count all other changes
		count++
	}

	// Apply configuration-based adjustments
	if hasCreateTable && indexCount > 0 && r.config.AllowTableWithIndexes {
		// Treat table + indexes as one logical operation
		count = 1
	}

	return count
}

// countSQLStatements counts the number of SQL statements in a SQL string
func (r *AtomicChangesetRule) countSQLStatements(sql string) int {
	// Remove comments and string literals to avoid false positives
	cleaned := r.preprocessSQL(sql)

	// Split by semicolon and count non-empty statements
	statements := strings.Split(cleaned, ";")
	count := 0

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt != "" {
			count++
		}
	}

	// If no semicolons found, it's likely a single statement
	if count == 0 && strings.TrimSpace(cleaned) != "" {
		count = 1
	}

	return count
}

// preprocessSQL removes comments and string literals to avoid false statement splits
func (r *AtomicChangesetRule) preprocessSQL(sql string) string {
	// Remove single-line comments (-- comment)
	sql = regexp.MustCompile(`--[^\n]*`).ReplaceAllString(sql, "")

	// Remove multi-line comments (/* comment */)
	sql = regexp.MustCompile(`/\*[\s\S]*?\*/`).ReplaceAllString(sql, "")

	// Remove single-quoted strings ('string')
	sql = regexp.MustCompile(`'(?:[^']|'')*'`).ReplaceAllString(sql, "")

	// Remove double-quoted strings ("string")
	sql = regexp.MustCompile(`"(?:[^"]|"")*"`).ReplaceAllString(sql, "")

	return sql
}

// getChangeTypes returns a list of change type names in the changeset
func (r *AtomicChangesetRule) getChangeTypes(cs *parser.ChangeSet) []string {
	types := make([]string, 0)
	seen := make(map[string]bool)

	for _, change := range cs.Changes {
		changeType := change.GetChangeType()
		if changeType == "" {
			changeType = "sql"
		}

		if !seen[changeType] {
			types = append(types, changeType)
			seen[changeType] = true
		}
	}

	return types
}
