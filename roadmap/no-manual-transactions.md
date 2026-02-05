# Task: Implement No Manual Transactions Rule

**Status**: 📋 Planned  
**Priority**: Medium  
**Estimated Effort**: 4-5 hours  
**Category**: Reliability Rule  
**Target File**: `internal/rules/reliability.go` (NEW FILE)

---

## Objective

Implement a linting rule that detects manual transaction control statements (BEGIN, COMMIT, ROLLBACK) that interfere with Liquibase's automatic transaction management.

---

## Requirements

### Functional Requirements
- Detect transaction control keywords in SQL
- Remove false positives (keywords in comments, strings, stored procedures)
- Support multiple SQL dialects (SQL Server, PostgreSQL, MySQL)
- Allow excluding specific change types (procedures, functions, triggers)
- Work across all changelog formats (XML, YAML, SQL)

### Non-Functional Requirements
- Performance: O(n*m) where n = changesets, m = SQL length
- Accurate pattern matching with minimal false positives
- Clear error messages indicating which keyword was detected

---

## Implementation Details

### 1. Create New File

Create `internal/rules/reliability.go` with package header:

```go
// Package rules provides reliability linting rules.
package rules

import (
    "fmt"
    "regexp"
    "strings"

    "github.com/n2jsoft/liquibase-linter/internal/config"
    "github.com/n2jsoft/liquibase-linter/internal/parser"
)
```

### 2. Configuration Structure

Add to `internal/config/config.go`:

```go
// NoManualTransactionsConfig represents configuration for no-manual-transactions rule.
type NoManualTransactionsConfig struct {
    Enabled            bool     `yaml:"enabled"`
    Patterns           []string `yaml:"patterns"`
    CaseInsensitive    bool     `yaml:"case_insensitive"`
    ExcludeChangeTypes []string `yaml:"exclude_change_types"`
    ExcludePatterns    []string `yaml:"exclude_patterns"`
}
```

Update `Config` struct:
```go
type Config struct {
    // ... existing fields ...
    NoManualTransactions NoManualTransactionsConfig `yaml:"no_manual_transactions"`
}
```

### 3. Type Definition

In `internal/rules/reliability.go`:

```go
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
                `\bROLLBACK\s+(TRANSACTION|TRAN|WORK)?\b`,
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
```

### 4. Rule Interface Implementation

```go
func (r *NoManualTransactionsRule) ID() string {
    return "no-manual-transactions"
}

func (r *NoManualTransactionsRule) Name() string {
    return "No Manual Transaction Control"
}

func (r *NoManualTransactionsRule) Description() string {
    return "Detects manual transaction control statements (BEGIN, COMMIT, ROLLBACK) that interfere with Liquibase's transaction management"
}

func (r *NoManualTransactionsRule) Severity() Severity {
    return SeverityWarning
}
```

### 5. Check Method

```go
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
```

### 6. SQL Preprocessing (Critical for Accuracy)

```go
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
```

### 7. Helper Methods

```go
func (r *NoManualTransactionsRule) shouldExcludeFile(filePath string) bool {
    if len(r.excludePatterns) == 0 {
        return false
    }

    for _, pattern := range r.excludePatterns {
        matched, err := filepath.Match(pattern, filePath)
        if err == nil && matched {
            return true
        }
        
        matched, err = filepath.Match(pattern, filepath.Base(filePath))
        if err == nil && matched {
            return true
        }
    }
    
    return false
}
```

---

## Configuration

### Update `internal/config/config.go`

Add to the `Default()` function:

```go
"no-manual-transactions": {
    Enabled:  true,
    Severity: "warning",
},
NoManualTransactions: NoManualTransactionsConfig{
    Patterns: []string{
        `\bBEGIN\s+(TRANSACTION|TRAN|WORK)?\b`,
        `\bSTART\s+TRANSACTION\b`,
        `\bCOMMIT\s+(TRANSACTION|TRAN|WORK)?\b`,
        `\bROLLBACK\s+(TRANSACTION|TRAN|WORK)?\b`,
        `\bSAVE(POINT)?\s+TRANSACTION\b`,
    },
    CaseInsensitive: true,
    ExcludeChangeTypes: []string{
        "createProcedure",
        "createFunction",
        "createTrigger",
    },
},
```

### Configuration Schema

```yaml
rules:
  no-manual-transactions:
    enabled: true
    severity: warning

no_manual_transactions:
  case_insensitive: true
  exclude_change_types:
    - 'createProcedure'
    - 'createFunction'
    - 'createTrigger'
  patterns:
    - '\bBEGIN\s+(TRANSACTION|TRAN|WORK)?\b'
    - '\bSTART\s+TRANSACTION\b'
    - '\bCOMMIT\s+(TRANSACTION|TRAN|WORK)?\b'
    - '\bROLLBACK\s+(TRANSACTION|TRAN|WORK)?\b'
```

