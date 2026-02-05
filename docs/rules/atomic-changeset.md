# atomic-changeset

**Severity**: Info (default)  
**Category**: Best Practice  
**Status**: 📋 Planned

## Description

Enforces that each changeset contains only a single change operation. This promotes atomicity, makes rollbacks cleaner, improves readability, and makes it easier to track which specific change caused issues.

## What it detects

- Changesets with multiple change operations (e.g., createTable + createIndex)
- Changesets with multiple SQL statements performing different operations
- Changesets mixing DDL and DML operations

## Why this matters

### Atomicity and Rollback Clarity
When a changeset contains multiple changes, rolling back becomes more complex. If one operation succeeds and another fails, the database may be left in an inconsistent state.

### Easier Debugging
Single-change changesets make it immediately clear which specific operation caused an issue during deployment.

### Better Change Tracking
Each change gets its own entry in DATABASECHANGELOG, providing clearer audit trails and making it easier to understand what changed when.

### Cleaner History
Version control history becomes more granular and meaningful when each commit addresses a single, focused database change.

### Improved Readability
Changesets with a single purpose are easier to understand and review.

## Example violations

### ❌ Incorrect (Multiple Changes)

**XML format:**
```xml
<changeSet id="1" author="john.doe">
    <createTable tableName="users">
        <column name="id" type="bigint" autoIncrement="true">
            <constraints primaryKey="true"/>
        </column>
        <column name="email" type="varchar(255)"/>
    </createTable>
    <createIndex tableName="users" indexName="idx_users_email">
        <column name="email"/>
    </createIndex>
    <addNotNullConstraint tableName="users" columnName="email"/>
</changeSet>
```

**SQL format:**
```sql
--changeset john.doe:1
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(255)
);

CREATE INDEX idx_users_email ON users(email);

ALTER TABLE users MODIFY email VARCHAR(255) NOT NULL;
```

### ❌ Incorrect (Mixing DDL and DML)

```xml
<changeSet id="2" author="john.doe">
    <createTable tableName="settings">
        <column name="key" type="varchar(50)"/>
        <column name="value" type="varchar(255)"/>
    </createTable>
    <insert tableName="settings">
        <column name="key" value="app_version"/>
        <column name="value" value="1.0.0"/>
    </insert>
</changeSet>
```

### ✅ Correct (Atomic Changesets)

**Split into separate changesets:**
```xml
<changeSet id="1" author="john.doe">
    <createTable tableName="users">
        <column name="id" type="bigint" autoIncrement="true">
            <constraints primaryKey="true"/>
        </column>
        <column name="email" type="varchar(255)"/>
    </createTable>
</changeSet>

<changeSet id="2" author="john.doe">
    <createIndex tableName="users" indexName="idx_users_email">
        <column name="email"/>
    </createIndex>
</changeSet>

<changeSet id="3" author="john.doe">
    <addNotNullConstraint tableName="users" columnName="email"/>
</changeSet>
```

**SQL format:**
```sql
--changeset john.doe:1
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    email VARCHAR(255)
);

--changeset john.doe:2
CREATE INDEX idx_users_email ON users(email);

--changeset john.doe:3
ALTER TABLE users MODIFY email VARCHAR(255) NOT NULL;
```

## Configuration

```yaml
rules:
  atomic-changeset:
    enabled: true
    severity: info  # info, warning, error, critical
    allow-table-with-indexes: false  # Allow table creation with indexes as single changeset
    allow-table-with-constraints: true  # Allow table creation with inline constraints
    max-sql-statements: 1  # Maximum SQL statements in a single changeset
    exclude-patterns:
      - "**/init/**"  # Exclude initialization scripts
```

### Configuration Options

- **enabled**: Enable or disable this rule (default: `true`)
- **severity**: Violation severity level (default: `info`)
- **allow-table-with-indexes**: Allow createTable with separate createIndex in same changeset (default: `false`)
- **allow-table-with-constraints**: Allow table creation with inline constraints like primary keys, foreign keys (default: `true`)
- **max-sql-statements**: Maximum number of SQL statements allowed in SQL changes (default: `1`)
- **exclude-patterns**: Glob patterns for files to exclude (optional)

