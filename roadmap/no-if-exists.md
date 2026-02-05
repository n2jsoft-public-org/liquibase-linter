# No IF EXISTS Rule Implementation

## Objective

Implement a rule that detects database-specific conditional existence checks (IF EXISTS, IF NOT EXISTS, IF OBJECT_ID) in SQL scripts and recommends using Liquibase preconditions instead.

## Requirements

### Functional Requirements

1. **Detection**: Identify IF EXISTS/IF NOT EXISTS patterns in SQL scripts
2. **Multi-Dialect Support**: Detect patterns from SQL Server, PostgreSQL, MySQL, Oracle
3. **SQL Preprocessing**: Preprocess SQL to remove comments and string literals before pattern matching
4. **Clear Messages**: Provide helpful violation messages with specific recommendations
5. **Configuration**: Support enabling/disabling and configurable severity
6. **Exclude Patterns**: Allow excluding specific files from this check

### Non-Functional Requirements

1. **Performance**: Efficient regex-based detection
2. **Accuracy**: Minimize false positives from comments and string literals
3. **Maintainability**: Clear code structure with well-documented patterns
4. **Testing**: Comprehensive test coverage including edge cases

## Implementation Details

### Rule Structure

**File**: `internal/rules/bestpractice.go`

Add the new rule to the existing best practices file:

