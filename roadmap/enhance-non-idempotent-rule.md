# Enhance Non-Idempotent Rule with Configurable Modes

## Objective

Enhance the existing `non-idempotent` rule to support configurable enforcement modes and exclude patterns, consolidating the functionality of three overlapping precondition-related rules into a single, flexible rule.

## Background

Currently, there are three precondition-related rules that overlap significantly:

1. **non-idempotent** (✅ Implemented): Checks specific risky operations for missing preconditions
2. **missing-preconditions** (❌ Never implemented): Documented but never coded, intended to do what non-idempotent already does
3. **mandatory-preconditions** (📋 Planned): Would require preconditions on ALL changesets

This creates confusion and maintenance overhead. The solution is to enhance the existing `non-idempotent` rule with configurable modes.

## Requirements

### Functional Requirements

1. **Mode Configuration**: Support two enforcement modes:
   - `"risky-only"`: Current behavior - only check specific risky operations (default for backward compatibility)
   - `"all"`: Check ALL changesets, requiring every one to have preconditions

2. **Exclude Patterns**: Support glob pattern matching to exclude specific paths:
   - `**/init/**` - initialization scripts
   - `**/seed/**` - seed data
   - `**/test-data/**` - test fixtures
   - Custom user-defined patterns

3. **Backward Compatibility**: Existing configurations must continue to work without changes

4. **Clear Messaging**: Violation messages should indicate which mode detected the issue

### Non-Functional Requirements

1. **Performance**: Glob pattern matching should be efficient
2. **Configuration Flexibility**: Support both simple and advanced configuration styles
3. **Documentation**: Comprehensive migration guide from old rule names
4. **Testing**: Extensive test coverage for both modes and exclude patterns

## Implementation Details

### Configuration Schema Enhancement

**File**: `internal/config/config.go`

Add new fields to support the enhanced configuration:

```go
// RuleConfig represents configuration for a single rule
type RuleConfig struct {
	Enabled         bool     `yaml:"enabled"`
	Severity        Severity `yaml:"severity"`
	Mode            string   `yaml:"mode"`             // NEW: enforcement mode
	ExcludePatterns []string `yaml:"exclude-patterns"` // NEW: glob patterns to exclude
}

// Constants for non-idempotent rule modes
const (
	NonIdempotentModeRiskyOnly = "risky-only"
	NonIdempotentModeAll       = "all"
)
```

Update the default configuration:

```go
func DefaultConfig() *Config {
	return &Config{
		Rules: map[string]RuleConfig{
			// ... existing rules ...
			"non-idempotent": {
				Enabled:         true,
				Severity:        SeverityWarning,
				Mode:            NonIdempotentModeRiskyOnly, // Default to current behavior
				ExcludePatterns: []string{},                 // No exclusions by default
			},
			// ... other rules ...
		},
	}
}
```

### Rule Enhancement

**File**: `internal/rules/bestpractice.go`

Enhance the existing `NonIdempotentChangesRule`:

```go
// NonIdempotentChangesRule checks for operations that may fail on re-run
type NonIdempotentChangesRule struct {
	config config.RuleConfig
}

func NewNonIdempotentChangesRule(cfg config.RuleConfig) *NonIdempotentChangesRule {
	// Set default mode if not specified
	if cfg.Mode == "" {
		cfg.Mode = config.NonIdempotentModeRiskyOnly
	}
	return &NonIdempotentChangesRule{config: cfg}
}

func (r *NonIdempotentChangesRule) ID() string {
	return "non-idempotent"
}

func (r *NonIdempotentChangesRule) Name() string {
	return "Non-Idempotent Changes"
}

func (r *NonIdempotentChangesRule) Description() string {
	return "Checks for changes that lack preconditions and may fail on re-run"
}

func (r *NonIdempotentChangesRule) Severity() config.Severity {
	return r.config.Severity
}

func (r *NonIdempotentChangesRule) Check(changelog *parser.Changelog) []config.Violation {
	if !r.config.Enabled {
		return nil
	}

	// Check if file should be excluded based on patterns
	if r.shouldExcludeFile(changelog.FilePath) {
		return nil
	}

	var violations []config.Violation
	seenChangesets := make(map[string]bool) // Track to avoid duplicate violations

	for _, cs := range changelog.ChangeSets {
		// Skip if changeset runs always (already handled elsewhere)
		if cs.RunAlways {
			continue
		}

		// Skip if we've already reported this changeset
		if seenChangesets[cs.ID] {
			continue
		}

		// Check based on mode
		var shouldCheck bool
		var reason string

		switch r.config.Mode {
		case config.NonIdempotentModeAll:
			// In "all" mode, every changeset without preconditions is a violation
			shouldCheck = !cs.HasPreconditions()
			reason = "All changesets require preconditions in strict mode"

		case config.NonIdempotentModeRiskyOnly:
			// In "risky-only" mode (default), only check specific risky operations
			shouldCheck = r.hasRiskyOperationWithoutPrecondition(cs)
			reason = "This operation may fail on re-run without preconditions"

		default:
			// Unknown mode, default to risky-only behavior
			shouldCheck = r.hasRiskyOperationWithoutPrecondition(cs)
			reason = "This operation may fail on re-run without preconditions"
		}

		if shouldCheck {
			violations = append(violations, config.Violation{
				Rule:        r.ID(),
				Severity:    r.Severity(),
				Message:     fmt.Sprintf("%s. Add preconditions to ensure idempotency (mode: %s)", reason, r.config.Mode),
				FilePath:    changelog.FilePath,
				ChangeSetID: cs.ID,
				Suggestion:  r.getSuggestion(cs),
			})
			seenChangesets[cs.ID] = true
		}
	}

	return violations
}

// shouldExcludeFile checks if the file path matches any exclude patterns
func (r *NonIdempotentChangesRule) shouldExcludeFile(filePath string) bool {
	if len(r.config.ExcludePatterns) == 0 {
		return false
	}

	for _, pattern := range r.config.ExcludePatterns {
		matched, err := filepath.Match(pattern, filePath)
		if err != nil {
			// Log error but don't exclude if pattern is invalid
			continue
		}
		if matched {
			return true
		}

		// Also try matching against the path components
		// This allows patterns like "**/init/**" to work
		if matchesGlobPattern(filePath, pattern) {
			return true
		}
	}

	return false
}

// matchesGlobPattern checks if a path matches a glob pattern with ** support
func matchesGlobPattern(path, pattern string) bool {
	// Convert pattern to regex for ** support
	// This is a simplified implementation - for production, consider using
	// a library like github.com/bmatcuk/doublestar
	
	// Replace ** with wildcard that matches across path separators
	regexPattern := strings.ReplaceAll(pattern, "**", ".*")
	regexPattern = strings.ReplaceAll(regexPattern, "*", "[^/]*")
	regexPattern = "^" + regexPattern + "$"
	
	matched, err := regexp.MatchString(regexPattern, path)
	return err == nil && matched
}

// hasRiskyOperationWithoutPrecondition checks if changeset has risky operations without preconditions
// This is the existing logic from the current implementation
func (r *NonIdempotentChangesRule) hasRiskyOperationWithoutPrecondition(cs parser.ChangeSet) bool {
	if cs.HasPreconditions() {
		return false
	}

	// Check for risky change types
	for _, change := range cs.Changes {
		if change.CreateTable != nil ||
			change.CreateIndex != nil ||
			change.AddColumn != nil ||
			change.Insert != nil {
			return true
		}

		// For SQL changes, check if they contain risky DDL
		if change.SQL != "" {
			sql := strings.ToUpper(change.SQL)
			if strings.Contains(sql, "CREATE TABLE") ||
				strings.Contains(sql, "CREATE INDEX") ||
				strings.Contains(sql, "ALTER TABLE") ||
				strings.Contains(sql, "INSERT INTO") {
				return true
			}
		}
	}

	return false
}

// getSuggestion provides context-appropriate suggestions
func (r *NonIdempotentChangesRule) getSuggestion(cs parser.ChangeSet) string {
	switch r.config.Mode {
	case config.NonIdempotentModeAll:
		return "Add preconditions like <preConditions onFail=\"MARK_RAN\"><tableExists/><columnExists/></preConditions> or use exclude-patterns to exempt this file"
	case config.NonIdempotentModeRiskyOnly:
		return "Add preconditions to check object existence before creating: <preConditions onFail=\"MARK_RAN\"><not><tableExists tableName=\"...\"/></not></preConditions>"
	default:
		return "Add preconditions to ensure this changeset can be safely re-run"
	}
}
```

