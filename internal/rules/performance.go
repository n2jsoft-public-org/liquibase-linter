// Package rules provides performance-related linting rules.
package rules

import (
	"regexp"
	"strings"

	"github.com/n2jsoft/liquibase-linter/internal/parser"
)

// MissingIndexRule detects tables without proper indexes on foreign keys.
type MissingIndexRule struct{}

// ID returns the rule identifier.
func (r *MissingIndexRule) ID() string {
	return "missing-index"
}

// Name returns the rule name.
func (r *MissingIndexRule) Name() string {
	return "Missing Index Detection"
}

// Description returns the rule description.
func (r *MissingIndexRule) Description() string {
	return "Detects foreign keys without corresponding indexes, which can cause performance issues"
}

// Severity returns the rule severity.
func (r *MissingIndexRule) Severity() Severity {
	return SeverityWarning
}

// Check analyzes the changelog for missing indexes.
func (r *MissingIndexRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	// Track foreign keys and indexes by table
	foreignKeys := make(map[string][]string)    // table -> column names
	indexes := make(map[string]map[string]bool) // table -> column -> exists

	// First pass: collect all foreign keys and indexes
	for _, cs := range changelog.ChangeSets {
		for _, change := range cs.Changes {
			tableName := strings.ToLower(change.TableName)

			if strings.Contains(strings.ToLower(change.Type), "foreignkey") {
				columnName := strings.ToLower(change.ColumnName)
				if columnName != "" {
					foreignKeys[tableName] = append(foreignKeys[tableName], columnName)
				}
			}

			if strings.Contains(strings.ToLower(change.Type), "index") {
				columnName := strings.ToLower(change.ColumnName)
				if columnName != "" {
					if indexes[tableName] == nil {
						indexes[tableName] = make(map[string]bool)
					}
					indexes[tableName][columnName] = true
				}
			}
		}
	}

	// Second pass: check for missing indexes on foreign keys
	for _, cs := range changelog.ChangeSets {
		for _, change := range cs.Changes {
			if strings.Contains(strings.ToLower(change.Type), "foreignkey") {
				tableName := strings.ToLower(change.TableName)
				columnName := strings.ToLower(change.ColumnName)

				if columnName != "" {
					// Check if an index exists for this column
					hasIndex := false
					if tableIndexes, exists := indexes[tableName]; exists {
						hasIndex = tableIndexes[columnName]
					}

					if !hasIndex {
						violations = append(violations, Violation{
							Rule:        r.ID(),
							Severity:    r.Severity(),
							Message:     "Foreign key on column '" + change.ColumnName + "' lacks a corresponding index",
							FilePath:    cs.FilePath,
							ChangeSetID: cs.ID,
							Author:      cs.Author,
						})
					}
				}
			}
		}
	}

	return violations
}

// TableLockRule detects operations that may cause table locks.
type TableLockRule struct{}

// ID returns the rule identifier.
func (r *TableLockRule) ID() string {
	return "table-lock"
}

// Name returns the rule name.
func (r *TableLockRule) Name() string {
	return "Table Lock Detection"
}

// Description returns the rule description.
func (r *TableLockRule) Description() string {
	return "Detects operations that may cause long table locks (e.g., ALTER TABLE on large tables without context)"
}

// Severity returns the rule severity.
func (r *TableLockRule) Severity() Severity {
	return SeverityWarning
}

// Check analyzes the changelog for table lock risks.
func (r *TableLockRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	lockingOperations := []string{
		"ALTER TABLE", "ADD COLUMN", "MODIFY COLUMN", "CHANGE COLUMN",
		"ADD CONSTRAINT", "ADD FOREIGN KEY", "ADD PRIMARY KEY",
	}

	for _, cs := range changelog.ChangeSets {
		// If no context, operation might run on production
		if cs.Context == "" {
			for _, change := range cs.Changes {
				sqlUpper := strings.ToUpper(change.SQL)
				changeTypeUpper := strings.ToUpper(change.Type)

				isLocking := false
				matchedOp := ""
				for _, op := range lockingOperations {
					if strings.Contains(sqlUpper, op) ||
						(strings.Contains(changeTypeUpper, "ADD") || strings.Contains(changeTypeUpper, "ALTER")) {
						isLocking = true
						matchedOp = op
						break
					}
				}

				if isLocking {
					// Find the line containing the locking operation
					lines := strings.Split(change.SQL, "\n")
					sqlSnippet := ""
					lineNum := 0
					for i, line := range lines {
						lineUpper := strings.ToUpper(line)
						if strings.Contains(lineUpper, matchedOp) || strings.Contains(lineUpper, "ALTER") || strings.Contains(lineUpper, "ADD") {
							sqlSnippet = strings.TrimSpace(line)
							lineNum = i + 1
							break
						}
					}
					// Fallback to first line if not found
					if sqlSnippet == "" && len(lines) > 0 {
						sqlSnippet = strings.TrimSpace(lines[0])
						lineNum = 1
					}
					if len(sqlSnippet) > 100 {
						sqlSnippet = sqlSnippet[:100] + "..."
					}

					violations = append(violations, Violation{
						Rule:        r.ID(),
						Severity:    r.Severity(),
						Message:     "Operation may cause table locks in production (consider using context or online DDL)",
						FilePath:    cs.FilePath,
						LineNumber:  lineNum,
						Line:        sqlSnippet,
						ChangeSetID: cs.ID,
						Author:      cs.Author,
					})
					break
				}
			}
		}
	}

	return violations
}

