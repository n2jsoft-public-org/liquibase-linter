# Task: Implement Label Pattern Rule

**Status**: 📋 Planned  
**Priority**: High  
**Estimated Effort**: 3-4 hours  
**Category**: Best Practices Rule  
**Target File**: `internal/rules/bestpractice.go`

---

## Objective

Implement a linting rule that enforces changeset labels match a configured pattern (e.g., sprint versions like "v123").

---

## Requirements

### Functional Requirements
- Validate labels against configured regex pattern(s)
- Support single or multiple patterns
- Optionally require at least one label per changeset
- Support exclude patterns for exempting specific files
- Work across all changelog formats (XML, YAML, SQL)

### Non-Functional Requirements
- Rule should be disabled by default (opt-in for projects wanting strict labeling)
- Performance: O(n*m) where n = changesets, m = labels per changeset
- Clear error messages showing which label is invalid

---

## Implementation Details

### 1. Configuration Structure

Add to `internal/config/config.go`:

```go
// LabelPatternConfig represents configuration for label pattern rule.
type LabelPatternConfig struct {
    Enabled         bool     `yaml:"enabled"`
    Pattern         string   `yaml:"pattern"`          // Single pattern
    Patterns        []string `yaml:"patterns"`         // Multiple patterns (alternative)
    RequireLabel    bool     `yaml:"require_label"`    // Default: true
    ExcludePatterns []string `yaml:"exclude_patterns"`
}
```

Update `Config` struct:
```go
type Config struct {
    // ... existing fields ...
    LabelPattern LabelPatternConfig `yaml:"label_pattern"`
}
```

### 2. Type Definition

Add to `internal/rules/bestpractice.go`:

```go
// LabelPatternRule enforces label naming conventions.
type LabelPatternRule struct {
    patterns        []*regexp.Regexp
    requireLabel    bool
    excludePatterns []string
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
```

### 3. Rule Interface Implementation

```go
func (r *LabelPatternRule) ID() string {
    return "label-pattern"
}

func (r *LabelPatternRule) Name() string {
    return "Label Pattern Enforcement"
}

func (r *LabelPatternRule) Description() string {
    return "Ensures changeset labels follow configured naming patterns"
}

func (r *LabelPatternRule) Severity() Severity {
    return SeverityWarning
}
```

### 4. Check Method

```go
func (r *LabelPatternRule) Check(changelog *parser.Changelog) []Violation {
    violations := make([]Violation, 0)

    for _, cs := range changelog.ChangeSets {
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
                    Message:     fmt.Sprintf("Label '%s' does not match required pattern", label),
                    FilePath:    cs.FilePath,
                    ChangeSetID: cs.ID,
                    Author:      cs.Author,
                })
            }
        }
    }

    return violations
}
```

### 5. Helper Methods

```go
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

func (r *LabelPatternRule) shouldExclude(filePath string) bool {
    if len(r.excludePatterns) == 0 {
        return false
    }

    for _, pattern := range r.excludePatterns {
        matched, err := filepath.Match(pattern, filePath)
        if err == nil && matched {
            return true
        }
        
        // Also try matching with the base name
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
"label-pattern": {
    Enabled:  false,  // Disabled by default (opt-in)
    Severity: "warning",
},
LabelPattern: LabelPatternConfig{
    Pattern:         `^v\d+$`,
    RequireLabel:    true,
    ExcludePatterns: []string{"**/init/**"},
},
```

### Configuration Schema

```yaml
rules:
  label-pattern:
    enabled: false
    severity: warning

label_pattern:
  pattern: '^v\d+$'
  require_label: true
  exclude_patterns:
    - "**/init/**"
    - "**/utilities/**"
```

### Multiple Patterns Example

```yaml
label_pattern:
  patterns:
    - '^v\d+$'           # Sprint versions (v123)
    - '^hotfix$'          # Hotfix label
    - '^init$'            # Initial setup
  require_label: true
```

---

## Testing

### Test File Location
`internal/rules/bestpractice_test.go`

### Test Cases

