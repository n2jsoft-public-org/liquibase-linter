# mandatory-preconditions

**Severity**: Warning  
**Category**: Reliability  
**Status**: ⚠️ Merged into [non-idempotent](non-idempotent.md)

> **Note**: This rule has been merged into the `non-idempotent` rule with configurable modes. Use `non-idempotent` with `mode: "all"` to enforce mandatory preconditions on all changesets.

## Description

Ensures that every changeset has at least one precondition to prevent unintended execution and improve idempotency. While some operations can safely run without preconditions, enforcing this rule promotes defensive database change management and makes changesets more robust.

## What it detects

- Changesets without any precondition block
- Empty precondition blocks

## Example violation

```xml
<changeSet id="1" author="john.doe">
    <addColumn tableName="users">
        <column name="email" type="VARCHAR(255)"/>
    </addColumn>
</changeSet>
```

**SQL format:**
```sql
--changeset john.doe:1
ALTER TABLE users ADD COLUMN email VARCHAR(255);
```

## Correct usage

```xml
<changeSet id="1" author="john.doe">
    <preConditions onFail="MARK_RAN">
        <tableExists tableName="users"/>
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
--preconditions onFail:MARK_RAN
--precondition-sql-check expectedResult:0 SELECT COUNT(*) FROM INFORMATION_SCHEMA.COLUMNS WHERE TABLE_NAME = 'users' AND COLUMN_NAME = 'email'
ALTER TABLE users ADD COLUMN email VARCHAR(255);
```

## Why this matters

### Idempotency
Preconditions ensure that changesets can be safely re-run without causing errors or duplicate operations.

### Safety
Preconditions act as guards against unintended changes, verifying that the database state is as expected before applying modifications.

### Clear intent
Explicit preconditions document the assumptions and requirements for each change, making the changelog more self-documenting.

### Error prevention
Preconditions catch issues early, preventing partial or failed deployments that could leave the database in an inconsistent state.

## Configuration

This functionality is now available in the `non-idempotent` rule. Use:

```yaml
rules:
  non-idempotent:
    enabled: true
    severity: warning
    mode: "all"                    # Enforce on ALL changesets
    exclude-patterns:              # Optional exclusions
      - "**/init/**"
      - "**/seed/**"
```

For detailed configuration options, see [non-idempotent rule documentation](non-idempotent.md).

## Related rules

- [non-idempotent](non-idempotent.md): Detects specific change types that may fail on re-run
- [missing-preconditions](missing-preconditions.md): Detects risky operations without preconditions (less strict)
- [dangerous-operations](dangerous-operations.md): Requires preconditions for DROP/TRUNCATE operations

## Implementation notes

### Detection strategy

1. Check if `ChangeSet.Preconditions` is `nil` or empty
2. Report a violation for each changeset without preconditions
3. Apply exclude patterns if configured

### Parser support

- **XML**: Parse `<preConditions>` element
- **YAML**: Parse `preConditions` key
- **SQL**: Parse `--preconditions` or `--precondition-*` comments

### Testing considerations

- Test with various precondition types (tableExists, columnExists, sqlCheck, etc.)
- Test with empty precondition blocks
- Test with exclude patterns
- Test across all changelog formats (XML, YAML, SQL)

## Examples from your codebase

**Good example** (from your SQL files):
```sql
--changeset maxime.charles:330
--preconditions onFail:MARK_RAN
--precondition-sql-check expectedResult:0 SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_SCHEMA = 'dbo' AND TABLE_NAME = 'BankTransactions'
CREATE TABLE dbo.BankTransactions (...)
```

This changeset properly checks that the table doesn't exist before creating it.