```go
// NoIfExistsRule detects database-specific IF EXISTS patterns in SQL scripts
// and recommends using Liquibase preconditions instead.
type NoIfExistsRule struct {
	config config.RuleConfig
}

func NewNoIfExistsRule(cfg config.RuleConfig) *NoIfExistsRule {
	return &NoIfExistsRule{config: cfg}
}

func (r *NoIfExistsRule) ID() string {
	return "no-if-exists"
}

func (r *NoIfExistsRule) Name() string {
	return "No IF EXISTS"
}

func (r *NoIfExistsRule) Description() string {
	return "Detects database-specific IF EXISTS patterns and recommends Liquibase preconditions"
}

func (r *NoIfExistsRule) Severity() config.Severity {
	return r.config.Severity
}

func (r *NoIfExistsRule) Check(changelog *parser.Changelog) []config.Violation {
	if !r.config.Enabled {
		return nil
	}

	var violations []config.Violation

	for _, cs := range changelog.ChangeSets {
		for _, change := range cs.Changes {
			// Only check SQL changes
			if change.SQL == "" {
				continue
			}

			// Preprocess SQL to remove comments and string literals
			cleanedSQL := preprocessSQLForPatternMatching(change.SQL)

			// Check for IF EXISTS patterns
			if violation := r.checkForIfExists(cleanedSQL, changelog, cs); violation != nil {
				violations = append(violations, *violation)
			}
		}
	}

	return violations
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
	regexp.MustCompile(`(?i)\b(DROP|CREATE)\s+(PROCEDURE|FUNCTION|TABLE|INDEX)\s+IF\s+(NOT\s+)?EXISTS\b`),
	
	// Generic: IF [NOT] EXISTS in procedural blocks
	regexp.MustCompile(`(?i)\bBEGIN\s+IF\s+(NOT\s+)?EXISTS\s*\(`),
}

func (r *NoIfExistsRule) checkForIfExists(sql string, changelog *parser.Changelog, cs parser.ChangeSet) *config.Violation {
	for _, pattern := range ifExistsPatterns {
		if pattern.MatchString(sql) {
			// Extract a snippet of the matching pattern for the message
			match := pattern.FindString(sql)
			
			return &config.Violation{
				Rule:        r.ID(),
				Severity:    r.Severity(),
				Message:     fmt.Sprintf("Use Liquibase preconditions instead of database-specific IF EXISTS patterns (found: %s). Liquibase preconditions are cross-database compatible and integrate better with change tracking.", match),
				FilePath:    changelog.FilePath,
				ChangeSetID: cs.ID,
				Suggestion:  "Replace IF EXISTS with Liquibase preconditions like <preConditions onFail=\"MARK_RAN\"><not><tableExists tableName=\"...\"/></not></preConditions>",
			}
		}
	}
	return nil
}

// preprocessSQLForPatternMatching removes comments and string literals to avoid false positives
func preprocessSQLForPatternMatching(sql string) string {
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
```

### Registration

**File**: `internal/rules/registry.go`

Update the rule registration in the `GetDefaultRules()` function:

```go
func GetDefaultRules(config *config.Config) []Rule {
	rules := []Rule{
		// ... existing rules ...
		NewNamingConventionsRule(getRuleConfig(config, "naming-conventions")),
		NewNoIfExistsRule(getRuleConfig(config, "no-if-exists")),
		// ... other rules ...
	}
	return rules
}
```

## Configuration

### Default Configuration

Add to `internal/config/config.go` default rules:

```go
func DefaultConfig() *Config {
	return &Config{
		Rules: map[string]RuleConfig{
			// ... existing rules ...
			"no-if-exists": {
				Enabled:  true,
				Severity: SeverityWarning,
			},
			// ... other rules ...
		},
	}
}
```

### YAML Configuration Schema

Users can configure this rule in their `.liquibase-linter.yaml`:

```yaml
rules:
  no-if-exists:
    enabled: true
    severity: warning  # info, warning, error, critical
    exclude-patterns:
      - "**/init/**"
      - "**/migration-legacy/**"
```

## Testing

### Unit Tests

**File**: `internal/rules/bestpractice_test.go`

Add comprehensive tests:

```go
func TestNoIfExistsRule_Check(t *testing.T) {
	tests := []struct {
		name           string
		changelog      *parser.Changelog
		wantViolations int
		wantMessage    string
	}{
		{
			name: "SQL Server IF EXISTS pattern",
			changelog: &parser.Changelog{
				FilePath: "test.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "test-1",
						Author: "test",
						Changes: []parser.Change{
							{
								SQL: `IF EXISTS (SELECT * FROM sys.tables WHERE name = 'users')
								      BEGIN
								          DROP TABLE users;
								      END`,
							},
						},
					},
				},
			},
			wantViolations: 1,
			wantMessage:    "IF EXISTS",
		},
		{
			name: "SQL Server IF NOT EXISTS pattern",
			changelog: &parser.Changelog{
				FilePath: "test.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "test-2",
						Author: "test",
						Changes: []parser.Change{
							{
								SQL: `IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
								                     WHERE TABLE_NAME = 'users' AND COLUMN_NAME = 'email')
								      BEGIN
								          ALTER TABLE users ADD email VARCHAR(255);
								      END`,
							},
						},
					},
				},
			},
			wantViolations: 1,
			wantMessage:    "IF NOT EXISTS",
		},
		{
			name: "SQL Server IF OBJECT_ID pattern",
			changelog: &parser.Changelog{
				FilePath: "test.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "test-3",
						Author: "test",
						Changes: []parser.Change{
							{
								SQL: `IF OBJECT_ID('dbo.users', 'U') IS NOT NULL
								      DROP TABLE dbo.users;`,
							},
						},
					},
				},
			},
			wantViolations: 1,
			wantMessage:    "IF OBJECT_ID",
		},
		{
			name: "PostgreSQL DO block with IF EXISTS",
			changelog: &parser.Changelog{
				FilePath: "test.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "test-4",
						Author: "test",
						Changes: []parser.Change{
							{
								SQL: `DO $$
								      BEGIN
								          IF NOT EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'users') THEN
								              CREATE TABLE users (id SERIAL PRIMARY KEY);
								          END IF;
								      END $$;`,
							},
						},
					},
				},
			},
			wantViolations: 1,
			wantMessage:    "IF NOT EXISTS",
		},
		{
			name: "MySQL DROP PROCEDURE IF EXISTS",
			changelog: &parser.Changelog{
				FilePath: "test.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "test-5",
						Author: "test",
						Changes: []parser.Change{
							{
								SQL: `DROP PROCEDURE IF EXISTS create_index_if_not_exists;
								      DELIMITER //
								      CREATE PROCEDURE create_index_if_not_exists()
								      BEGIN
								          -- procedure body
								      END//`,
							},
						},
					},
				},
			},
			wantViolations: 1,
			wantMessage:    "DROP PROCEDURE IF EXISTS",
		},
		{
			name: "IF EXISTS in comment should not trigger",
			changelog: &parser.Changelog{
				FilePath: "test.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "test-6",
						Author: "test",
						Changes: []parser.Change{
							{
								SQL: `-- This comment mentions IF EXISTS but shouldn't trigger
								      CREATE TABLE users (id INT PRIMARY KEY);`,
							},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "IF EXISTS in string literal should not trigger",
			changelog: &parser.Changelog{
				FilePath: "test.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "test-7",
						Author: "test",
						Changes: []parser.Change{
							{
								SQL: `INSERT INTO logs (message) VALUES ('IF EXISTS should not trigger');`,
							},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "Clean SQL without IF EXISTS",
			changelog: &parser.Changelog{
				FilePath: "test.sql",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "test-8",
						Author: "test",
						Changes: []parser.Change{
							{
								SQL: `ALTER TABLE users ADD COLUMN email VARCHAR(255);`,
							},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "Non-SQL change should be ignored",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "test-9",
						Author: "test",
						Changes: []parser.Change{
							{
								AddColumn: &parser.AddColumn{
									TableName: "users",
									Columns: []parser.Column{
										{Name: "email", Type: "varchar(255)"},
									},
								},
							},
						},
					},
				},
			},
			wantViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewNoIfExistsRule(config.RuleConfig{
				Enabled:  true,
				Severity: config.SeverityWarning,
			})

			violations := rule.Check(tt.changelog)

			if len(violations) != tt.wantViolations {
				t.Errorf("Check() violations = %d, want %d", len(violations), tt.wantViolations)
			}

			if tt.wantViolations > 0 && len(violations) > 0 {
				if !contains(violations[0].Message, tt.wantMessage) {
					t.Errorf("Check() message = %s, want to contain %s", violations[0].Message, tt.wantMessage)
				}
			}
		})
	}
}

