# large-data-operations

**Severity**: Warning  
**Category**: Performance  
**Status**: ✅ Implemented

## Description

Detects operations that manipulate large amounts of data without proper safeguards.

## What it detects

- Unbounded UPDATE/DELETE statements
- Large batch inserts
- Full table scans

## Example violation

**XML format:**
```xml
<changeSet id="1" author="test">
    <sql>
        UPDATE users SET status = 'inactive';
    </sql>
</changeSet>
```

**SQL format:**
```sql
--liquibase formatted sql
--changeset test:1
UPDATE users SET status = 'inactive';

--changeset test:2
DELETE FROM audit_logs;

--changeset test:3
INSERT INTO archive SELECT * FROM transactions;
```

## Correct usage

Use conditions to limit scope:

**XML format:**
```xml
<changeSet id="1" author="test">
    <sql>
        UPDATE users 
        SET status = 'inactive' 
        WHERE last_login < DATE_SUB(NOW(), INTERVAL 1 YEAR);
    </sql>
</changeSet>
```

**SQL format:**
```sql
--liquibase formatted sql
--changeset test:1
UPDATE users 
SET status = 'inactive' 
WHERE last_login < DATE_SUB(NOW(), INTERVAL 1 YEAR);

--changeset test:2
DELETE FROM audit_logs 
WHERE created_at < DATE_SUB(NOW(), INTERVAL 90 DAY)
LIMIT 10000;
```

Or consider breaking into batches for very large operations.

## Configuration

```yaml
rules:
  large-data-operations:
    enabled: true
    severity: warning
```

## Why this matters

Large data operations can:
- Lock tables for extended periods
- Consume excessive database resources
- Cause transaction log growth
- Lead to deployment timeouts
- Impact application performance

For large data migrations, consider using batch processing or running operations outside of deployment windows.
