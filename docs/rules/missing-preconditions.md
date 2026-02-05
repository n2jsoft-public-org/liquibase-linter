# missing-preconditions

**Severity**: Warning  
**Category**: Reliability  
**Status**: ⚠️ Merged into [non-idempotent](non-idempotent.md)

> **Note**: This rule was documented but never actually implemented. Its intended functionality is available in the `non-idempotent` rule with `mode: "risky-only"` (the default behavior).

## Description

Ensures risky operations have appropriate preconditions.

## What it detects

- ALTER TABLE without column existence checks
- DROP operations without existence checks
- Data manipulation without validation

## Example violation

```xml
<changeSet id="1" author="test">
    <addColumn tableName="users">
        <column name="email" type="VARCHAR(255)"/>
    </addColumn>
</changeSet>
```

## Correct usage

```xml
<changeSet id="1" author="test">
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

## Configuration

This functionality is now available in the `non-idempotent` rule. Use:

```yaml
rules:
  non-idempotent:
    enabled: true
    severity: warning
    mode: "risky-only"  # Default: checks risky operations only
```

For detailed configuration options, see [non-idempotent rule documentation](non-idempotent.md).

## Why this matters

Preconditions help:
- Prevent errors from changes applied to unexpected database states
- Make changesets safer to run in different environments
- Document assumptions about database state
- Enable better error handling with onFail attributes

Preconditions act as a safety net, ensuring changes only execute when the database is in the expected state.

For complete information on why preconditions matter and how to use them effectively, see the [non-idempotent rule documentation](non-idempotent.md#why-this-matters).