### Glob Pattern Matching Utility

For production-quality glob matching with `**` support, consider adding a dependency:

**Option 1**: Use existing `filepath.Match` with manual path splitting (simple but limited)

**Option 2**: Add dependency on `github.com/bmatcuk/doublestar` (recommended):

```go
import "github.com/bmatcuk/doublestar/v4"

func (r *NonIdempotentChangesRule) shouldExcludeFile(filePath string) bool {
	if len(r.config.ExcludePatterns) == 0 {
		return false
	}

	for _, pattern := range r.config.ExcludePatterns {
		matched, err := doublestar.Match(pattern, filePath)
		if err != nil {
			continue
		}
		if matched {
			return true
		}
	}

	return false
}
```

Add to `go.mod`:
```bash
go get github.com/bmatcuk/doublestar/v4
```

## Testing

### Unit Tests

**File**: `internal/rules/bestpractice_test.go`

Enhance existing tests and add new test cases:

```go
func TestNonIdempotentChangesRule_ModeRiskyOnly(t *testing.T) {
	tests := []struct {
		name           string
		changelog      *parser.Changelog
		wantViolations int
	}{
		{
			name: "Risky operation without preconditions",
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
			wantViolations: 1,
		},
		{
			name: "Risky operation with preconditions",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:           "1",
						Author:       "test",
						Preconditions: &parser.Precondition{OnFail: "MARK_RAN"},
						Changes: []parser.Change{
							{CreateTable: &parser.CreateTable{TableName: "users"}},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "Non-risky operation without preconditions",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{AddNotNullConstraint: &parser.AddNotNullConstraint{TableName: "users"}},
						},
					},
				},
			},
			wantViolations: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewNonIdempotentChangesRule(config.RuleConfig{
				Enabled:  true,
				Severity: config.SeverityWarning,
				Mode:     config.NonIdempotentModeRiskyOnly,
			})

			violations := rule.Check(tt.changelog)

			if len(violations) != tt.wantViolations {
				t.Errorf("Check() violations = %d, want %d", len(violations), tt.wantViolations)
			}
		})
	}
}

func TestNonIdempotentChangesRule_ModeAll(t *testing.T) {
	tests := []struct {
		name           string
		changelog      *parser.Changelog
		wantViolations int
	}{
		{
			name: "Any changeset without preconditions",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{AddNotNullConstraint: &parser.AddNotNullConstraint{TableName: "users"}},
						},
					},
				},
			},
			wantViolations: 1,
		},
		{
			name: "Changeset with preconditions",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{
						ID:           "1",
						Author:       "test",
						Preconditions: &parser.Precondition{OnFail: "MARK_RAN"},
						Changes: []parser.Change{
							{AddNotNullConstraint: &parser.AddNotNullConstraint{TableName: "users"}},
						},
					},
				},
			},
			wantViolations: 0,
		},
		{
			name: "Multiple changesets without preconditions",
			changelog: &parser.Changelog{
				FilePath: "test.xml",
				ChangeSets: []parser.ChangeSet{
					{ID: "1", Author: "test", Changes: []parser.Change{{SQL: "UPDATE users SET active = 1"}}},
					{ID: "2", Author: "test", Changes: []parser.Change{{SQL: "DELETE FROM logs WHERE date < '2020-01-01'"}}},
				},
			},
			wantViolations: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewNonIdempotentChangesRule(config.RuleConfig{
				Enabled:  true,
				Severity: config.SeverityWarning,
				Mode:     config.NonIdempotentModeAll,
			})

			violations := rule.Check(tt.changelog)

			if len(violations) != tt.wantViolations {
				t.Errorf("Check() violations = %d, want %d", len(violations), tt.wantViolations)
			}
		})
	}
}

func TestNonIdempotentChangesRule_ExcludePatterns(t *testing.T) {
	tests := []struct {
		name            string
		filePath        string
		excludePatterns []string
		wantExcluded    bool
	}{
		{
			name:            "Match init directory",
			filePath:        "changelog/init/0-structure/tables.sql",
			excludePatterns: []string{"**/init/**"},
			wantExcluded:    true,
		},
		{
			name:            "Match seed directory",
			filePath:        "db/seed/test-data.sql",
			excludePatterns: []string{"**/seed/**"},
			wantExcluded:    true,
		},
		{
			name:            "No match",
			filePath:        "changelog/sprints/v123/structure/tables.sql",
			excludePatterns: []string{"**/init/**", "**/seed/**"},
			wantExcluded:    false,
		},
		{
			name:            "Multiple patterns with match",
			filePath:        "changelog/init/data.sql",
			excludePatterns: []string{"**/test/**", "**/init/**", "**/examples/**"},
			wantExcluded:    true,
		},
		{
			name:            "Specific file pattern",
			filePath:        "db/changelog/001-create-schema.sql",
			excludePatterns: []string{"**/*-create-schema.sql"},
			wantExcluded:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rule := NewNonIdempotentChangesRule(config.RuleConfig{
				Enabled:         true,
				Severity:        config.SeverityWarning,
				Mode:            config.NonIdempotentModeAll,
				ExcludePatterns: tt.excludePatterns,
			})

			changelog := &parser.Changelog{
				FilePath: tt.filePath,
				ChangeSets: []parser.ChangeSet{
					{
						ID:     "1",
						Author: "test",
						Changes: []parser.Change{
							{SQL: "CREATE TABLE test (id INT)"},
						},
					},
				},
			}

			violations := rule.Check(changelog)

			gotExcluded := len(violations) == 0
			if gotExcluded != tt.wantExcluded {
				t.Errorf("shouldExclude() = %v, want %v", gotExcluded, tt.wantExcluded)
			}
		})
	}
}

func TestNonIdempotentChangesRule_BackwardCompatibility(t *testing.T) {
	// Test that existing configurations without mode specified still work
	rule := NewNonIdempotentChangesRule(config.RuleConfig{
		Enabled:  true,
		Severity: config.SeverityWarning,
		// Mode not specified - should default to "risky-only"
	})

	changelog := &parser.Changelog{
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
	}

	violations := rule.Check(changelog)

	if len(violations) != 1 {
		t.Errorf("Backward compatibility broken: expected 1 violation, got %d", len(violations))
	}
}
```