// LargeDataOperationRule detects potentially large data operations.
type LargeDataOperationRule struct{}

// ID returns the rule identifier.
func (r *LargeDataOperationRule) ID() string {
	return "large-data-operation"
}

// Name returns the rule name.
func (r *LargeDataOperationRule) Name() string {
	return "Large Data Operation Detection"
}

// Description returns the rule description.
func (r *LargeDataOperationRule) Description() string {
	return "Detects operations that may affect large amounts of data without proper safeguards (batching, limits)"
}

// Severity returns the rule severity.
func (r *LargeDataOperationRule) Severity() Severity {
	return SeverityInfo
}

// Check analyzes the changelog for large data operations.
func (r *LargeDataOperationRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	// Pattern to detect UPDATE/DELETE without WHERE clause
	updateWithoutWhere := regexp.MustCompile(`(?i)UPDATE\s+\w+\s+SET\s+.*(?:;|$)`)
	deleteWithoutWhere := regexp.MustCompile(`(?i)DELETE\s+FROM\s+\w+\s*(?:;|$)`)
	whereClause := regexp.MustCompile(`(?i)WHERE`)

	for _, cs := range changelog.ChangeSets {
		for _, change := range cs.Changes {
			sql := change.SQL

			// Check for UPDATE without WHERE
			if updateWithoutWhere.MatchString(sql) && !whereClause.MatchString(sql) {
				violations = append(violations, Violation{
					Rule:        r.ID(),
					Severity:    r.Severity(),
					Message:     "UPDATE statement without WHERE clause may affect all rows",
					FilePath:    cs.FilePath,
					ChangeSetID: cs.ID,
					Author:      cs.Author,
				})
			}

			// Check for DELETE without WHERE
			if deleteWithoutWhere.MatchString(sql) && !whereClause.MatchString(sql) {
				violations = append(violations, Violation{
					Rule:        r.ID(),
					Severity:    SeverityWarning, // More severe for DELETE
					Message:     "DELETE statement without WHERE clause will delete all rows",
					FilePath:    cs.FilePath,
					ChangeSetID: cs.ID,
					Author:      cs.Author,
				})
			}
		}
	}

	return violations
}

// SelectStarRule detects SELECT * in loadData or SQL statements.
type SelectStarRule struct{}

// ID returns the rule identifier.
func (r *SelectStarRule) ID() string {
	return "select-star"
}

// Name returns the rule name.
func (r *SelectStarRule) Name() string {
	return "SELECT * Detection"
}

// Description returns the rule description.
func (r *SelectStarRule) Description() string {
	return "Detects SELECT * usage which can cause performance issues and schema coupling"
}

// Severity returns the rule severity.
func (r *SelectStarRule) Severity() Severity {
	return SeverityInfo
}

// Check analyzes the changelog for SELECT * usage.
func (r *SelectStarRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	selectStarPattern := regexp.MustCompile(`(?i)SELECT\s+\*\s+FROM`)

	for _, cs := range changelog.ChangeSets {
		for _, change := range cs.Changes {
			if selectStarPattern.MatchString(change.SQL) {
				violations = append(violations, Violation{
					Rule:        r.ID(),
					Severity:    r.Severity(),
					Message:     "SELECT * detected; specify explicit columns for better performance and maintainability",
					FilePath:    cs.FilePath,
					ChangeSetID: cs.ID,
					Author:      cs.Author,
				})
			}
		}
	}

	return violations
}