func TestNoIfExistsRule_Disabled(t *testing.T) {
	rule := NewNoIfExistsRule(config.RuleConfig{
		Enabled:  false,
		Severity: config.SeverityWarning,
	})

	changelog := &parser.Changelog{
		FilePath: "test.sql",
		ChangeSets: []parser.ChangeSet{
			{
				ID:     "test",
				Author: "test",
				Changes: []parser.Change{
					{
						SQL: "IF EXISTS (SELECT * FROM sys.tables) DROP TABLE users;",
					},
				},
			},
		},
	}

	violations := rule.Check(changelog)

	if len(violations) != 0 {
		t.Errorf("Check() should return no violations when disabled, got %d", len(violations))
	}
}

func TestPreprocessSQLForPatternMatching(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Remove single-line comment",
			input: "SELECT * FROM users -- IF EXISTS comment",
			want:  "SELECT * FROM users ",
		},
		{
			name:  "Remove multi-line comment",
			input: "SELECT * /* IF EXISTS in comment */ FROM users",
			want:  "SELECT *  FROM users",
		},
		{
			name:  "Remove string literal",
			input: "INSERT INTO logs VALUES ('IF EXISTS in string')",
			want:  "INSERT INTO logs VALUES ('')",
		},
		{
			name:  "Complex SQL with multiple elements",
			input: `-- Comment with IF EXISTS
			        SELECT * FROM users WHERE name = 'IF EXISTS';
			        /* Multi-line
			           IF NOT EXISTS
			           comment */`,
			want:  "\n\t\t        SELECT * FROM users WHERE name = '';\n\t\t        ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := preprocessSQLForPatternMatching(tt.input)
			if got != tt.want {
				t.Errorf("preprocessSQLForPatternMatching() = %q, want %q", got, tt.want)
			}
		})
	}
}
```

### Test Fixtures

Create test files in `testdata/` directory:

**File**: `testdata/if-exists-violations.sql`
```sql
--liquibase formatted sql

--changeset test:sql-server-if-exists
IF EXISTS (SELECT * FROM sys.tables WHERE name = 'temp_table')
BEGIN
    DROP TABLE temp_table;
END;

--changeset test:sql-server-if-not-exists
IF NOT EXISTS (SELECT * FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'users' AND COLUMN_NAME = 'email')
BEGIN
    ALTER TABLE users ADD email VARCHAR(255);
END;

--changeset test:postgresql-do-block
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'audit_log') THEN
        CREATE TABLE audit_log (id SERIAL PRIMARY KEY);
    END IF;
