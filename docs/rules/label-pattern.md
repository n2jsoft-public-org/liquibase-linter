# label-pattern

**Severity**: Warning  
**Category**: Best Practices  
**Status**: 📋 Planned

## Description

Enforces that all changesets have labels that match a configured pattern (e.g., sprint version labels like "v123"). Labels are used to selectively deploy changesets and provide traceability by associating changes with specific releases or sprints.

## What it detects

- Changesets without any labels
- Changesets with labels that don't match the configured pattern
- Multiple labels where at least one doesn't match the pattern

## Example violation

**No label:**
```xml
<changeSet id="1" author="john.doe">
    <addColumn tableName="users">
        <column name="email" type="VARCHAR(255)"/>
    </addColumn>
</changeSet>
```

**Invalid label format:**
```xml
<changeSet id="1" author="john.doe" labels="sprint123">
    <addColumn tableName="users">
        <column name="email" type="VARCHAR(255)"/>
    </addColumn>
</changeSet>
```

**SQL format without label:**
```sql
--changeset john.doe:1
ALTER TABLE users ADD COLUMN email VARCHAR(255);
```

## Correct usage

```xml
<changeSet id="1" author="john.doe" labels="v123">
    <preConditions onFail="MARK_RAN">
        <not>
            <columnExists tableName="users" columnName="email"/>
        </not>
    </preConditions>
    <addColumn tableName="users">
        <column name="email" type="VARCHAR(255)"/>
    </addColumn>
</changeSet>
```

**SQL format:**
```sql
--changeset john.doe:1
--labels: v123
--preconditions onFail:MARK_RAN
--precondition-sql-check expectedResult:0 SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'users' AND COLUMN_NAME = 'email'
ALTER TABLE users ADD COLUMN email VARCHAR(255);
```

**Multiple labels (all matching pattern):**
```xml
<changeSet id="1" author="john.doe" labels="v123, hotfix">
    <addColumn tableName="users">
        <column name="email" type="VARCHAR(255)"/>
    </addColumn>
</changeSet>
```

## Why this matters

### Deployment control
Labels allow selective deployment of changesets using the `--labels` runtime parameter, enabling feature flags and controlled rollouts.

### Traceability
Labels link database changes to specific releases, sprints, or features, making it easier to track which changes were deployed in which version.

### Organization
Consistent labeling creates a standardized way to organize and filter changesets across large projects with many contributors.

### Release management
Labels facilitate release planning by clearly marking which changes belong to which sprint or version.

## Configuration

### Default configuration
```yaml
rules:
  label-pattern:
    enabled: true
    severity: warning
    # Pattern for valid labels (default: matches v followed by digits, e.g., v123)
    pattern: '^v\d+$'
    # Require at least one label per changeset
    require_label: true
```

### Custom patterns

**Sprint-based (default):**
```yaml
rules:
  label-pattern:
    enabled: true
    pattern: '^v\d+$'  # Matches: v116, v117, v123
```

**Semantic versioning:**
```yaml
rules:
  label-pattern:
    enabled: true
    pattern: '^v\d+\.\d+\.\d+$'  # Matches: v1.2.3, v2.0.0
```

**Feature flags:**
```yaml
rules:
  label-pattern:
    enabled: true
    pattern: '^(feature|bugfix|hotfix|release)-[a-z0-9-]+$'
    # Matches: feature-user-auth, bugfix-123, hotfix-security
```

**Jira ticket format:**
```yaml
rules:
  label-pattern:
    enabled: true
    pattern: '^[A-Z]+-\d+$'  # Matches: PROJ-123, BUG-456
```

**Multiple patterns (any match is valid):**
```yaml
rules:
  label-pattern:
    enabled: true
    patterns:
      - '^v\d+$'           # Sprint versions
      - '^hotfix$'          # Hotfix label
      - '^init$'            # Initial setup
```

## Exceptions

Some changesets may not require labels:

- **Initialization scripts**: Database setup scripts in the `init/` directory
- **Shared utilities**: Common procedures or functions used across versions

To exclude specific patterns:

```yaml
rules:
  label-pattern:
    enabled: true
    severity: warning
    pattern: '^v\d+$'
    exclude_patterns:
      - "**/init/**"
      - "**/utilities/**"
```

## Label validation logic

### Single pattern mode
- If `require_label: true`, changeset must have at least one label
- All labels must match the configured pattern
- Empty labels or whitespace-only labels are invalid

### Multiple patterns mode
- At least one label must be present
- Each label must match at least one of the configured patterns
- Violation only if a label matches none of the patterns

## Related rules

- [sprint-folder-structure](file-structure-sprint.md): Enforces folder organization by sprint
- [changelog-organization](changelog-organization.md): Ensures proper changelog file organization

## Implementation notes

### Detection strategy

1. Check if `ChangeSet.Labels` is empty or nil
   - If `require_label: true`, report violation
2. For each label in `ChangeSet.Labels`:
   - Trim whitespace
   - Test against configured pattern(s)
   - Report violation if no pattern matches
3. Apply exclude patterns if configured

### Parser support

All parsers already support labels:
- **XML**: Parse `labels` attribute (comma-separated)
- **YAML**: Parse `labels` key (string or array)
- **SQL**: Parse `--labels:` comment (comma-separated)

Labels are stored in `ChangeSet.Labels` as `[]string`.

### Configuration structure

```go
type LabelPatternConfig struct {
    Enabled         bool     `yaml:"enabled"`
    Severity        string   `yaml:"severity"`
    Pattern         string   `yaml:"pattern"`          // Single pattern
    Patterns        []string `yaml:"patterns"`         // Multiple patterns (alternative)
    RequireLabel    bool     `yaml:"require_label"`    // Default: true
    ExcludePatterns []string `yaml:"exclude_patterns"`
}
```

### Testing considerations

- Test with no labels
- Test with single valid label
- Test with multiple valid labels
- Test with invalid label format
- Test with mixed valid/invalid labels
- Test with empty string labels
- Test with whitespace-only labels
- Test exclude patterns
- Test across all changelog formats (XML, YAML, SQL)
- Test with multiple patterns configuration

## Examples from your codebase

Based on your file structure (`v116/`, `v117/`, `v123/`, etc.), the default pattern `^v\d+$` is appropriate:

**Expected usage:**
```sql
--changeset maxime.charles:1234
--labels: v123
--preconditions onFail:MARK_RAN
--precondition-sql-check expectedResult:0 SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = 'dbo' AND TABLE_NAME = 'NewTable'
CREATE TABLE dbo.NewTable (...)
```

This ensures that all changesets are tagged with the sprint version they belong to, matching your folder structure convention.
