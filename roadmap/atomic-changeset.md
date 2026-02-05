# Atomic Changeset Rule Implementation

## Objective

Implement a rule that enforces each changeset contains only a single change operation, promoting atomicity, cleaner rollbacks, and better change tracking.

## Requirements

### Functional Requirements

1. **Change Counting**: Count the number of distinct change operations in each changeset
2. **XML/YAML Detection**: Detect multiple change elements (createTable, createIndex, addColumn, etc.)
3. **SQL Statement Counting**: Count SQL statements in SQL-based changes, accounting for:
   - Multiple statements separated by semicolons
   - Different operation types (CREATE, ALTER, INSERT, etc.)
4. **Configurable Exceptions**: Support configuration options for:
   - Allowing table creation with inline constraints
   - Allowing table creation with indexes
   - Setting maximum SQL statement count
5. **Exclude Patterns**: Support glob patterns to exclude init scripts and legacy migrations
6. **Clear Violations**: Provide clear messages indicating which changes were found

### Non-Functional Requirements

1. **Performance**: Efficient O(n) scanning of changesets
2. **Accuracy**: Correctly identify SQL statement boundaries
3. **Maintainability**: Clear separation between counting logic and violation detection
4. **Testing**: Comprehensive coverage of edge cases

## Implementation Details

### Rule Structure

**File**: `internal/rules/bestpractice.go`

Add the new rule to the existing best practices file:

```go
// AtomicChangesetRule enforces one change per changeset.
type AtomicChangesetRule struct {
	config config.RuleConfig
}

func NewAtomicChangesetRule(cfg config.RuleConfig) *AtomicChangesetRule {
	return &AtomicChangesetRule{config: cfg}
}

func (r *AtomicChangesetRule) ID() string {
	return "atomic-changeset"
}

func (r *AtomicChangesetRule) Name() string {
	return "Atomic Changeset"
}

func (r *AtomicChangesetRule) Description() string {
	return "Enforces that each changeset contains only a single change operation"
}

func (r *AtomicChangesetRule) Severity() config.Severity {
	return r.config.Severity
}

func (r *AtomicChangesetRule) Check(changelog *parser.Changelog) []config.Violation {
	if !r.config.Enabled {
		return nil
	}

	var violations []config.Violation

	for _, cs := range changelog.ChangeSets {
		// Check if file should be excluded
		if r.shouldExcludeFile(cs.FilePath) {
			continue
		}

		// Count changes in this changeset
		changeCount := r.countChanges(cs)

		if changeCount > 1 {
			changeTypes := r.getChangeTypes(cs)
			violations = append(violations, config.Violation{
				Rule:        r.ID(),
				Severity:    r.Severity(),
				Message:     fmt.Sprintf("Changeset contains %d changes (%s). Consider splitting into separate changesets for better atomicity and rollback clarity.", changeCount, strings.Join(changeTypes, ", ")),
				FilePath:    cs.FilePath,
				ChangeSetID: cs.ID,
				Suggestion:  "Split this changeset into multiple changesets, each containing a single change operation",
			})
		}
	}

	return violations
}

// shouldExcludeFile checks if the file path matches any exclude patterns
func (r *AtomicChangesetRule) shouldExcludeFile(filePath string) bool {
	excludePatterns := r.config.GetStringSlice("exclude-patterns")
	if len(excludePatterns) == 0 {
		return false
	}

	for _, pattern := range excludePatterns {
		// Use filepath.Match or doublestar for glob matching
		matched, err := filepath.Match(pattern, filePath)
		if err == nil && matched {
			return true
		}

		// Also try matching with ** support
		if matchesGlobPattern(filePath, pattern) {
			return true
		}
	}

	return false
}

// countChanges counts the number of distinct changes in a changeset
func (r *AtomicChangesetRule) countChanges(cs parser.ChangeSet) int {
	count := 0
	allowTableWithConstraints := r.config.GetBool("allow-table-with-constraints", true)
	allowTableWithIndexes := r.config.GetBool("allow-table-with-indexes", false)
	maxSQLStatements := r.config.GetInt("max-sql-statements", 1)

	hasCreateTable := false
	indexCount := 0

	for _, change := range cs.Changes {
		changeType := change.GetChangeType()

		// Handle SQL changes specially - count statements
		if change.SQL != "" {
			sqlCount := r.countSQLStatements(change.SQL)
			if sqlCount > maxSQLStatements {
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
	if hasCreateTable && indexCount > 0 {
		if allowTableWithIndexes {
			// Treat table + indexes as one logical operation
			count = 1
		}
	}

	// If table has inline constraints, they're already part of createTable
	// and counted as 1, so no adjustment needed for allowTableWithConstraints

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
			// Check if it's a substantive statement (not just whitespace/comments)
			if len(stmt) > 0 {
				count++
			}
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
	// Remove single-line comments (-- and //)
	sql = regexp.MustCompile(`--[^\n]*`).ReplaceAllString(sql, "")
	sql = regexp.MustCompile(`//[^\n]*`).ReplaceAllString(sql, "")

	// Remove multi-line comments (/* ... */)
	sql = regexp.MustCompile(`(?s)/\*.*?\*/`).ReplaceAllString(sql, "")

	// Remove string literals (single quotes)
	sql = regexp.MustCompile(`'[^']*'`).ReplaceAllString(sql, "''")

	// Remove string literals (double quotes)
	sql = regexp.MustCompile(`"[^"]*"`).ReplaceAllString(sql, `""`)

	return sql
}

