# non-idempotent

**Severity**: Warning  
**Category**: Reliability  
**Status**: ✅ Implemented

## Description

Detects operations that are not idempotent and may fail on re-run.

## What it detects

- CREATE TABLE without IF NOT EXISTS checks
- INSERT without preconditions
- ALTER TABLE ADD COLUMN without existence checks

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

```yaml
rules:
  non-idempotent:
    enabled: true
    severity: warning
```

## Why this matters

Non-idempotent changes can:
- Fail when deployment is retried
- Cause deployment pipeline failures
- Make testing and development more difficult
- Lead to inconsistent database states

Idempotent changes can be safely re-run multiple times with the same result, which is crucial for reliable deployments.
