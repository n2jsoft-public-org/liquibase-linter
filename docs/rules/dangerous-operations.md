# dangerous-operations

**Severity**: Critical  
**Category**: Security  
**Status**: ✅ Implemented

## Description

Detects dangerous database operations that could cause data loss.

## What it detects

- DROP TABLE without preconditions
- TRUNCATE TABLE without safety checks
- DROP COLUMN operations
- DROP DATABASE statements

## Example violation

**XML format:**
```xml
<changeSet id="1" author="test">
    <dropTable tableName="important_data"/>
</changeSet>
```

**SQL format:**
```sql
--liquibase formatted sql
--changeset test:1
DROP TABLE important_data;

--changeset test:2
TRUNCATE TABLE user_logs;
```

## Correct usage

**XML format:**
```xml
<changeSet id="1" author="test">
    <preConditions onFail="MARK_RAN">
        <tableExists tableName="important_data"/>
    </preConditions>
    <dropTable tableName="important_data"/>
</changeSet>
```

**SQL format:**
```sql
--liquibase formatted sql
--changeset test:1
--preconditions onFail:MARK_RAN
--precondition-sql-check expectedResult:1 SELECT COUNT(*) FROM information_schema.tables WHERE table_name='important_data'
DROP TABLE important_data;
```

## Configuration

```yaml
rules:
  dangerous-operations:
    enabled: true
    severity: critical
```

## Why this matters

Dangerous operations without proper safety checks can:
- Cause irreversible data loss
- Break existing applications
- Result in costly downtime
- Require complex recovery procedures

Always use preconditions to ensure dangerous operations only execute when appropriate, and ensure you have proper backups before running destructive changes.
