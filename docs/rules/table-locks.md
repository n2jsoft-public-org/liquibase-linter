# table-locks

**Severity**: Warning  
**Category**: Performance  
**Status**: ✅ Implemented

## Description

Identifies operations that may cause prolonged table locks and impact application availability.

## What it detects

- ALTER TABLE on large tables without online options
- Rebuilding indexes on production tables
- Adding NOT NULL constraints without defaults

## Example violation

**XML format:**
```xml
<changeSet id="1" author="test">
    <addColumn tableName="large_table">
        <column name="status" type="VARCHAR(50)">
            <constraints nullable="false"/>
        </column>
    </addColumn>
</changeSet>
```

**SQL format:**
```sql
--liquibase formatted sql
--changeset test:1
ALTER TABLE large_table ADD COLUMN status VARCHAR(50) NOT NULL;

--changeset test:2
ALTER TABLE large_table ADD INDEX idx_name (column_name);
```

## Correct usage

Add the column as nullable first, populate it, then add the constraint:

**XML format:**
```xml
<changeSet id="1" author="test">
    <addColumn tableName="large_table">
        <column name="status" type="VARCHAR(50)" defaultValue="active"/>
    </addColumn>
</changeSet>

<changeSet id="2" author="test">
    <addNotNullConstraint tableName="large_table" columnName="status"/>
</changeSet>
```

**SQL format:**
```sql
--liquibase formatted sql
--changeset test:1
ALTER TABLE large_table ADD COLUMN status VARCHAR(50) DEFAULT 'active';

--changeset test:2
UPDATE large_table SET status = 'active' WHERE status IS NULL;

--changeset test:3
ALTER TABLE large_table MODIFY COLUMN status VARCHAR(50) NOT NULL;
```

## Configuration

```yaml
rules:
  table-locks:
    enabled: true
    severity: warning
```

## Why this matters

Long-running table locks can:
- Block application queries and transactions
- Cause timeouts and errors
- Degrade user experience
- Lead to deployment failures in production

Consider using online schema change tools or breaking changes into smaller, less-blocking operations for large production tables.