```go
func TestLabelPatternRule_Check(t *testing.T) {
    tests := []struct {
        name          string
        rule          *LabelPatternRule
        changelog     *parser.Changelog
        wantViolation bool
        violationMsg  string
    }{
        {
            name: "valid single label",
            rule: NewLabelPatternRule(&config.LabelPatternConfig{
                Pattern:      `^v\d+$`,
                RequireLabel: true,
            }),
            changelog: &parser.Changelog{
                ChangeSets: []parser.ChangeSet{
                    {
                        ID:       "1",
                        Author:   "test",
                        FilePath: "/test/changelog.xml",
                        Labels:   []string{"v123"},
                    },
                },
            },
            wantViolation: false,
        },
        {
            name: "invalid label format",
            rule: NewLabelPatternRule(&config.LabelPatternConfig{
                Pattern:      `^v\d+$`,
                RequireLabel: true,
            }),
            changelog: &parser.Changelog{
                ChangeSets: []parser.ChangeSet{
                    {
                        ID:       "1",
                        Author:   "test",
                        FilePath: "/test/changelog.xml",
                        Labels:   []string{"sprint123"},
                    },
                },
            },
            wantViolation: true,
            violationMsg:  "does not match required pattern",
        },
        {
            name: "missing required label",
            rule: NewLabelPatternRule(&config.LabelPatternConfig{
                Pattern:      `^v\d+$`,
                RequireLabel: true,
            }),
            changelog: &parser.Changelog{
                ChangeSets: []parser.ChangeSet{
                    {
                        ID:       "1",
                        Author:   "test",
                        FilePath: "/test/changelog.xml",
                        Labels:   []string{},
                    },
                },
            },
            wantViolation: true,
            violationMsg:  "lacks required label",
        },
        {
            name: "multiple valid labels",
            rule: NewLabelPatternRule(&config.LabelPatternConfig{
                Patterns:     []string{`^v\d+$`, `^hotfix$`},
                RequireLabel: true,
            }),
            changelog: &parser.Changelog{
                ChangeSets: []parser.ChangeSet{
                    {
                        ID:       "1",
                        Author:   "test",
                        FilePath: "/test/changelog.xml",
                        Labels:   []string{"v123", "hotfix"},
                    },
                },
            },
            wantViolation: false,
        },
        {
            name: "empty label",
            rule: NewLabelPatternRule(&config.LabelPatternConfig{
                Pattern:      `^v\d+$`,
                RequireLabel: true,
            }),
            changelog: &parser.Changelog{
                ChangeSets: []parser.ChangeSet{
                    {
                        ID:       "1",
                        Author:   "test",
                        FilePath: "/test/changelog.xml",
                        Labels:   []string{""},
                    },
                },
            },
            wantViolation: true,
            violationMsg:  "empty label",
        },
        {
            name: "excluded file",
            rule: NewLabelPatternRule(&config.LabelPatternConfig{
                Pattern:         `^v\d+$`,
                RequireLabel:    true,
                ExcludePatterns: []string{"**/init/**"},
            }),
            changelog: &parser.Changelog{
                ChangeSets: []parser.ChangeSet{
                    {
                        ID:       "1",
                        Author:   "test",
                        FilePath: "/test/init/setup.sql",
                        Labels:   []string{},
                    },
                },
            },
            wantViolation: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            violations := tt.rule.Check(tt.changelog)
            
            if tt.wantViolation && len(violations) == 0 {
                t.Error("Expected violation but got none")
            }
            if !tt.wantViolation && len(violations) > 0 {
                t.Errorf("Expected no violations but got %d: %v", len(violations), violations)
            }
            if tt.wantViolation && tt.violationMsg != "" && len(violations) > 0 {
                if !strings.Contains(violations[0].Message, tt.violationMsg) {
                    t.Errorf("Expected message to contain '%s', got '%s'", tt.violationMsg, violations[0].Message)
                }
            }
        })
    }
}
```

### Test Fixtures

Create `testdata/label-pattern-test.xml`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<databaseChangeLog
    xmlns="http://www.liquibase.org/xml/ns/dbchangelog"
    xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
    xsi:schemaLocation="http://www.liquibase.org/xml/ns/dbchangelog
    http://www.liquibase.org/xml/ns/dbchangelog/dbchangelog-3.8.xsd">

    <!-- Valid: Label matches pattern v\d+ -->
    <changeSet id="1" author="test" labels="v123">
        <addColumn tableName="users">
            <column name="email" type="VARCHAR(255)"/>
        </addColumn>
    </changeSet>

    <!-- Invalid: No label -->
    <changeSet id="2" author="test">
        <addColumn tableName="users">
            <column name="phone" type="VARCHAR(20)"/>
        </addColumn>
    </changeSet>

    <!-- Invalid: Label doesn't match pattern -->
    <changeSet id="3" author="test" labels="sprint123">
        <addColumn tableName="users">
            <column name="address" type="VARCHAR(255)"/>
        </addColumn>
    </changeSet>

</databaseChangeLog>
```

---

## Integration Steps

### 1. Register Rule in Main

Update `cmd/liquibase-linter/main.go`:

```go
// Load label pattern config
labelPatternCfg := &cfg.LabelPattern
if !cfg.Rules["label-pattern"].Enabled {
    labelPatternCfg = nil
}

// Register rules
allRules := []rules.Rule{
    // ... existing rules ...
    rules.NewLabelPatternRule(labelPatternCfg),
}
```

### 2. Update Documentation

File: `docs/rules.md`

Add entry:
```markdown
| label-pattern | Warning | Best Practices | Enforces label naming patterns (e.g., v123) |
```

---

## Validation Checklist

- [ ] Configuration structure added to `config.go`
- [ ] Rule implementation added to `bestpractice.go`
- [ ] Tests added to `bestpractice_test.go`
- [ ] Test fixtures created in `testdata/`
- [ ] Configuration updated with defaults
- [ ] Rule registered in `main.go`
- [ ] Documentation file exists: `docs/rules/label-pattern.md`
- [ ] Documentation updated in `docs/rules.md`
- [ ] All tests pass: `go test ./...`
- [ ] Rule tested manually with real changelogs
- [ ] Multiple pattern support tested
- [ ] Exclude patterns work correctly

---

## Expected Output Example

```
WARNING: label-pattern
File: changelog/sprints/v123/0 - structure/add-column.xml
Changeset: john.doe:456
Message: Label 'sprint123' does not match required pattern

WARNING: label-pattern
File: changelog/sprints/v123/1 - data/insert-data.sql
Changeset: jane.smith:789
Message: Changeset lacks required label
```

---

## Dependencies

- Parser already supports `ChangeSet.Labels` ✅
- No parser modifications needed ✅
- Uses existing `Violation` structure ✅

---

## Related Documentation

- [docs/rules/label-pattern.md](../docs/rules/label-pattern.md)
- [docs/rules/sprint-folder-structure.md](../docs/rules/file-structure-sprint.md) (complementary)

---

## Notes

- Default pattern `^v\d+$` matches your sprint naming convention (v116, v117, etc.)
- Consider providing common pattern examples in documentation
- Regex compilation errors should be handled gracefully
