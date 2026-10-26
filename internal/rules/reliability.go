package rules

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/n2jsoft-public-org/liquibase-linter/internal/config"
	"github.com/n2jsoft-public-org/liquibase-linter/internal/parser"
)

// NoManualTransactionsRule detects manual transaction control in SQL.
type NoManualTransactionsRule struct {
	patterns           []*regexp.Regexp
	excludeChangeTypes map[string]bool
	excludePatterns    []string
	caseInsensitive    bool
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

	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
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
	return matchesAnyPattern(filePath, r.excludePatterns)
}

// matchesAnyPattern checks if a file path matches any of the given glob patterns.
func matchesAnyPattern(filePath string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := parser.MatchesResourceFilter(filePath, pattern)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// DBMSScopeRule checks that changesets declare which database engines they run on,
// name only known engines, and keep engine-specific SQL out of the preconditions of
// changesets that run on other engines.
type DBMSScopeRule struct {
	allowed           map[string]bool
	forbiddenPatterns map[string][]*regexp.Regexp // engine -> compiled patterns
	engines           []string                    // keys of forbiddenPatterns, sorted for stable output
	excludePatterns   []string
	requireDBMS       bool
}

// NewDBMSScopeRule creates a new dbms-scope rule with configuration.
func NewDBMSScopeRule(cfg *config.DBMSScopeConfig) *DBMSScopeRule {
	if cfg == nil {
		cfg = &config.DBMSScopeConfig{RequireDBMS: true, AllowedValues: config.LiquibaseDBMSNames}
	}

	allowed := map[string]bool{"all": true, "none": true}
	for _, name := range cfg.AllowedValues {
		allowed[strings.ToLower(strings.TrimSpace(name))] = true
	}

	// Invalid patterns are rejected by config.Validate; skipping them here keeps the rule usable
	// when it is built from an unvalidated config.
	forbidden := make(map[string][]*regexp.Regexp)
	for engine, patterns := range cfg.ForbiddenPreconditionPatterns {
		engine = strings.ToLower(strings.TrimSpace(engine))
		for _, p := range patterns {
			if re, err := regexp.Compile(p); err == nil {
				forbidden[engine] = append(forbidden[engine], re)
			}
		}
	}
	engines := slices.Sorted(maps.Keys(forbidden))

	return &DBMSScopeRule{
		allowed:           allowed,
		forbiddenPatterns: forbidden,
		engines:           engines,
		excludePatterns:   cfg.ExcludePatterns,
		requireDBMS:       cfg.RequireDBMS,
	}
}

// ID returns the unique identifier for this rule.
func (r *DBMSScopeRule) ID() string {
	return "dbms-scope"
}

// Name returns the human-readable name of this rule.
func (r *DBMSScopeRule) Name() string {
	return "DBMS Scope"
}

// Description returns a detailed description of what this rule checks.
func (r *DBMSScopeRule) Description() string {
	return "Ensures changesets declare a valid dbms and keep engine-specific SQL out of preconditions of other engines"
}

// Severity returns the severity level of violations found by this rule.
func (r *DBMSScopeRule) Severity() Severity {
	return SeverityCritical
}

// Check examines a changelog for violations of this rule.
func (r *DBMSScopeRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
		if matchesAnyPattern(cs.FilePath, r.excludePatterns) {
			continue
		}

		dbms := strings.Join(cs.DBMSList, ",")
		ident := cs.Author + ":" + cs.ID

		if r.requireDBMS && strings.TrimSpace(dbms) == "" {
			violations = append(violations, r.violation(cs, fmt.Sprintf(
				"changeset %s has no dbms. Add dbms:<engine> to its header, or dbms:all for dialect-neutral SQL.", ident)))
		}

		for _, value := range cs.DBMSList {
			name := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(value)), "!")
			if name != "" && !r.allowed[name] {
				violations = append(violations, r.violation(cs, fmt.Sprintf(
					"changeset %s: unknown engine '%s' in dbms:%s. An unknown name matches no database, so Liquibase skips the changeset everywhere.",
					ident, name, dbms)))
			}
		}

		for _, engine := range r.engines {
			if !cs.ReachesDBMS(engine) {
				continue
			}
			if pattern := r.firstForbiddenMatch(engine, cs.PreconditionSQL); pattern != "" {
				violations = append(violations, r.violation(cs, fmt.Sprintf(
					"changeset %s runs on %s (dbms:%s) but a precondition matches '%s', which %s does not support. Restrict the changeset to other engines with dbms:, or rewrite the precondition.",
					ident, engine, dbmsOrAll(dbms), pattern, engine)))
			}
		}
	}

	return violations
}

// firstForbiddenMatch returns the first pattern of the engine matched by one of the SQL
// statements, or an empty string when none matches.
func (r *DBMSScopeRule) firstForbiddenMatch(engine string, sqls []string) string {
	for _, sql := range sqls {
		for _, re := range r.forbiddenPatterns[engine] {
			if re.MatchString(sql) {
				return re.String()
			}
		}
	}
	return ""
}

// violation builds a violation for the changeset, pointing at its header when the line is known.
func (r *DBMSScopeRule) violation(cs *parser.ChangeSet, message string) Violation {
	v := Violation{
		Rule:        r.ID(),
		Severity:    r.Severity(),
		Message:     message,
		FilePath:    cs.FilePath,
		ChangeSetID: cs.ID,
		Author:      cs.Author,
	}
	if cs.LineNumber > 0 {
		v.LineNumber = cs.LineNumber
		v.Line = ReadLineFromFile(cs.FilePath, cs.LineNumber)
	}
	return v
}

// dbmsOrAll returns the dbms list, or "all" when the changeset declares none.
func dbmsOrAll(dbms string) string {
	if strings.TrimSpace(dbms) == "" {
		return "all"
	}
	return dbms
}
