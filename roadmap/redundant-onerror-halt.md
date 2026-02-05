# Task: Implement Redundant onError:HALT Rule

**Status**: 📋 Planned  
**Priority**: Low  
**Estimated Effort**: 1-2 hours  
**Category**: Best Practices Rule  
**Target File**: `internal/rules/bestpractice.go`

---

## Objective

Implement a linting rule that detects redundant `onError:HALT` configuration in preconditions, since `HALT` is the default value and doesn't need to be explicitly specified.

---

## Requirements

### Functional Requirements
- Detect preconditions with `onError` set to `"HALT"`
- Only flag `HALT` - other values (`WARN`, `CONTINUE`, `MARK_RAN`) are valid non-defaults
- Work across all changelog formats (XML, YAML, SQL)
- Provide clear message explaining why it's redundant

### Non-Functional Requirements
- Rule should be enabled by default (cosmetic issue, non-breaking)
- Performance: O(n) where n is number of changesets with preconditions
- Informational severity (not a critical issue)

---

## Implementation Details

### 1. Type Definition

Add to `internal/rules/bestpractice.go`:

```go
// RedundantOnErrorHaltRule detects redundant onError:HALT in preconditions.
type RedundantOnErrorHaltRule struct{}

// NewRedundantOnErrorHaltRule creates a new redundant onError:HALT rule.
func NewRedundantOnErrorHaltRule() *RedundantOnErrorHaltRule {
    return &RedundantOnErrorHaltRule{}
}
```

### 2. Rule Interface Implementation

```go
func (r *RedundantOnErrorHaltRule) ID() string {
    return "redundant-onerror-halt"
}

func (r *RedundantOnErrorHaltRule) Name() string {
    return "Redundant onError:HALT Detection"
}

func (r *RedundantOnErrorHaltRule) Description() string {
    return "Detects redundant onError:HALT configuration - HALT is the default and doesn't need to be specified"
}

func (r *RedundantOnErrorHaltRule) Severity() Severity {
    return SeverityInfo
}
```

### 3. Check Method

```go
func (r *RedundantOnErrorHaltRule) Check(changelog *parser.Changelog) []Violation {
    violations := make([]Violation, 0)

    for _, cs := range changelog.ChangeSets {
        // Only check changesets that have preconditions
        if cs.Preconditions == nil {
            continue
        }

        // Check if onError is explicitly set to "HALT"
        if strings.ToUpper(cs.Preconditions.OnError) == "HALT" {
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
```

---

## Configuration

### Update `internal/config/config.go`

Add to the `Default()` function:

```go
"redundant-onerror-halt": {
    Enabled:  true,  // Enabled by default (cosmetic issue)
    Severity: "info",
},
```

### Configuration Schema

```yaml
rules:
  redundant-onerror-halt:
    enabled: true
    severity: info
```

---

## Testing

### Test File Location
`internal/rules/bestpractice_test.go`

### Test Cases

```go
func TestRedundantOnErrorHaltRule_Check(t *testing.T) {
    rule := NewRedundantOnErrorHaltRule()

    tests := []struct {
        name          string
        changelog     *parser.Changelog
        wantViolation bool
    }{
        {
            name: "redundant onError:HALT",
            changelog: &parser.Changelog{
                ChangeSets: []parser.ChangeSet{
                    {
                        ID:       "1",
                        Author:   "test",
                        FilePath: "/test/changelog.xml",
                        Preconditions: &parser.Precondition{
                            Type:    "tableExists",
                            OnFail:  "MARK_RAN",
                            OnError: "HALT",
                        },
                    },
                },
            },
            wantViolation: true,
        },
        {
            name: "onError:HALT lowercase",
            changelog: &parser.Changelog{
                ChangeSets: []parser.ChangeSet{
                    {
                        ID:       "2",
                        Author:   "test",
                        FilePath: "/test/changelog.xml",
                        Preconditions: &parser.Precondition{
                            Type:    "tableExists",
                            OnFail:  "MARK_RAN",
                            OnError: "halt",
                        },
                    },
                },
            },
            wantViolation: true,
        },
        {
            name: "onError:WARN - no violation",
            changelog: &parser.Changelog{
                ChangeSets: []parser.ChangeSet{
                    {
                        ID:       "3",
                        Author:   "test",
                        FilePath: "/test/changelog.xml",
                        Preconditions: &parser.Precondition{
                            Type:    "tableExists",
                            OnFail:  "MARK_RAN",
                            OnError: "WARN",
                        },
                    },
                },
            },
            wantViolation: false,
        },
        {
            name: "onError:CONTINUE - no violation",
            changelog: &parser.Changelog{
                ChangeSets: []parser.ChangeSet{
                    {
                        ID:       "4",
                        Author:   "test",
                        FilePath: "/test/changelog.xml",
                        Preconditions: &parser.Precondition{
                            Type:    "tableExists",
                            OnFail:  "MARK_RAN",
                            OnError: "CONTINUE",
                        },
                    },
                },
            },
            wantViolation: false,
        },
        {
            name: "onError:MARK_RAN - no violation",
            changelog: &parser.Changelog{
                ChangeSets: []parser.ChangeSet{
                    {
                        ID:       "5",
                        Author:   "test",
                        FilePath: "/test/changelog.xml",
                        Preconditions: &parser.Precondition{
                            Type:    "tableExists",
                            OnFail:  "MARK_RAN",
                            OnError: "MARK_RAN",
                        },
                    },
                },
            },
            wantViolation: false,
        },
        {
            name: "no onError attribute - no violation",
            changelog: &parser.Changelog{
                ChangeSets: []parser.ChangeSet{
                    {
                        ID:       "6",
                        Author:   "test",
                        FilePath: "/test/changelog.xml",
                        Preconditions: &parser.Precondition{
                            Type:    "tableExists",
                            OnFail:  "MARK_RAN",
                            OnError: "",
                        },
                    },
                },
            },
            wantViolation: false,
        },
        {
            name: "no preconditions - no violation",
            changelog: &parser.Changelog{
                ChangeSets: []parser.ChangeSet{
                    {
                        ID:            "7",
                        Author:        "test",
                        FilePath:      "/test/changelog.xml",
                        Preconditions: nil,
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
        })
    }
}
```