## Integration Steps

1. **Update configuration schema** in `internal/config/config.go`
   - Add `Mode` and `ExcludePatterns` fields
   - Add constants for mode values
   - Update default config

2. **Enhance rule implementation** in `internal/rules/bestpractice.go`
   - Add mode checking logic
   - Implement exclude pattern matching
   - Update violation messages
   - Add suggestion method

3. **Add glob pattern dependency** (optional but recommended)
   - `go get github.com/bmatcuk/doublestar/v4`

4. **Update tests** in `internal/rules/bestpractice_test.go`
   - Add mode-specific tests
   - Add exclude pattern tests
   - Add backward compatibility tests

5. **Update documentation**
   - Enhance `docs/rules/non-idempotent.md` ✅ DONE
   - Mark `docs/rules/mandatory-preconditions.md` as merged ✅ DONE
   - Mark `docs/rules/missing-preconditions.md` as merged ✅ DONE
   - Update `docs/configuration.md` with new options

6. **Update roadmap**
   - Remove `mandatory-preconditions.md` from planned rules
   - Update `roadmap/README.md` ✅ DONE

7. **Run tests and validation**
   - `go test ./internal/rules/... -v`
   - `go test ./internal/config/... -v`
   - Test with real changelogs

8. **Update examples and README**
   - Add configuration examples
   - Update rules list

