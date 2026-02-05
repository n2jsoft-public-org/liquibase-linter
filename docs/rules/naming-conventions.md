# naming-conventions

**Severity**: Info  
**Category**: Best Practices  
**Status**: ✅ Implemented

## Description

Enforces consistent naming conventions for database objects.

## What it checks

- Table names (lowercase, snake_case)
- Column names (consistent format)
- Index naming patterns (e.g., `idx_tablename_columnname`)
- Constraint naming patterns (e.g., `fk_table1_table2`, `uk_table_column`)

## Example violation

**XML format:**
```xml
<changeSet id="1" author="test">
    <createTable tableName="UserAccounts">
        <column name="userId" type="INT"/>
        <column name="UserName" type="VARCHAR(100)"/>
    </createTable>
</changeSet>
```

**SQL format:**
```sql
--liquibase formatted sql
--changeset test:1
CREATE TABLE UserAccounts (
    userId INT,
    UserName VARCHAR(100)
);

--changeset test:2
CREATE INDEX IDX_UserAccounts_UserId ON UserAccounts(userId);
```

## Correct usage

**XML format:**
```xml
<changeSet id="1" author="test">
    <createTable tableName="user_accounts">
        <column name="user_id" type="INT"/>
        <column name="user_name" type="VARCHAR(100)"/>
    </createTable>
</changeSet>
```

**SQL format:**
```sql
--liquibase formatted sql
--changeset test:1
CREATE TABLE user_accounts (
    user_id INT,
    user_name VARCHAR(100)
);

--changeset test:2
CREATE INDEX idx_user_accounts_user_id ON user_accounts(user_id);
```

## Configuration

```yaml
rules:
  naming-conventions:
    enabled: true
    severity: info
```

## Why this matters

Consistent naming conventions:
- Improve code readability
- Reduce confusion and errors
- Make the database schema more maintainable
- Follow industry best practices
- Make automated tooling easier

Choose a convention that works for your team and stick to it across all database objects.
