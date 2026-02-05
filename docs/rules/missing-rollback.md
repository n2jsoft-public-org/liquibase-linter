# missing-rollback

**Severity**: Warning  
**Category**: Reliability  
**Status**: ✅ Implemented

## Description

Ensures changesets have proper rollback scripts.

## What it detects

- Changesets without rollback definitions
- Empty rollback blocks
- Rollback marked as not supported

## Example violation

```xml
<changeSet id="1" author="test">
    <createTable tableName="users">
        <column name="id" type="INT"/>
    </createTable>
</changeSet>
```

## Correct usage

```xml
<changeSet id="1" author="test">
    <createTable tableName="users">
        <column name="id" type="INT"/>
    </createTable>
    <rollback>
        <dropTable tableName="users"/>
    </rollback>
</changeSet>
```

## Configuration

```yaml
rules:
  missing-rollback:
    enabled: true
    severity: warning
```

## Why this matters

Rollback scripts are essential for:
- Quick recovery from failed deployments
- Testing deployment procedures safely
- Rolling back problematic changes
- Maintaining database version control

While Liquibase can auto-generate rollbacks for some operations, explicit rollback definitions ensure correct behavior and demonstrate that rollback scenarios have been considered.