---

## Testing

### Test File Location
`internal/rules/reliability_test.go` (NEW FILE)

### Test Cases

```go
package rules

import (
    "testing"

    "github.com/n2jsoft/liquibase-linter/internal/config"
    "github.com/n2jsoft/liquibase-linter/internal/parser"
)

func TestNoManualTransactionsRule_Check(t *testing.T) {
    rule := NewNoManualTransactionsRule(nil)

    tests := []struct {
        name          string
        changelog     *parser.Changelog
        wantViolation bool
        expectedMsg   string
    }{
        {
            name: "BEGIN TRANSACTION detected",
            changelog: &parser.Changelog{
                ChangeSets: []parser.ChangeSet{
                    {
                        ID:       "1",
                        Author:   "test",
                        FilePath: "/test/changelog.sql",
                        Changes: []parser.Change{
                            {
                                Type: "sql",
                                SQL:  "BEGIN TRANSACTION;\nINSERT INTO users VALUES (1);\nCOMMIT;",
                            },
                        },
                    },
                },
            },
            wantViolation: true,
            expectedMsg:   "BEGIN TRANSACTION",
        },
        {
            name: "START TRANSACTION detected",
            changelog: &parser.Changelog{
                ChangeSets: []parser.ChangeSet{
                    {
                        ID:       "2",
                        Author:   "test",
                        FilePath: "/test/changelog.sql",
                        Changes: []parser.Change{
                            {
                                Type: "sql",
                                SQL:  "START TRANSACTION;\nUPDATE users SET active = 1;",
                            },
                        },
                    },
                },
            },
            wantViolation: true,
            expectedMsg:   "START TRANSACTION",
        },
        {
            name: "transaction keyword in comment - should not detect",
            changelog: &parser.Changelog{
                ChangeSets: []parser.ChangeSet{
                    {
                        ID:       "3",
                        Author:   "test",
                        FilePath: "/test/changelog.sql",
                        Changes: []parser.Change{
                            {
                                Type: "sql",
                                SQL:  "-- BEGIN TRANSACTION for testing\nINSERT INTO users VALUES (1);",
                            },
                        },
                    },
                },
            },
            wantViolation: false,
        },
        {
            name: "transaction keyword in string - should not detect",
            changelog: &parser.Changelog{
                ChangeSets: []parser.ChangeSet{
                    {
                        ID:       "4",
                        Author:   "test",
                        FilePath: "/test/changelog.sql",
                        Changes: []parser.Change{
                            {
                                Type: "sql",
                                SQL:  "INSERT INTO logs (message) VALUES ('BEGIN TRANSACTION complete');",
                            },
                        },
                    },
                },
            },
            wantViolation: false,
        },
        {
            name: "stored procedure with transaction - excluded",
            changelog: &parser.Changelog{
                ChangeSets: []parser.ChangeSet{
                    {
                        ID:       "5",
                        Author:   "test",
                        FilePath: "/test/changelog.sql",
                        Changes: []parser.Change{
                            {
                                Type: "createProcedure",
                                SQL:  "CREATE PROCEDURE test AS BEGIN BEGIN TRANSACTION; UPDATE users; COMMIT; END;",
                            },
                        },
                    },
                },
            },
            wantViolation: false,
        },
        {
            name: "COMMIT detected",
            changelog: &parser.Changelog{
                ChangeSets: []parser.ChangeSet{
                    {
                        ID:       "6",
                        Author:   "test",
                        FilePath: "/test/changelog.sql",
                        Changes: []parser.Change{
                            {
                                Type: "sql",
                                SQL:  "UPDATE users SET active = 1;\nCOMMIT TRANSACTION;",
                            },
                        },
                    },
                },
            },
            wantViolation: true,
            expectedMsg:   "COMMIT",
        },
        {
            name: "no transaction keywords",
            changelog: &parser.Changelog{
                ChangeSets: []parser.ChangeSet{
                    {
                        ID:       "7",
                        Author:   "test",
                        FilePath: "/test/changelog.sql",
                        Changes: []parser.Change{
                            {
                                Type: "sql",
                                SQL:  "INSERT INTO users (name, email) VALUES ('John', 'john@example.com');",
                            },
                        },
                    },
                },
            },
            wantViolation: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            violations := rule.Check(tt.changelog)
            
            if tt.wantViolation && len(violations) == 0 {
                t.Error("Expected violation but got none")
            }
            if !tt.wantViolation && len(violations) > 0 {
                t.Errorf("Expected no violations but got %d: %v", len(violations), violations)
            }
            if tt.wantViolation && tt.expectedMsg != "" && len(violations) > 0 {
                if !strings.Contains(violations[0].Message, tt.expectedMsg) {
                    t.Errorf("Expected message to contain '%s', got '%s'", tt.expectedMsg, violations[0].Message)
                }
            }
        })
    }
}

func TestNoManualTransactionsRule_PreprocessSQL(t *testing.T) {
    rule := NewNoManualTransactionsRule(nil)

    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {
            name:     "remove single-line comment",
            input:    "SELECT * FROM users; -- BEGIN TRANSACTION",
            expected: "SELECT * FROM users; ",
        },
        {
            name:     "remove multi-line comment",
            input:    "SELECT * /* BEGIN TRANSACTION */ FROM users;",
            expected: "SELECT *  FROM users;",
        },
        {
            name:     "remove single-quoted string",
            input:    "INSERT INTO logs VALUES ('BEGIN TRANSACTION');",
            expected: "INSERT INTO logs VALUES ();",
        },
        {
            name:     "remove double-quoted string",
            input:    "INSERT INTO logs VALUES (\"START TRANSACTION\");",
            expected: "INSERT INTO logs VALUES ();",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := rule.preprocessSQL(tt.input)
            // Normalize whitespace for comparison
            result = strings.TrimSpace(result)
            expected := strings.TrimSpace(tt.expected)
            if result != expected {
                t.Errorf("Expected '%s', got '%s'", expected, result)
            }
        })
    }
}
```