## Exceptions

Some scenarios may legitimately benefit from multiple changes in one changeset:

### Table Creation with Inline Constraints
Creating a table with its primary key, foreign keys, and constraints is typically acceptable as a single logical operation:

```xml
<changeSet id="1" author="john.doe">
    <createTable tableName="users">
        <column name="id" type="bigint" autoIncrement="true">
            <constraints primaryKey="true"/>
        </column>
        <column name="org_id" type="bigint">
            <constraints foreignKeyName="fk_users_org" references="organizations(id)"/>
        </column>
        <column name="email" type="varchar(255)">
            <constraints nullable="false" unique="true"/>
        </column>
    </createTable>
</changeSet>
```

This is controlled by `allow-table-with-constraints: true` (default).

### Initialization Scripts
Initial schema setup scripts in `init/` directories often combine multiple operations for convenience:

```yaml
rules:
  atomic-changeset:
    exclude-patterns:
      - "**/init/**"
```

### Tightly Coupled Operations
In rare cases, operations that must succeed or fail together (e.g., renaming a column and updating all references to it) might be better kept in a single changeset for transactional consistency.

## Related Rules

- [non-idempotent](non-idempotent.md): Ensures changes can be safely re-run
- [missing-rollback](missing-rollback.md): Ensures changesets have rollback instructions
- [documentation](documentation.md): Ensures changesets are documented
- [changelog-organization](changelog-organization.md): Enforces file organization patterns

## Migration Strategy

When adopting this rule on an existing project:

1. **Start with `severity: info`** to identify violations without blocking builds
2. **Review and categorize** violations - some may be legitimate multi-step operations
3. **Add exclude patterns** for init scripts and legacy migrations
4. **Gradually refactor** new changes to follow the single-change pattern
5. **Escalate to `warning`** once team is comfortable with the practice
6. **Consider `error`** for strict enforcement on new projects

## Examples from Common Patterns

### Pattern: Create Table + Index
**Instead of:**
```sql
--changeset dev:create-users-with-index
CREATE TABLE users (id BIGINT PRIMARY KEY, email VARCHAR(255));
CREATE INDEX idx_users_email ON users(email);
```

**Use:**
```sql
--changeset dev:create-users
CREATE TABLE users (id BIGINT PRIMARY KEY, email VARCHAR(255));

--changeset dev:index-users-email
CREATE INDEX idx_users_email ON users(email);
```

### Pattern: Alter Table Multiple Columns
**Instead of:**
```sql
--changeset dev:modify-users
ALTER TABLE users ADD COLUMN phone VARCHAR(20);
ALTER TABLE users ADD COLUMN address TEXT;
ALTER TABLE users ADD COLUMN city VARCHAR(100);
```

**Use:**
```sql
--changeset dev:add-user-phone
ALTER TABLE users ADD COLUMN phone VARCHAR(20);

--changeset dev:add-user-address
ALTER TABLE users ADD COLUMN address TEXT;

--changeset dev:add-user-city
ALTER TABLE users ADD COLUMN city VARCHAR(100);
```

### Pattern: Create + Populate Table
**Instead of:**
```sql
--changeset dev:setup-settings
CREATE TABLE settings (key VARCHAR(50), value VARCHAR(255));
INSERT INTO settings VALUES ('version', '1.0');
```

**Use:**
```sql
--changeset dev:create-settings
CREATE TABLE settings (key VARCHAR(50), value VARCHAR(255));

--changeset dev:populate-settings
INSERT INTO settings VALUES ('version', '1.0');
```

## See Also

- [Liquibase Best Practices](https://docs.liquibase.com/concepts/bestpractices.html)
- [Atomic Database Changes Pattern](https://martinfowler.com/articles/evodb.html)
