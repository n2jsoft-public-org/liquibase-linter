// Package rules provides security-related linting rules.
package rules

import (
	"regexp"
	"strings"

	"github.com/n2jsoft-public-org/liquibase-linter/internal/parser"
)

// SQLInjectionRule detects potential SQL injection vulnerabilities.
type SQLInjectionRule struct{}

// ID returns the rule identifier.
func (r *SQLInjectionRule) ID() string {
	return "sql-injection"
}

// Name returns the rule name.
func (r *SQLInjectionRule) Name() string {
	return "SQL Injection Detection"
}

// Description returns the rule description.
func (r *SQLInjectionRule) Description() string {
	return "Detects potential SQL injection vulnerabilities from string concatenation or unsafe variable usage in SQL statements"
}

// Severity returns the rule severity.
func (r *SQLInjectionRule) Severity() Severity {
	return SeverityCritical
}

// Check analyzes the changelog for SQL injection risks.
func (r *SQLInjectionRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	// Patterns that indicate potential SQL injection
	injectionPatterns := []*regexp.Regexp{
		regexp.MustCompile(`\$\{[^}]+\}`),              // Variable interpolation: ${var}
		regexp.MustCompile(`\+\s*['"]|['"]\s*\+`),      // String concatenation with quotes
		regexp.MustCompile(`CONCAT\s*\(`),              // CONCAT function usage
		regexp.MustCompile(`\|\|\s*['"r]|['"]\s*\|\|`), // String concatenation with ||
	}

	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
		for _, change := range cs.Changes {
			sql := change.SQL

			// Check all SQL patterns
			for _, pattern := range injectionPatterns {
				if !pattern.MatchString(sql) {
					continue
				}
				// Find the line containing the pattern
				lines := strings.Split(sql, "\n")
				sqlSnippet := ""
				lineNum := 0
				for i, line := range lines {
					if pattern.MatchString(line) {
						sqlSnippet = strings.TrimSpace(line)
						lineNum = i + 1
						break
					}
				}
				// Fallback to first line if no specific line found
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
					Message:     "Potential SQL injection: SQL statement contains string concatenation or variable interpolation",
					FilePath:    cs.FilePath,
					LineNumber:  lineNum,
					Line:        sqlSnippet,
					ChangeSetID: cs.ID,
					Author:      cs.Author,
				})
				break // Only report once per change
			}
		}
	}

	return violations
}

// HardcodedCredentialsRule detects hardcoded passwords and API keys.
type HardcodedCredentialsRule struct{}

// ID returns the rule identifier.
func (r *HardcodedCredentialsRule) ID() string {
	return "hardcoded-credentials"
}

// Name returns the rule name.
func (r *HardcodedCredentialsRule) Name() string {
	return "Hardcoded Credentials Detection"
}

// Description returns the rule description.
func (r *HardcodedCredentialsRule) Description() string {
	return "Detects hardcoded passwords, API keys, and other sensitive credentials in SQL statements"
}

// Severity returns the rule severity.
func (r *HardcodedCredentialsRule) Severity() Severity {
	return SeverityCritical
}

