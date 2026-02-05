# Rules Reference

This document describes all available linting rules in the Liquibase Linter.

## Rule Categories

- **Security**: Detect security vulnerabilities and risks
- **Performance**: Identify performance issues
- **Reliability**: Ensure reliable database migrations
- **Best Practices**: Enforce coding standards and conventions

## Security Rules

### sql-injection

**Severity**: Critical  
**Category**: Security  
**Status**: ✅ Implemented

Detects potential SQL injection vulnerabilities in Liquibase changesets.

**What it detects**:
- String concatenation in SQL statements
- Unescaped variable substitution
- Dynamic SQL construction without parameterization

**Example violation**:
```xml
<changeSet id="1" author="test">
    <sql>
        DELETE FROM users WHERE username = '${username}';
    </sql>
</changeSet>
```

**Correct usage**:
```xml
<changeSet id="1" author="test">
    <delete tableName="users">
        <where>username = :username</where>
    </delete>
</changeSet>
```

### hardcoded-credentials

**Severity**: Critical  
**Category**: Security  
**Status**: ✅ Implemented

Finds hardcoded passwords, API keys, and other sensitive credentials.

**What it detects**:
- Hardcoded passwords in CREATE USER statements
- API keys and tokens in plain text
- Connection strings with embedded credentials

**Example violation**:
```xml
<changeSet id="1" author="test">
    <sql>
        CREATE USER 'app_user'@'localhost' IDENTIFIED BY 'password123';
    </sql>
</changeSet>
```

### dangerous-operations

**Severity**: Critical  
**Category**: Security  
**Status**: ✅ Implemented

Detects dangerous database operations that could cause data loss.

**What it detects**:
- DROP TABLE without preconditions
- TRUNCATE TABLE without safety checks
- DROP COLUMN operations
- DROP DATABASE statements

**Example violation**:
```xml
<changeSet id="1" author="test">
    <dropTable tableName="important_data"/>
</changeSet>
```

**Correct usage**:
```xml
<changeSet id="1" author="test">
    <preConditions onFail="MARK_RAN">
        <tableExists tableName="important_data"/>
    </preConditions>
    <dropTable tableName="important_data"/>
</changeSet>
```

### privilege-escalation

**Severity**: Critical  
**Category**: Security  
**Status**: ✅ Implemented

Detects excessive privilege grants that could lead to security issues.

**What it detects**:
- GRANT ALL PRIVILEGES
- Creating superusers
- Granting permissions to wildcard users ('%')

## Reliability Rules

### missing-rollback

**Severity**: Warning  
**Category**: Reliability  
**Status**: ✅ Implemented

Ensures changesets have proper rollback scripts.

**What it detects**:
- Changesets without rollback definitions
- Empty rollback blocks
- Rollback marked as not supported

**Example violation**:
```xml
<changeSet id="1" author="test">
    <createTable tableName="users">
        <!-- table definition -->
    </createTable>
</changeSet>
```

**Correct usage**:
```xml
<changeSet id="1" author="test">
    <createTable tableName="users">
        <!-- table definition -->
    </createTable>
    <rollback>
        <dropTable tableName="users"/>
    </rollback>
</changeSet>
```

### non-idempotent

**Severity**: Warning  
**Category**: Reliability  
**Status**: ✅ Implemented

Detects operations that are not idempotent and may fail on re-run.

**What it detects**:
- CREATE TABLE without IF NOT EXISTS checks
- INSERT without preconditions
- ALTER TABLE ADD COLUMN without existence checks

### missing-preconditions

**Severity**: Warning  
**Category**: Reliability  
**Status**: ✅ Implemented

Ensures risky operations have appropriate preconditions.

**What it detects**:
- ALTER TABLE without column existence checks
- DROP operations without existence checks
- Data manipulation without validation

## Performance Rules

### missing-indexes

**Severity**: Info  
**Category**: Performance  
**Status**: ✅ Implemented

Detects tables without proper indexes on foreign keys.

**What it detects**:
- Foreign key columns without indexes
- Large tables without any indexes
- Missing indexes on frequently queried columns

### table-locks

**Severity**: Warning  
**Category**: Performance  
**Status**: ✅ Implemented

Identifies operations that may cause prolonged table locks.

**What it detects**:
- ALTER TABLE on large tables without online options
- Rebuilding indexes on production tables
- Adding NOT NULL constraints without defaults

### large-data-operations

**Severity**: Warning  
**Category**: Performance  
**Status**: ✅ Implemented

Detects operations that manipulate large amounts of data.

**What it detects**:
- Unbounded UPDATE/DELETE statements
- Large batch inserts
- Full table scans

## Best Practices Rules

### naming-conventions

**Severity**: Info  
**Category**: Best Practices  
**Status**: ✅ Implemented

Enforces consistent naming conventions for database objects.

**What it checks**:
- Table names (lowercase, snake_case)
- Column names (consistent format)
- Index naming patterns
- Constraint naming patterns

### changelog-organization

**Severity**: Info  
**Category**: Best Practices  
**Status**: ✅ Implemented

Ensures proper organization of changelog files.

**What it checks**:
- ChangeSet IDs are sequential
- Author information is present
- Contexts are used appropriately
- Labels are meaningful

### documentation

**Severity**: Info  
**Category**: Best Practices  
**Status**: ✅ Implemented

Ensures changesets are properly documented.

**What it checks**:
- Comments describing the changeset purpose
- JIRA ticket references
- Complex changes have explanations

## Configuring Rules

Rules can be enabled/disabled and their severity adjusted in the configuration file:

```yaml
rules:
  sql-injection:
    enabled: true
    severity: critical
  
  missing-rollback:
    enabled: true
    severity: warning
  
  naming-conventions:
    enabled: false  # Disable this rule
```

## Rule Severity Levels

- **Critical**: Security vulnerabilities and data loss risks
- **Warning**: Reliability issues and potential problems
- **Info**: Best practice violations and code quality issues

## Next Steps

- Configure rules in [configuration.md](configuration.md)
- Set up [CI/CD integration](cicd.md)