## Validation Checklist

- [ ] `mode: "risky-only"` works (existing behavior)
- [ ] `mode: "all"` enforces preconditions on every changeset
- [ ] Exclude patterns work with simple globs (`*.sql`)
- [ ] Exclude patterns work with `**` recursive matching
- [ ] Exclude patterns work with directory patterns (`**/init/**`)
- [ ] Configurations without `mode` default to `"risky-only"`
- [ ] Violation messages indicate which mode detected the issue
- [ ] Suggestions are appropriate for each mode
- [ ] All existing tests still pass
- [ ] New tests cover both modes and exclude patterns
- [ ] Documentation is clear and comprehensive
- [ ] Migration path from old rule names is documented

## Expected Output

### Mode: "risky-only" (default)
```
$ ./liquibase-linter check changelog/sprints/v123/structure/tables.sql

Checking: changelog/sprints/v123/structure/tables.sql
✗ changelog/sprints/v123/structure/tables.sql

WARNING: [non-idempotent] This operation may fail on re-run without preconditions (mode: risky-only)
  ChangeSet: john.doe:create-users-table
  File: changelog/sprints/v123/structure/tables.sql
  Suggestion: Add preconditions to check object existence before creating: <preConditions onFail="MARK_RAN"><not><tableExists tableName="..."/></not></preConditions>
```

### Mode: "all"
```
$ ./liquibase-linter check --config=strict-config.yaml changelog/

Checking: changelog/sprints/v123/structure/tables.sql
✗ changelog/sprints/v123/structure/tables.sql

ERROR: [non-idempotent] All changesets require preconditions in strict mode (mode: all)
  ChangeSet: john.doe:update-user-status
  File: changelog/sprints/v123/data/updates.sql
  Suggestion: Add preconditions like <preConditions onFail="MARK_RAN"><tableExists/><columnExists/></preConditions> or use exclude-patterns to exempt this file

Files excluded by patterns: 3 (init/, seed/)
```

## Dependencies

- **Parser**: No changes needed - `HasPreconditions()` already exists
- **Config**: Schema enhancement needed
- **Glob Matching**: Optional dependency on `github.com/bmatcuk/doublestar/v4` (recommended)
- **Tests**: New test cases needed

## Estimated Effort

- **Configuration Schema**: 30 minutes
- **Rule Enhancement**: 2-3 hours
- **Glob Pattern Matching**: 1 hour
- **Testing**: 2 hours
- **Documentation Updates**: 1 hour (already done)
- **Integration & Validation**: 1 hour
- **Total**: 7-8 hours

## Migration Notes

### For Users

If you have existing configurations, no changes are required. The default behavior (`mode: "risky-only"`) matches the current implementation.

To enforce strict preconditions:
```yaml
rules:
  non-idempotent:
    enabled: true
    severity: error
    mode: "all"
    exclude-patterns:
      - "**/init/**"
```

### For Documentation

Old rule references should be updated:
- `mandatory-preconditions` → `non-idempotent` with `mode: "all"`
- `missing-preconditions` → `non-idempotent` with `mode: "risky-only"` (default)

## Notes

- This consolidation reduces code duplication and maintenance overhead
- The flexible configuration allows teams to adopt preconditions gradually
- Exclude patterns provide escape hatches for legitimate exceptions
- Backward compatibility is maintained for existing users
- The rule name `non-idempotent` is kept for consistency (not renamed to "preconditions")