### Test Fixtures

Create `testdata/redundant-onerror-halt-test.xml`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<databaseChangeLog
    xmlns="http://www.liquibase.org/xml/ns/dbchangelog"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    xsi:schemaLocation="http://www.liquibase.org/xml/ns/dbchangelog
    http://www.liquibase.org/xml/ns/dbchangelog/dbchangelog-3.8.xsd">

    <!-- Invalid: Redundant onError="HALT" -->
    <changeSet id="1" author="test">
        <preConditions onFail="MARK_RAN" onError="HALT">
            <tableExists tableName="users"/>
        </preConditions>
        <addColumn tableName="users">
            <column name="email" type="VARCHAR(255)"/>
        </addColumn>
    </changeSet>

    <!-- Valid: No onError attribute (uses default) -->
    <changeSet id="2" author="test">
        <preConditions onFail="MARK_RAN">
            <tableExists tableName="users"/>
        </preConditions>
        <addColumn tableName="users">
            <column name="phone" type="VARCHAR(20)"/>
        </addColumn>
    </changeSet>

    <!-- Valid: Non-default onError value -->
    <changeSet id="3" author="test">
        <preConditions onFail="MARK_RAN" onError="WARN">
            <tableExists tableName="users"/>
        </preConditions>
        <addColumn tableName="users">
            <column name="status" type="VARCHAR(50)"/>
        </addColumn>
    </changeSet>

</databaseChangeLog>
```

Create `testdata/redundant-onerror-halt-test.sql`:

```sql
--liquibase formatted sql

-- Invalid: Redundant onError:HALT
--changeset test:1
--preconditions onFail:MARK_RAN onError:HALT
--precondition-sql-check expectedResult:0 SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'users'
CREATE TABLE users (id INT, name VARCHAR(100));

-- Valid: No onError (uses default)
--changeset test:2
--preconditions onFail:MARK_RAN
--precondition-sql-check expectedResult:0 SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'users' AND COLUMN_NAME = 'email'
ALTER TABLE users ADD COLUMN email VARCHAR(255);

-- Valid: Non-default onError value
--changeset test:3
--preconditions onFail:MARK_RAN onError:WARN
--precondition-sql-check expectedResult:0 SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'users' AND COLUMN_NAME = 'phone'
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
```

---

## Integration Steps

### 1. Register Rule in Main

Update `cmd/liquibase-linter/main.go`:

```go
// Register rules
allRules := []rules.Rule{
    // ... existing rules ...
    rules.NewRedundantOnErrorHaltRule(),
}
```

### 2. Update Documentation

File: `docs/rules.md`

Add entry:
```markdown
| redundant-onerror-halt | Info | Best Practices | Detects redundant onError:HALT (default value) |
```

---

## Validation Checklist

- [ ] Rule implementation added to `bestpractice.go`
- [ ] Tests added to `bestpractice_test.go`
- [ ] Test fixtures created in `testdata/`
- [ ] Configuration updated in `config.go`
- [ ] Rule registered in `main.go`
- [ ] Documentation file exists: `docs/rules/redundant-onerror-halt.md`
- [ ] Documentation updated in `docs/rules.md`
- [ ] All tests pass: `go test ./...`
- [ ] Rule tested manually with real changelogs
- [ ] Case-insensitive matching works (HALT, halt, Halt)

---

## Expected Output Example

```
INFO: redundant-onerror-halt
File: changelog/sprints/v123/0 - structure/add-column.xml
Changeset: john.doe:456
Message: Redundant 'onError="HALT"' - this is the default behavior and can be omitted
```

---

## Dependencies

- Parser already supports `Precondition.OnError` ✅
- No parser modifications needed ✅
- Uses existing `Violation` structure ✅

---

## Related Documentation

- [docs/rules/redundant-onerror-halt.md](../docs/rules/redundant-onerror-halt.md)
- [docs/rules/mandatory-preconditions.md](../docs/rules/mandatory-preconditions.md) (related)
- [docs/rules/missing-preconditions.md](../docs/rules/missing-preconditions.md) (related)
- [Liquibase Preconditions Documentation](https://docs.liquibase.com/concepts/changelogs/preconditions.html)

---

## Notes

- This is the simplest rule to implement (1-2 hours including tests)
- Good candidate for first implementation from this roadmap
- Could be enhanced with auto-fix capability in the future
- Case-insensitive check ensures we catch "HALT", "halt", "Halt", etc.
- Only informational severity - doesn't indicate a real problem, just a style preference

---

## Future Enhancement: Auto-fix

This rule is an excellent candidate for auto-fix functionality:

```go
// AutoFix removes redundant onError:HALT from changelog files
func (r *RedundantOnErrorHaltRule) AutoFix(changelog *parser.Changelog) error {
    // Implementation would:
    // 1. Read the original file
    // 2. Remove onError="HALT" attributes
    // 3. Write back the modified file
    // 4. Preserve formatting and comments
}
```

This could be implemented in a future phase once the basic rule is working.
