# file-structure-ddl

**Severity**: Critical  
**Category**: Best Practices  
**Status**: ✅ Implemented

## Description

Enforces DDL (Data Definition Language) changes to be placed in structure directories.

## What it detects

- DDL operations (CREATE TABLE, ALTER TABLE, CREATE INDEX, etc.) in data directories
- Schema changes mixed with data modifications
- Structural changes in wrong directory hierarchy

## DDL Operations Detected

### Table Operations
- `createTable` - Create new table
- `dropTable` - Drop existing table
- `alterTable` - Alter table structure
- `renameTable` - Rename table

### Column Operations
- `addColumn` - Add new column
- `dropColumn` - Drop existing column
- `modifyColumn` - Modify column definition
- `renameColumn` - Rename column

### Index Operations
- `createIndex` - Create index
- `dropIndex` - Drop index

### Constraint Operations
- `addPrimaryKey` - Add primary key
- `addForeignKeyConstraint` - Add foreign key
- `addUniqueConstraint` - Add unique constraint
- `dropPrimaryKey` - Drop primary key
- `dropForeignKeyConstraint` - Drop foreign key
- `dropUniqueConstraint` - Drop unique constraint

### Other DDL Operations
- `createView`, `dropView` - View management
- `createProcedure`, `dropProcedure` - Stored procedure management
- `createSequence`, `alterSequence`, `dropSequence` - Sequence management
- `grant`, `revoke` - Privilege management
- `createUser`, `dropUser` - User management

## Example violation

**XML format:**
```xml
<!-- File: db/changelog/sprints/v116/1 - data/tables.xml -->
<changeSet id="1" author="dev">
    <createTable tableName="users">  <!-- ❌ DDL in data directory -->
        <column name="id" type="INT"/>
        <column name="username" type="VARCHAR(100)"/>
    </createTable>
</changeSet>
```

**SQL format:**
```sql
-- File: db/changelog/sprints/v116/1 - data/tables.sql
--liquibase formatted sql
--changeset dev:1
CREATE TABLE users (  -- ❌ DDL in data directory
    id INT,
    username VARCHAR(100)
);

--changeset dev:2
ALTER TABLE orders ADD COLUMN status VARCHAR(50);  -- ❌ DDL in data directory
```

## Correct usage

**XML format:**
```xml
<!-- File: db/changelog/sprints/v116/0 - structure/tables.xml -->
<changeSet id="1" author="dev">
    <createTable tableName="users">  <!-- ✅ DDL in structure directory -->
        <column name="id" type="INT"/>
        <column name="username" type="VARCHAR(100)"/>
    </createTable>
</changeSet>
```

**SQL format:**
```sql
-- File: db/changelog/sprints/v116/0 - structure/tables.sql
--liquibase formatted sql
--changeset dev:1
CREATE TABLE users (  -- ✅ DDL in structure directory
    id INT,
    username VARCHAR(100)
);

--changeset dev:2
ALTER TABLE orders ADD COLUMN status VARCHAR(50);  -- ✅ DDL in structure directory
```

## Configuration

```yaml
file_structure:
  sprint_pattern: "^v\\d+$"
  structure_pattern: "^\\d+ - structure$"  # Matches "0 - structure", "1 - structure"
  exclude_patterns:
    - "**/init/**"
```

### Custom Structure Patterns

Customize the pattern to match your organization:

```yaml
file_structure:
  # Simple naming without numbers
  structure_pattern: "^structure$"
  
  # Alternative naming
  structure_pattern: "^schema$"
  structure_pattern: "^ddl$"
```

## Why this matters

Separating DDL from DML provides:
- **Clear separation of concerns**: Schema changes vs data changes
- **Easier code review**: Reviewers can focus on structural vs data changes
- **Better risk assessment**: DDL changes typically have higher impact
- **Simplified deployment**: Can apply schema changes before data changes
- **Improved testing**: Can test schema migrations separately from data migrations

## Important Notes

**Generic SQL**: The `sql` change type is allowed in both directories since its content cannot be automatically classified as DDL or DML. Use specific Liquibase change types when possible.

## Benefits

- Clear organization for database architects and DBAs
- Reduced risk of data loss during schema changes
- Better understanding of migration complexity
- Easier to identify breaking changes

## See Also

- [file-structure-sprint](file-structure-sprint.md) - Enforces sprint-based organization
- [file-structure-dml](file-structure-dml.md) - Ensures DML changes in data directories
