# non-idempotent (Preconditions)

**Severity**: Warning  
**Category**: Reliability  
**Status**: ✅ Implemented

## Description

Detects changesets that lack preconditions, which can lead to non-idempotent operations and deployment failures. This rule supports two enforcement modes:

- **risky-only** (default): Only flags specific risky operations (CREATE TABLE, ADD COLUMN, etc.)
- **all**: Requires preconditions on every changeset (strictest enforcement)

This rule consolidates the functionality of the previously separate `missing-preconditions` and `mandatory-preconditions` rules into a single configurable rule.

## What it detects

### Mode: "risky-only" (default)
Detects specific operations known to be risky without preconditions:
- CREATE TABLE without existence checks
- CREATE INDEX without existence checks
- ADD COLUMN without existence checks
- INSERT statements without preconditions
- Other structural changes that may fail on re-run

### Mode: "all"
Requires preconditions on **every** changeset, regardless of operation type. This is the strictest enforcement mode and promotes defensive database change management.

## Example violation

**XML format:**
```xml
<changeSet id="1" author="test">
    <sql>
        CREATE TABLE users (id INT);
    </sql>
</changeSet>
```

**SQL format:**
```sql
--liquibase formatted sql
--changeset test:1
CREATE TABLE users (id INT);

--changeset test:2
INSERT INTO settings (key, value) VALUES ('version', '1.0');
```

## Correct usage

**XML format:**
```xml
<changeSet id="1" author="test">
    <preConditions onFail="MARK_RAN">
        <not>
            <tableExists tableName="users"/>
        </not>
    </preConditions>
    <createTable tableName="users">
        <column name="id" type="INT"/>
    </createTable>
</changeSet>
```

**SQL format:**
```sql
--liquibase formatted sql
--changeset test:1
--preconditions onFail:MARK_RAN
--precondition-sql-check expectedResult:0 SELECT COUNT(*) FROM information_schema.tables WHERE table_name='users'
CREATE TABLE users (id INT);

--changeset test:2
CREATE TABLE IF NOT EXISTS users (id INT);
```

Or use Liquibase's built-in idempotency by relying on the DATABASECHANGELOG tracking.

## Configuration

### Basic Configuration

```yaml
rules:
  non-idempotent:
    enabled: true
    severity: warning
    mode: "risky-only"  # or "all"
```
### Idempotency Benefits
- **Retry Safety**: Changes can be safely re-run without causing errors
- **Development Efficiency**: Developers can reset and re-apply changes during development
- **Deployment Reliability**: Failed deployments can be retried without manual intervention
- **Testing**: Integration tests can be run repeatedly with consistent results

### Safety and Documentation
Preconditions provide multiple benefits beyond idempotency:
- **Guards**: Verify database state before applying changes
- **Documentation**: Make assumptions and requirements explicit
- **Error Prevention**: Catch issues early before they cause partial deployments
- **Intent Clarity**: Show what conditions must be true for the change to be safe

### Mode Selection Guide

**Use `mode: "risky-only"`** (default) when:
- Migrating existing changelogs to use preconditions
- Team is learning Liquibase best practices
- Some legacy scripts legitimately don't need preconditions
- Gradual adoption is preferred

**Use `mode: "all"`** when:
- Establishing strict standards for new projects
- Maximum safety and idempotency are required
- Team is experienced with Liquibase preconditions
- Combined with exclude patterns for legitimate exceptions

## Exceptions and Patterns

Some changesets may legitimately not require preconditions:

### Inherently Idempotent Operations
- `CREATE OR REPLACE VIEW` (database-specific)
- Operations using Liquibase's DATABASECHANGELOG tracking
- Changes with `runAlways="true"` (already flagged by parser)

### Initialization and Setup
- Initial schema creation in `init/` directories
- Seed data that uses conflict resolution (INSERT ... ON CONFLICT)
- Test fixtures and example data

### Configuration for Exceptions
Use `exclude-patterns` to exempt specific paths:
```yaml
rules:
  non-idempotent:
    mode: "all"
    exclude-patterns:
      - "**/init/**"
      - "**/seed/**"
      - "**/examples/**"
```

## Migration from Previous Rules

This rule consolidates three previously separate precondition rules:

| Old Rule                  | Status              | Equivalent Configuration       |
| ------------------------- | ------------------- | ------------------------------ |
| `non-idempotent`          | ✅ Implemented       | `mode: "risky-only"` (default) |
| `missing-preconditions`   | ❌ Never implemented | Same as `non-idempotent`       |
| `mandatory-preconditions` | 📋 Planned           | `mode: "all"`                  |

If you have documentation or configuration referencing `mandatory-preconditions` or `missing-preconditions`, update to use `non-idempotent` with the appropriate `mode` setting.

## Related Rules

- [dangerous-operations](dangerous-operations.md): Requires preconditions specifically for DROP/TRUNCATE operations
- [redundant-onerror-halt](redundant-onerror-halt.md): Detects redundant default values in preconditions
- [no-if-exists](no-if-exists.md): Recommends Liquibase preconditions over database-specific IF EXISTS patterns

## See Also

- [Liquibase Preconditions Documentation](https://docs.liquibase.com/concepts/changelogs/preconditions.html)
- [Best Practices for Idempotent Database Changes](https://docs.liquibase.com/concepts/bestpractices.html)
  - `"all"`: Require preconditions on every changeset
- **exclude-patterns**: Glob patterns for files/paths to exclude (optional)

### Advanced Configuration Examples

#### Strict enforcement with exclusions
```yaml
rules:
  non-idempotent:
    enabled: true
    severity: error
    mode: "all"
    exclude-patterns:
      - "**/init/**"              # Exclude initialization scripts
      - "**/seed/**"              # Exclude seed data
      - "**/test-data/**"         # Exclude test fixtures
```

#### Risky operations only (default)
```yaml
rules:
  non-idempotent:
    enabled: true
    severity: warning
    mode: "risky-only"
```

#### Team migration path
```yaml
# Start with info severity to identify issues
rules:
  non-idempotent:
    enabled: true
    severity: info
    mode: "risky-only"

# Later, escalate to warnings
# severity: warning

# Finally, enforce strictly
# severity: error
# mode: "all"
```

## Why this matters

Non-idempotent changes can:
- Fail when deployment is retried
- Cause deployment pipeline failures
- Make testing and development more difficult
- Lead to inconsistent database states

Idempotent changes can be safely re-run multiple times with the same result, which is crucial for reliable deployments.
