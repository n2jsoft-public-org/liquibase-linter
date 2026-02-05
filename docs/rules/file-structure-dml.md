# file-structure-dml

**Severity**: Critical  
**Category**: Best Practices  
**Status**: ✅ Implemented

## Description

Enforces DML (Data Manipulation Language) changes to be placed in data directories.

## What it detects

- DML operations (INSERT, UPDATE, DELETE) in structure directories
- Data modifications mixed with schema changes
- Data changes in wrong directory hierarchy

## DML Operations Detected

- `insert` - Insert new rows into tables
- `update` - Update existing rows
- `delete` - Delete rows from tables
- `loadData` - Load data from CSV/XML files
- `loadUpdateData` - Load or update data from files

## Example violation

**XML format:**
```xml
<!-- File: db/changelog/sprints/v116/0 - structure/seed.xml -->
<changeSet id="1" author="dev">
    <insert tableName="users">  <!-- ❌ DML in structure directory -->
        <column name="username" value="admin"/>
        <column name="email" value="admin@example.com"/>
    </insert>
</changeSet>
```

**SQL format:**
```sql
-- File: db/changelog/sprints/v116/0 - structure/seed.sql
--liquibase formatted sql
--changeset dev:1
INSERT INTO users (username, email)   -- ❌ DML in structure directory
VALUES ('admin', 'admin@example.com');

--changeset dev:2
UPDATE settings SET value = '1.0' WHERE key = 'version';  -- ❌ DML in structure directory
```

## Correct usage

**XML format:**
```xml
<!-- File: db/changelog/sprints/v116/1 - data/seed.xml -->
<changeSet id="1" author="dev">
    <insert tableName="users">  <!-- ✅ DML in data directory -->
        <column name="username" value="admin"/>
        <column name="email" value="admin@example.com"/>
    </insert>
</changeSet>
```

**SQL format:**
```sql
-- File: db/changelog/sprints/v116/1 - data/seed.sql
--liquibase formatted sql
--changeset dev:1
INSERT INTO users (username, email)   -- ✅ DML in data directory
VALUES ('admin', 'admin@example.com');

--changeset dev:2
UPDATE settings SET value = '1.0' WHERE key = 'version';  -- ✅ DML in data directory
```

## Configuration

```yaml
file_structure:
  sprint_pattern: "^v\\d+$"
  data_pattern: "^\\d+ - data$"  # Matches "0 - data", "1 - data"
  exclude_patterns:
    - "**/init/**"
```

### Custom Data Patterns

Customize the pattern to match your organization:

```yaml
file_structure:
  # Simple naming without numbers
  data_pattern: "^data$"
  
  # Alternative naming
  data_pattern: "^dml$"
  data_pattern: "^seeds$"
  data_pattern: "^migrations$"
```

## Why this matters

Separating DML from DDL provides:
- **Clear separation of concerns**: Data changes vs schema changes
- **Easier data migrations**: Can identify and review data modifications separately
- **Better control**: Data changes often require different approval processes
- **Improved testing**: Can test data migrations separately from schema changes
- **Risk management**: Data modifications need different rollback strategies

## Common Use Cases

### Reference Data
```xml
<!-- data/reference_data.xml -->
<changeSet id="1" author="dev">
    <insert tableName="countries">
        <column name="code" value="US"/>
        <column name="name" value="United States"/>
    </insert>
</changeSet>
```

### Data Fixes
```xml
<!-- data/fix_user_emails.xml -->
<changeSet id="1" author="dev">
    <update tableName="users">
        <column name="email" value="corrected@example.com"/>
        <where>id = 123</where>
    </update>
</changeSet>
```

### Data Migration
```xml
<!-- data/migrate_old_format.xml -->
<changeSet id="1" author="dev">
    <sql>
        UPDATE products 
        SET new_category_id = old_category_id 
        WHERE new_category_id IS NULL;
    </sql>
</changeSet>
```

## Important Notes

**Generic SQL**: The `sql` change type is allowed in both directories since its content cannot be automatically classified as DDL or DML. Use specific Liquibase change types when possible.

**Data vs Structure**: If unsure whether a change is DDL or DML:
- **DDL**: Changes the schema/structure (CREATE, ALTER, DROP)
- **DML**: Changes the data (INSERT, UPDATE, DELETE, LOAD)

## Benefits

- Clear identification of data migrations
- Better control over data seeding and updates
- Reduced risk of accidental data modifications during schema changes
- Easier to coordinate with application deployments
- Improved visibility into what data is being changed

## See Also

- [file-structure-sprint](file-structure-sprint.md) - Enforces sprint-based organization
- [file-structure-ddl](file-structure-ddl.md) - Ensures DDL changes in structure directories