END $$;

--changeset test:mysql-drop-if-exists
DROP PROCEDURE IF EXISTS temp_procedure;
```

**File**: `testdata/if-exists-clean.sql`
```sql
--liquibase formatted sql

--changeset test:clean-change-with-precondition
--preconditions onFail:MARK_RAN
--precondition-not-table-exists tableName:users
CREATE TABLE users (
    id BIGINT PRIMARY KEY,
    name VARCHAR(100)
);

--changeset test:another-clean-change
ALTER TABLE users ADD COLUMN email VARCHAR(255);
```

## Integration Steps

1. **Add rule implementation** to `internal/rules/bestpractice.go`
2. **Register rule** in `internal/rules/registry.go`
3. **Add default configuration** in `internal/config/config.go`
4. **Add tests** to `internal/rules/bestpractice_test.go`
5. **Create test fixtures** in `testdata/`
6. **Run tests**: `go test ./internal/rules/... -v`
7. **Test with real changelogs**: `./liquibase-linter check testdata/if-exists-violations.sql`
8. **Update documentation**: Ensure `docs/rules/no-if-exists.md` is complete
9. **Update README.md**: Add rule to the list of implemented rules

## Validation Checklist

- [ ] Rule detects SQL Server IF EXISTS patterns
- [ ] Rule detects SQL Server IF NOT EXISTS patterns
- [ ] Rule detects SQL Server IF OBJECT_ID patterns
- [ ] Rule detects PostgreSQL DO blocks with IF EXISTS
- [ ] Rule detects MySQL DROP/CREATE IF EXISTS
- [ ] Rule ignores IF EXISTS in comments
- [ ] Rule ignores IF EXISTS in string literals
- [ ] Rule ignores non-SQL changes (XML/YAML change types)
- [ ] Rule can be disabled via configuration
- [ ] Rule severity can be configured
- [ ] All tests pass: `go test ./internal/rules/... -v`
- [ ] Test coverage > 80% for new code
- [ ] Documentation is complete in `docs/rules/no-if-exists.md`
- [ ] Example fixtures are in `testdata/`
- [ ] CLI help text includes the new rule: `./liquibase-linter rules`

## Expected Output

When running the linter on a changelog with IF EXISTS patterns:

```
$ ./liquibase-linter check testdata/if-exists-violations.sql

Checking: testdata/if-exists-violations.sql
✗ testdata/if-exists-violations.sql

WARNING: [no-if-exists] Use Liquibase preconditions instead of database-specific IF EXISTS patterns (found: IF EXISTS). Liquibase preconditions are cross-database compatible and integrate better with change tracking.
  ChangeSet: test:sql-server-if-exists
  File: testdata/if-exists-violations.sql
  Suggestion: Replace IF EXISTS with Liquibase preconditions like <preConditions onFail="MARK_RAN"><not><tableExists tableName="..."/></not></preConditions>

WARNING: [no-if-exists] Use Liquibase preconditions instead of database-specific IF EXISTS patterns (found: IF NOT EXISTS). Liquibase preconditions are cross-database compatible and integrate better with change tracking.
  ChangeSet: test:sql-server-if-not-exists
  File: testdata/if-exists-violations.sql
  Suggestion: Replace IF EXISTS with Liquibase preconditions like <preConditions onFail="MARK_RAN"><not><tableExists tableName="..."/></not></preConditions>

Summary:
  Files checked: 1
  Changesets: 4
  Violations: 4 (0 critical, 0 errors, 4 warnings, 0 info)
```

## Dependencies

- **Parser**: No changes needed - SQL content already parsed
- **Config**: Add rule to default configuration
- **Reporter**: Works with existing reporters (text, JSON, SARIF)

## Estimated Effort

- **Implementation**: 2-3 hours
- **Testing**: 1 hour
- **Documentation**: 30 minutes
- **Total**: 3-4 hours

## Notes

- This rule is particularly useful for teams migrating from database-specific SQL to Liquibase
- The preprocessing function helps avoid false positives from comments and strings
- The rule only checks SQL changes (formatted SQL and `<sql>` tags), not XML/YAML change types
- Consider adding auto-fix capability in a future version to convert simple patterns to Liquibase preconditions
