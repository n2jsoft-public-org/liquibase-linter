# No IF EXISTS (no-if-exists)

**Category**: Best Practice  
**Severity**: Warning (default)  
**Since**: v1.1.0

## What It Detects

This rule detects the use of database-specific conditional existence checks (IF EXISTS, IF NOT EXISTS, IF OBJECT_ID) in SQL scripts and recommends using Liquibase preconditions instead.

## Why It Matters

Using database-specific IF EXISTS patterns has several drawbacks:

1. **Database Portability**: IF EXISTS syntax varies significantly across database vendors (SQL Server, PostgreSQL, MySQL, Oracle), making migrations harder
2. **Liquibase Integration**: Liquibase preconditions are the standard way to check object existence and integrate better with Liquibase's change tracking
3. **Idempotency**: Liquibase preconditions with `onFail="MARK_RAN"` provide cleaner idempotent behavior
4. **Error Handling**: Preconditions offer better control over what happens when conditions aren't met (HALT, CONTINUE, MARK_RAN, WARN)
5. **Cross-Database Support**: Liquibase preconditions work consistently across all supported databases

## Examples

### ❌ Incorrect (SQL Server)

```sql
--liquibase formatted sql
--changeset john.doe:add-user-email-column

IF NOT EXISTS (
    SELECT * FROM INFORMATION_SCHEMA.COLUMNS 
    WHERE TABLE_NAME = 'users' AND COLUMN_NAME = 'email'
)
BEGIN
    ALTER TABLE users ADD email VARCHAR(255);
END;
```

### ❌ Incorrect (PostgreSQL)

```sql
--liquibase formatted sql
--changeset john.doe:create-users-table

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_tables WHERE tablename = 'users') THEN
        CREATE TABLE users (
            id SERIAL PRIMARY KEY,
            name VARCHAR(100)
        );
    END IF;
END $$;
```

### ❌ Incorrect (MySQL)

```sql
--liquibase formatted sql
--changeset john.doe:create-index

DROP PROCEDURE IF EXISTS create_index_if_not_exists;
DELIMITER //
CREATE PROCEDURE create_index_if_not_exists()
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.statistics 
        WHERE table_name = 'users' AND index_name = 'idx_email'
    ) THEN
        CREATE INDEX idx_email ON users(email);
    END IF;
END//
DELIMITER ;
CALL create_index_if_not_exists();
DROP PROCEDURE create_index_if_not_exists;
```

### ✅ Correct (Using Liquibase Preconditions)

```xml
<changeSet id="add-user-email-column" author="john.doe">
    <preConditions onFail="MARK_RAN">
        <not>
            <columnExists tableName="users" columnName="email"/>
        </not>
    </preConditions>
    <addColumn tableName="users">
        <column name="email" type="varchar(255)"/>
    </addColumn>
</changeSet>
```

```xml
<changeSet id="create-users-table" author="john.doe">
    <preConditions onFail="MARK_RAN">
        <not>
            <tableExists tableName="users"/>
        </not>
    </preConditions>
    <createTable tableName="users">
        <column name="id" type="bigint" autoIncrement="true">
            <constraints primaryKey="true"/>
        </column>
        <column name="name" type="varchar(100)"/>
    </createTable>
</changeSet>
```

```xml
<changeSet id="create-index" author="john.doe">
    <preConditions onFail="MARK_RAN">
        <not>
            <indexExists tableName="users" indexName="idx_email"/>
        </not>
    </preConditions>
    <createIndex tableName="users" indexName="idx_email">
        <column name="email"/>
    </createIndex>
</changeSet>
```

## Configuration

```yaml
rules:
  no-if-exists:
    enabled: true
    severity: warning  # info, warning, error, critical
    exclude-patterns:
      - "**/init/**"              # Exclude initialization scripts
      - "**/migration-legacy/**"  # Exclude legacy migrations
```

### Configuration Options

- **enabled**: Enable or disable this rule (default: `true`)
- **severity**: Violation severity level (default: `warning`)
- **exclude-patterns**: Glob patterns for files to exclude from this check (optional)

## Common Patterns Detected

The rule detects various IF EXISTS patterns across different SQL dialects:

### SQL Server
- `IF EXISTS (SELECT ...)`
- `IF NOT EXISTS (SELECT ...)`
- `IF OBJECT_ID('table_name', 'U') IS NOT NULL`
- `IF OBJECT_ID('proc_name', 'P') IS NOT NULL`

### PostgreSQL
- `DO $$ BEGIN IF EXISTS (...) THEN ... END IF; END $$;`
- `DO $$ BEGIN IF NOT EXISTS (...) THEN ... END IF; END $$;`
- `SELECT ... WHERE EXISTS (...)`

### MySQL
- `DROP PROCEDURE IF EXISTS`
- `CREATE PROCEDURE ... IF NOT EXISTS (...)`
- `IF EXISTS (SELECT ... FROM information_schema ...)`

## Migration Guide

### Step 1: Identify the Object Type
Determine what type of database object you're checking:
- Table → `<tableExists>`
- Column → `<columnExists>`
- Index → `<indexExists>`
- View → `<viewExists>`
- Sequence → `<sequenceExists>`

### Step 2: Choose the Appropriate Precondition
Use Liquibase's built-in preconditions:
```xml
<preConditions onFail="MARK_RAN">
    <not>
        <tableExists tableName="your_table"/>
    </not>
</preConditions>
```

### Step 3: Set Failure Behavior
- `onFail="MARK_RAN"`: Mark changeset as executed without running (most common for IF NOT EXISTS)
- `onFail="CONTINUE"`: Skip this changeset and continue
- `onFail="HALT"`: Stop execution (default, used for required conditions)

### Step 4: Convert SQL to Liquibase Changes
If possible, replace raw SQL with Liquibase change types:
- `ALTER TABLE ADD COLUMN` → `<addColumn>`
- `CREATE TABLE` → `<createTable>`
- `CREATE INDEX` → `<createIndex>`

If SQL is necessary, wrap it with appropriate preconditions:
```xml
<changeSet id="custom-change" author="john.doe">
    <preConditions onFail="MARK_RAN">
        <not>
            <tableExists tableName="audit_log"/>
        </not>
    </preConditions>
    <sql>
        CREATE TABLE audit_log (
            id BIGINT PRIMARY KEY,
            event_type VARCHAR(50),
            created_at TIMESTAMP
        );
    </sql>
</changeSet>
```

## See Also

- [Mandatory Preconditions](mandatory-preconditions.md) - Rule requiring preconditions on all changesets
- [Non-Idempotent Changes](non-idempotent.md) - Rule detecting changes that can't be safely re-run
- [Liquibase Preconditions Documentation](https://docs.liquibase.com/concepts/changelogs/preconditions.html)
