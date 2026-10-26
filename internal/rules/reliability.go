package rules

import (
	"fmt"
	"maps"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
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

// violation builds a violation of this rule for the changeset.
func (r *DBMSScopeRule) violation(cs *parser.ChangeSet, message string) Violation {
	return changeSetViolation(r, cs, message)
}

// changeSetViolation builds a violation of the rule for the changeset, pointing at its
// header when the line is known.
func changeSetViolation(rule Rule, cs *parser.ChangeSet, message string) Violation {
	v := Violation{
		Rule:        rule.ID(),
		Severity:    rule.Severity(),
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

// DBMSTwinRule checks that every changeset written for the source engine only has a
// counterpart (its "twin") for the target engine in the same file, and that twins
// target exactly the target engine.
type DBMSTwinRule struct {
	sincePath       *regexp.Regexp // nil when every file is in scope
	sourceDBMS      string
	targetDBMS      string
	idSuffix        string
	excludePatterns []string
	since           int
}

// NewDBMSTwinRule creates a new dbms-twin rule with configuration.
func NewDBMSTwinRule(cfg *config.DBMSTwinConfig) *DBMSTwinRule {
	r := &DBMSTwinRule{
		sourceDBMS:      strings.ToLower(strings.TrimSpace(cfg.SourceDBMS)),
		targetDBMS:      strings.ToLower(strings.TrimSpace(cfg.TargetDBMS)),
		idSuffix:        cfg.IDSuffix,
		excludePatterns: cfg.ExcludePatterns,
		since:           cfg.Since,
	}
	// An invalid pattern is rejected by config.Validate; ignoring it here checks every file.
	if cfg.SincePathPattern != "" {
		if re, err := regexp.Compile(cfg.SincePathPattern); err == nil && re.NumSubexp() > 0 {
			r.sincePath = re
		}
	}
	return r
}

// ID returns the unique identifier for this rule.
func (r *DBMSTwinRule) ID() string {
	return "dbms-twin"
}

// Name returns the human-readable name of this rule.
func (r *DBMSTwinRule) Name() string {
	return "DBMS Twin"
}

// Description returns a detailed description of what this rule checks.
func (r *DBMSTwinRule) Description() string {
	return "Ensures every changeset written for the source engine only has a twin for the target engine in the same file"
}

// Severity returns the severity level of violations found by this rule.
func (r *DBMSTwinRule) Severity() Severity {
	return SeverityCritical
}

// Check examines a changelog for violations of this rule.
func (r *DBMSTwinRule) Check(changelog *parser.Changelog) []Violation {
	violations := make([]Violation, 0)

	// Twins are looked up in the same file, by author and ID.
	byFile := make(map[string]map[string]*parser.ChangeSet)
	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
		if byFile[cs.FilePath] == nil {
			byFile[cs.FilePath] = make(map[string]*parser.ChangeSet)
		}
		byFile[cs.FilePath][cs.Author+":"+cs.ID] = cs
	}

	for i := range changelog.ChangeSets {
		cs := &changelog.ChangeSets[i]
		if !r.inScope(cs.FilePath) {
			continue
		}
		ident := cs.Author + ":" + cs.ID

		if strings.HasSuffix(cs.ID, r.idSuffix) {
			if !r.targetsOnly(cs) {
				violations = append(violations, changeSetViolation(r, cs, fmt.Sprintf(
					"changeset %s is a %s twin (%s) but declares dbms:%s. Use dbms:%s.",
					ident, r.targetDBMS, r.idSuffix, dbmsOrAll(strings.Join(cs.DBMSList, ",")), r.targetDBMS)))
			}
			continue
		}

		if !r.namesSource(cs) || cs.ReachesDBMS(r.targetDBMS) {
			continue
		}
		twin := byFile[cs.FilePath][ident+r.idSuffix]
		if twin == nil || !twin.ReachesDBMS(r.targetDBMS) {
			violations = append(violations, changeSetViolation(r, cs, fmt.Sprintf(
				"changeset %s runs on %s only and has no %s twin. Add changeset %s%s with dbms:%s to this file, or suppress dbms-twin with a reason.",
				ident, r.sourceDBMS, r.targetDBMS, cs.ID, r.idSuffix, r.targetDBMS)))
		}
	}

	return violations
}

// inScope reports whether the file is checked: not excluded and, when a threshold is
// configured, matching since_path_pattern with a number at least since.
func (r *DBMSTwinRule) inScope(filePath string) bool {
	if matchesAnyPattern(filePath, r.excludePatterns) {
		return false
	}
	if r.sincePath == nil {
		return true
	}
	match := r.sincePath.FindStringSubmatch(filepath.ToSlash(filePath))
	if match == nil {
		return false
	}
	number, err := strconv.Atoi(match[1])
	return err == nil && number >= r.since
}

// namesSource reports whether the changeset lists the source engine explicitly in its dbms.
func (r *DBMSTwinRule) namesSource(cs *parser.ChangeSet) bool {
	for _, value := range cs.DBMSList {
		if strings.ToLower(strings.TrimSpace(value)) == r.sourceDBMS {
			return true
		}
	}
	return false
}

// targetsOnly reports whether the changeset declares exactly the target engine.
func (r *DBMSTwinRule) targetsOnly(cs *parser.ChangeSet) bool {
	return len(cs.DBMSList) == 1 && strings.ToLower(strings.TrimSpace(cs.DBMSList[0])) == r.targetDBMS
}