// getChangeTypes returns a list of change type names in the changeset
func (r *AtomicChangesetRule) getChangeTypes(cs parser.ChangeSet) []string {
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

// matchesGlobPattern checks if a path matches a glob pattern with ** support
func matchesGlobPattern(path, pattern string) bool {
	// Convert pattern to regex for ** support
	regexPattern := strings.ReplaceAll(pattern, "**", ".*")
	regexPattern = strings.ReplaceAll(regexPattern, "*", "[^/]*")
	regexPattern = "^" + regexPattern + "$"

	matched, err := regexp.MatchString(regexPattern, path)
	return err == nil && matched
}
```

### Configuration Schema

**File**: `internal/config/config.go`

Ensure RuleConfig supports nested configuration:

```go
// RuleConfig represents configuration for a single rule
type RuleConfig struct {
	Enabled  bool                   `yaml:"enabled"`
	Severity Severity               `yaml:"severity"`
	Options  map[string]interface{} `yaml:",inline"` // Allow additional fields
}

// Helper methods for type-safe config access
func (rc RuleConfig) GetBool(key string, defaultVal bool) bool {
	if val, ok := rc.Options[key].(bool); ok {
		return val
	}
	return defaultVal
}

func (rc RuleConfig) GetInt(key string, defaultVal int) int {
	if val, ok := rc.Options[key].(int); ok {
		return val
	}
	return defaultVal
}

func (rc RuleConfig) GetStringSlice(key string) []string {
	if val, ok := rc.Options[key].([]interface{}); ok {
		result := make([]string, len(val))
		for i, v := range val {
			if s, ok := v.(string); ok {
				result[i] = s
			}
		}
		return result
	}
	return nil
}
```

Update default configuration:

```go
func DefaultConfig() *Config {
	return &Config{
		Rules: map[string]RuleConfig{
			// ... existing rules ...
			"atomic-changeset": {
				Enabled:  true,
				Severity: SeverityInfo,
				Options: map[string]interface{}{
					"allow-table-with-indexes":     false,
					"allow-table-with-constraints": true,
					"max-sql-statements":           1,
					"exclude-patterns":             []string{},
				},
			},
			// ... other rules ...
		},
	}
}
```

### Parser Enhancement

**File**: `internal/parser/types.go`

Ensure Change struct has GetChangeType() method:

```go
// GetChangeType returns a string representation of the change type
func (c *Change) GetChangeType() string {
	if c.CreateTable != nil {
		return "createTable"
	}
	if c.CreateIndex != nil {
		return "createIndex"
	}
	if c.AddColumn != nil {
		return "addColumn"
	}
	if c.AddForeignKey != nil {
		return "addForeignKey"
	}
	if c.AddPrimaryKey != nil {
		return "addPrimaryKey"
	}
	if c.AddNotNullConstraint != nil {
		return "addNotNullConstraint"
	}
	if c.AddUniqueConstraint != nil {
		return "addUniqueConstraint"
	}
	if c.DropTable != nil {
		return "dropTable"
	}
	if c.DropColumn != nil {
		return "dropColumn"
	}
	if c.DropIndex != nil {
		return "dropIndex"
	}
	if c.Insert != nil {
		return "insert"
	}
	if c.Update != nil {
		return "update"
	}
	if c.Delete != nil {
		return "delete"
	}
	if c.SQL != "" {
		return "sql"
	}
	return "unknown"
}
```

### Registration

**File**: `internal/rules/registry.go`

Register the new rule:

```go
func GetDefaultRules(config *config.Config) []Rule {
	rules := []Rule{
		// ... existing rules ...
		NewAtomicChangesetRule(getRuleConfig(config, "atomic-changeset")),
		// ... other rules ...
	}
	return rules
}
```

## Testing

### Unit Tests

**File**: `internal/rules/bestpractice_test.go`

```go
func TestAtomicChangesetRule_Check(t *testing.T) {
	tests := []struct {
		name           string
		changelog      *parser.Changelog
		config         config.RuleConfig
		wantViolations int
	}{
		{
			name: "Single change - no violation",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{CreateTable: &parser.CreateTable{TableName: "users"}},
						},
					},
				},
			},
			config: config.RuleConfig{
				Enabled:  true,
				Severity: config.SeverityInfo,
			},
			wantViolations: 0,
		},
		{
			name: "Multiple changes - violation",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{CreateTable: &parser.CreateTable{TableName: "users"}},
							{CreateIndex: &parser.CreateIndex{TableName: "users", IndexName: "idx_email"}},
							{AddColumn: &parser.AddColumn{TableName: "users"}},
						},
					},
				},
			},
			config: config.RuleConfig{
				Enabled:  true,
				Severity: config.SeverityInfo,
			},
			wantViolations: 1,
		},
		{
			name: "Multiple SQL statements - violation",
			changelog: &parser.Changelog{
				FilePath: "test.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{
								SQL: `CREATE TABLE users (id INT);
								      CREATE INDEX idx_email ON users(email);`,
							},
						},
					},
				},
			},
			config: config.RuleConfig{
				Enabled:  true,
				Severity: config.SeverityInfo,
			},
			wantViolations: 1,
		},
		{
			name: "Table with index - allowed",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{CreateTable: &parser.CreateTable{TableName: "users"}},
							{CreateIndex: &parser.CreateIndex{TableName: "users"}},
						},
					},
				},
			},
			config: config.RuleConfig{
				Enabled:  true,
				Severity: config.SeverityInfo,
				Options: map[string]interface{}{
					"allow-table-with-indexes": true,
				},
			},
			wantViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewAtomicChangesetRule(tt.config)
			violations := rule.Check(tt.changelog)

			if len(violations) != tt.wantViolations {
				t.Errorf("Check() violations = %d, want %d", len(violations), tt.wantViolations)
			}
		})
	}
}
```

## Integration Steps

1. **Add rule implementation** to `internal/rules/bestpractice.go`
2. **Update configuration schema** in `internal/config/config.go`
3. **Add parser helper** (GetChangeType) to `internal/parser/types.go` if needed
4. **Register rule** in `internal/rules/registry.go`
5. **Add tests** to `internal/rules/bestpractice_test.go`
6. **Run tests**: `go test ./internal/rules/... -v`
7. **Test with real changelogs**
8. **Update documentation**: Ensure `docs/rules/atomic-changeset.md` is complete
9. **Update README.md**: Add rule to the list

## Validation Checklist

- [ ] Rule detects multiple XML/YAML changes in one changeset
- [ ] Rule detects multiple SQL statements
- [ ] Rule correctly preprocesses SQL to ignore semicolons in comments/strings
- [ ] Rule respects `allow-table-with-indexes` configuration
- [ ] Rule respects `allow-table-with-constraints` configuration
- [ ] Rule applies exclude patterns correctly
- [ ] Rule can be disabled via configuration
- [ ] Violation messages are clear and actionable
- [ ] All tests pass
- [ ] Documentation is complete

## Expected Output

```
$ ./liquibase-linter check changelog/sprints/v123/structure/tables.xml

INFO: [atomic-changeset] Changeset contains 3 changes (createTable, createIndex, addColumn). Consider splitting into separate changesets for better atomicity and rollback clarity.
  ChangeSet: john.doe:setup-users-table
  File: changelog/sprints/v123/structure/tables.xml
  Suggestion: Split this changeset into multiple changesets, each containing a single change operation
```

## Dependencies

- **Parser**: Requires `GetChangeType()` helper method on Change struct
- **Config**: Requires support for nested configuration options

## Estimated Effort

- **Implementation**: 3-4 hours
- **Testing**: 2 hours
- **Documentation**: 1 hour (complete)
- **Integration**: 1 hour
- **Total**: 7-8 hours

## Notes

- Default severity is `info` since this is a best practice guideline
- SQL statement counting handles edge cases like semicolons in strings/comments
- Complements the `non-idempotent` rule by promoting cleaner changesets