### Test Fixtures

Create `testdata/no-manual-transactions-test.sql`:

```sql
--liquibase formatted sql

-- Valid: No transaction control
--changeset test:1
INSERT INTO users (name, email) VALUES ('John', 'john@example.com');

-- Invalid: BEGIN TRANSACTION
--changeset test:2
BEGIN TRANSACTION;
UPDATE users SET active = 1;
COMMIT;

-- Valid: Transaction keyword in comment
--changeset test:3
-- This will run in a transaction managed by Liquibase
-- Don't manually BEGIN TRANSACTION
UPDATE users SET verified = 1;

-- Valid: Transaction keyword in string
--changeset test:4
INSERT INTO audit_log (message) VALUES ('Transaction BEGIN detected');

-- Valid: Stored procedure with transactions (if excluded)
--changeset test:5
CREATE PROCEDURE UpdateUsers
AS
BEGIN
    BEGIN TRANSACTION;
    UPDATE users SET status = 'active';
    COMMIT TRANSACTION;
END;
```

---

## Integration Steps

### 1. Register Rule in Main

Update `cmd/liquibase-linter/main.go`:

```go
// Load no-manual-transactions config
noManualTransactionsCfg := &cfg.NoManualTransactions
if !cfg.Rules["no-manual-transactions"].Enabled {
    noManualTransactionsCfg = nil
}

// Register rules
allRules := []rules.Rule{
    // ... existing rules ...
    rules.NewNoManualTransactionsRule(noManualTransactionsCfg),
}
```

### 2. Update Documentation

File: `docs/rules.md`

Add entry:
```markdown
| no-manual-transactions | Warning | Reliability | Detects manual transaction control interfering with Liquibase |
```

---

## Validation Checklist

- [ ] New file created: `internal/rules/reliability.go`
- [ ] Configuration structure added to `config.go`
- [ ] Rule implementation complete
- [ ] Tests created: `internal/rules/reliability_test.go`
- [ ] SQL preprocessing tests pass
- [ ] Test fixtures created in `testdata/`
- [ ] Configuration updated with defaults
- [ ] Rule registered in `main.go`
- [ ] Documentation file exists: `docs/rules/no-manual-transactions.md`
- [ ] Documentation updated in `docs/rules.md`
- [ ] All tests pass: `go test ./...`
- [ ] False positive tests pass (comments, strings, procedures)
- [ ] Multi-dialect support verified (SQL Server, PostgreSQL, MySQL)

---

## Expected Output Example

```
WARNING: no-manual-transactions
File: changelog/sprints/v123/1 - data/migration.sql
Changeset: john.doe:456
Message: Manual transaction control detected: 'BEGIN TRANSACTION' - let Liquibase manage transactions

WARNING: no-manual-transactions
File: changelog/sprints/v123/1 - data/update.xml
Changeset: jane.smith:789
Message: Manual transaction control detected: 'COMMIT' - let Liquibase manage transactions
```

---

## Dependencies

- Parser already supports `Change.SQL` ✅
- No parser modifications needed ✅
- Uses existing `Violation` structure ✅
- Requires `regexp` package (standard library) ✅

---

## Related Documentation

- [docs/rules/no-manual-transactions.md](../docs/rules/no-manual-transactions.md)
- [docs/rules/dangerous-operations.md](../docs/rules/dangerous-operations.md) (related)

---

## Notes

- SQL preprocessing is critical to avoid false positives
- Consider providing migration guide for fixing violations
- Test thoroughly with real-world stored procedures
- Different SQL dialects use different transaction syntax - patterns should cover all