// Check analyzes the changelog for hardcoded credentials.
func (r *HardcodedCredentialsRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	// Patterns for credential detection
	credentialPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)password\s*=\s*['"][^'"]+['"]`),
		regexp.MustCompile(`(?i)pwd\s*=\s*['"][^'"]+['"]`),
		regexp.MustCompile(`(?i)api[_-]?key\s*=\s*['"][^'"]+['"]`),
		regexp.MustCompile(`(?i)secret\s*=\s*['"][^'"]+['"]`),
		regexp.MustCompile(`(?i)token\s*=\s*['"][^'"]+['"]`),
		regexp.MustCompile(`(?i)IDENTIFIED\s+BY\s+['"][^'"]+['"]`),
	}

	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
		for _, change := range cs.Changes {
			sql := change.SQL

			for _, pattern := range credentialPatterns {
				if match := pattern.FindString(sql); match == "" {
					continue
				}
				match := pattern.FindString(sql)
				// Extract the line containing the credential
				lines := strings.Split(sql, "\n")
				sqlSnippet := ""
				lineNum := 0
				for i, line := range lines {
					if strings.Contains(line, match) {
						sqlSnippet = strings.TrimSpace(line)
						lineNum = i + 1
						if len(sqlSnippet) > 100 {
							sqlSnippet = sqlSnippet[:100] + "..."
						}
						break
					}
				}

				violations = append(violations, Violation{
					Rule:        r.ID(),
					Severity:    r.Severity(),
					Message:     "Hardcoded credentials detected in SQL statement",
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

	return violations
}

// DangerousOperationsRule detects dangerous SQL operations without preconditions.
type DangerousOperationsRule struct{}

// ID returns the rule identifier.
func (r *DangerousOperationsRule) ID() string {
	return "dangerous-operations"
}

// Name returns the rule name.
func (r *DangerousOperationsRule) Name() string {
	return "Dangerous Operations Detection"
}

// Description returns the rule description.
func (r *DangerousOperationsRule) Description() string {
	return "Detects dangerous operations like DROP, TRUNCATE without preconditions or context restrictions"
}

// Severity returns the rule severity.
func (r *DangerousOperationsRule) Severity() Severity {
	return SeverityCritical
}

// Check analyzes the changelog for dangerous operations.
func (r *DangerousOperationsRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	dangerousOps := []string{"DROP TABLE", "DROP DATABASE", "TRUNCATE", "DELETE FROM"}

	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
		// Check if changeset has preconditions or context
		hasSafeguards := cs.HasPreconditions() || cs.Context != ""

		for _, change := range cs.Changes {
			sqlUpper := strings.ToUpper(change.SQL)
			changeTypeUpper := strings.ToUpper(change.Type)

			// Check for dangerous operations
			isDangerous := false
			operation := ""

			for _, op := range dangerousOps {
				if strings.Contains(sqlUpper, op) || strings.Contains(changeTypeUpper, strings.ReplaceAll(op, " ", "")) {
					isDangerous = true
					operation = op
					break
				}
			}

			if isDangerous && !hasSafeguards {
				// Find the line containing the dangerous operation
				lines := strings.Split(change.SQL, "\n")
				sqlSnippet := ""
				lineNum := 0
				for i, line := range lines {
					if strings.Contains(strings.ToUpper(line), operation) {
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
					Message:     "Dangerous operation " + operation + " without preconditions or context restrictions",
					FilePath:    cs.FilePath,
					LineNumber:  lineNum,
					Line:        sqlSnippet,
					ChangeSetID: cs.ID,
					Author:      cs.Author,
				})
			}
		}
	}

	return violations
}

// PrivilegeEscalationRule detects excessive privilege grants.
type PrivilegeEscalationRule struct{}

// ID returns the rule identifier.
func (r *PrivilegeEscalationRule) ID() string {
	return "privilege-escalation"
}

// Name returns the rule name.
func (r *PrivilegeEscalationRule) Name() string {
	return "Privilege Escalation Detection"
}

// Description returns the rule description.
func (r *PrivilegeEscalationRule) Description() string {
	return "Detects excessive privilege grants like GRANT ALL, superuser creation, or broad wildcard permissions"
}

// Severity returns the rule severity.
func (r *PrivilegeEscalationRule) Severity() Severity {
	return SeverityCritical
}

// Check analyzes the changelog for privilege escalation risks.
func (r *PrivilegeEscalationRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	privilegePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)GRANT\s+ALL`),
		regexp.MustCompile(`(?i)GRANT\s+.*\s+ON\s+\*\.\*`),
		regexp.MustCompile(`(?i)WITH\s+GRANT\s+OPTION`),
		regexp.MustCompile(`(?i)SUPERUSER`),
		regexp.MustCompile(`(?i)CREATEDB.*CREATEROLE`),
		regexp.MustCompile(`(?i)@['"]%['"]`), // Wildcard host
	}

	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
		for _, change := range cs.Changes {
			sql := change.SQL

			for _, pattern := range privilegePatterns {
				if match := pattern.FindString(sql); match == "" {
					continue
				}
				match := pattern.FindString(sql)
				// Extract the line containing the privilege grant
				lines := strings.Split(sql, "\n")
				sqlSnippet := ""
				lineNum := 0
				for i, line := range lines {
					if strings.Contains(strings.ToUpper(line), strings.ToUpper(match)) {
						sqlSnippet = strings.TrimSpace(line)
						lineNum = i + 1
						if len(sqlSnippet) > 100 {
							sqlSnippet = sqlSnippet[:100] + "..."
						}
						break
					}
				}

				violations = append(violations, Violation{
					Rule:        r.ID(),
					Severity:    r.Severity(),
					Message:     "Excessive privilege grant detected",
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

	return violations
}
