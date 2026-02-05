# no-manual-transactions

**Severity**: Warning  
**Category**: Best Practices / Reliability  
**Status**: 📋 Planned

## Description

Detects manual transaction control statements (BEGIN TRANSACTION, START TRANSACTION, COMMIT, ROLLBACK) within changeset SQL. Liquibase manages transactions automatically based on changeset configuration, and manual transaction control can interfere with this mechanism, potentially causing issues with rollback functionality and change tracking.

## What it detects

- `BEGIN TRANSACTION`, `BEGIN TRAN`, `BEGIN`
- `START TRANSACTION`
- `COMMIT TRANSACTION`, `COMMIT TRAN`, `COMMIT`
- `ROLLBACK TRANSACTION`, `ROLLBACK TRAN`, `ROLLBACK`
- `SAVE TRANSACTION`, `SAVEPOINT`

## Example violation

**SQL format:**
```sql
--changeset john.doe:1
BEGIN TRANSACTION;
    INSERT INTO users (name, email) VALUES ('John', 'john@example.com');
    UPDATE settings SET value = 'new' WHERE key = 'config';
COMMIT TRANSACTION;
```

**XML format:**
```xml
<changeSet id="1" author="john.doe">
    <sql>
        BEGIN TRANSACTION;
        INSERT INTO users (name, email) VALUES ('John', 'john@example.com');
        UPDATE settings SET value = 'new' WHERE key = 'config';
        COMMIT TRANSACTION;
    </sql>
</changeSet>
```

## Correct usage

**Let Liquibase manage transactions (default):**
```sql
--changeset john.doe:1
INSERT INTO users (name, email) VALUES ('John', 'john@example.com');

--changeset john.doe:2
UPDATE settings SET value = 'new' WHERE key = 'config';
```

**Use runInTransaction attribute for control:**
```xml
<changeSet id="1" author="john.doe" runInTransaction="true">
    <sql>
        INSERT INTO users (name, email) VALUES ('John', 'john@example.com');
        UPDATE settings SET value = 'new' WHERE key = 'config';
    </sql>
</changeSet>
```

**For operations that must run outside transactions:**
```xml
<changeSet id="2" author="john.doe" runInTransaction="false">
    <sql>
        CREATE INDEX idx_users_email ON users(email);
    </sql>
</changeSet>
```

**SQL format with runInTransaction:**
```sql
--changeset john.doe:1 runInTransaction:false
CREATE INDEX idx_users_email ON users(email);
```

## Why this matters

### Transaction management conflicts
Liquibase automatically wraps changesets in transactions (by default). Manual transaction control can conflict with this behavior.

### Rollback functionality
Manual transaction control can interfere with Liquibase's automatic rollback tracking and execution.

### Database compatibility
Different databases have different transaction control syntax. Letting Liquibase manage transactions improves cross-database compatibility.

### Change tracking
If a manual transaction is rolled back within a changeset, Liquibase may still mark the changeset as executed, causing inconsistencies.

### Nested transaction issues
Manual transactions can create nested transaction scenarios that behave differently across database systems.

## Configuration

```yaml
rules:
  no-manual-transactions:
    enabled: true
    severity: warning
```

### Advanced configuration

```yaml
rules:
  no-manual-transactions:
    enabled: true
    severity: warning
    # Patterns to detect (default covers most SQL dialects)
    patterns:
      - '\bBEGIN\s+(TRANSACTION|TRAN|WORK)?\b'
      - '\bSTART\s+TRANSACTION\b'
      - '\bCOMMIT\s+(TRANSACTION|TRAN|WORK)?\b'
      - '\bROLLBACK\s+(TRANSACTION|TRAN|WORK)?\b'
      - '\bSAVE(POINT)?\s+TRANSACTION\b'
    # Case insensitive matching
    case_insensitive: true
    # Exclude stored procedures and functions
    exclude_change_types:
      - 'createProcedure'
      - 'createFunction'
```

## Exceptions

Some legitimate cases where transaction keywords may appear:

### 1. Stored procedures and functions
Stored procedures often manage their own transactions:

```sql
--changeset john.doe:1
CREATE PROCEDURE UpdateUserStatus
AS
BEGIN
    BEGIN TRANSACTION;
    UPDATE users SET status = 'active';
    COMMIT TRANSACTION;
END;
```

**Solution**: Exclude `createProcedure` change types in configuration.

### 2. Dynamic SQL in strings
Transaction keywords in string literals should not trigger violations:

```sql
--changeset john.doe:1
INSERT INTO audit_log (message) VALUES ('Transaction completed successfully');
```

**Solution**: Use smart pattern matching to ignore keywords in string literals.

### 3. Comments
Transaction keywords in comments:

```sql
--changeset john.doe:1
-- This change will run in a transaction
UPDATE users SET active = 1;
```

**Solution**: Strip SQL comments before pattern matching.

## Detection strategy

### Implementation approach

1. **Extract SQL content** from each `Change` object
2. **Normalize SQL**:
   - Remove SQL comments (`--` and `/* */`)
   - Remove string literals (single and double quotes)
   - Convert to uppercase for case-insensitive matching
3. **Apply regex patterns** to detect transaction keywords
4. **Filter by change type**:
   - Skip `createProcedure`, `createFunction`, `createTrigger` (configurable)
5. **Report violations** with line numbers when possible

### Regex patterns

```regex
(?i)\b(BEGIN)\s+(TRANSACTION|TRAN|WORK)?\b
(?i)\bSTART\s+TRANSACTION\b
(?i)\b(COMMIT)\s+(TRANSACTION|TRAN|WORK)?\b
(?i)\b(ROLLBACK)\s+(TRANSACTION|TRAN|WORK)?\b
(?i)\bSAVE(POINT)?\s+\w+\b
```

### SQL preprocessing

```go
// Pseudo-code for SQL preprocessing
func preprocessSQL(sql string) string {
    // Remove single-line comments
    sql = removeLineComments(sql)
    // Remove multi-line comments
    sql = removeBlockComments(sql)
    // Remove string literals
    sql = removeStringLiterals(sql)
    return sql
}
```

## Related rules

- [dangerous-operations](dangerous-operations.md): Detects risky database operations
- [non-idempotent](non-idempotent.md): Ensures changes can be safely re-run

## Implementation notes

### Parser support

All parsers support extracting SQL:
- **XML**: From `<sql>` elements and various change types
- **YAML**: From `sql` property and change definitions
- **SQL**: From raw SQL content in changesets

SQL is available in `Change.SQL` field.

### Configuration structure

```go
type NoManualTransactionsConfig struct {
    Enabled              bool     `yaml:"enabled"`
    Severity             string   `yaml:"severity"`
    Patterns             []string `yaml:"patterns"`
    CaseInsensitive      bool     `yaml:"case_insensitive"`
    ExcludeChangeTypes   []string `yaml:"exclude_change_types"`
    ExcludePatterns      []string `yaml:"exclude_patterns"` // File patterns
}
```

### Testing considerations

- Test with various transaction keywords (BEGIN, START, COMMIT, ROLLBACK)
- Test with different SQL dialects (SQL Server, PostgreSQL, MySQL)
- Test with stored procedures (should be excluded)
- Test with transaction keywords in comments (should be ignored)
- Test with transaction keywords in string literals (should be ignored)
- Test across all changelog formats (XML, YAML, SQL)
- Test with `runInTransaction` attribute set
- Test with nested BEGIN/COMMIT blocks

## Examples

### Bad: Manual transaction control

```sql
--changeset john.doe:update-multiple-tables
BEGIN TRANSACTION;

UPDATE table1 SET status = 'active';
UPDATE table2 SET updated_at = GETDATE();

IF @@ERROR != 0
    ROLLBACK TRANSACTION;
ELSE
    COMMIT TRANSACTION;
```

**Issue**: Manual transaction control interferes with Liquibase's transaction management.

### Good: Separate changesets

```sql
--changeset john.doe:update-table1
UPDATE table1 SET status = 'active';

--changeset john.doe:update-table2
UPDATE table2 SET updated_at = GETDATE();
```

**Why**: Each changeset is automatically wrapped in a transaction by Liquibase.

### Good: Use runInTransaction

```sql
--changeset john.doe:update-multiple-tables runInTransaction:true
UPDATE table1 SET status = 'active';
UPDATE table2 SET updated_at = GETDATE();
```

**Why**: Explicitly tells Liquibase to run both updates in a single transaction.

### Acceptable: Stored procedure with transactions

```sql
--changeset john.doe:create-update-proc
CREATE PROCEDURE UpdateWithTransaction
AS
BEGIN
    BEGIN TRANSACTION;
    -- Procedure logic with transaction control
    COMMIT TRANSACTION;
END;
```

**Why**: Stored procedures often need their own transaction management. This should be excluded via `exclude_change_types: ['createProcedure']`.

## Database-specific considerations

### SQL Server
- `BEGIN TRANSACTION` / `BEGIN TRAN`
- `COMMIT TRANSACTION` / `COMMIT TRAN`
- `ROLLBACK TRANSACTION` / `ROLLBACK TRAN`
- `SAVE TRANSACTION` for savepoints

### PostgreSQL
- `BEGIN` / `BEGIN WORK` / `BEGIN TRANSACTION`
- `COMMIT` / `COMMIT WORK` / `COMMIT TRANSACTION`
- `ROLLBACK` / `ROLLBACK WORK` / `ROLLBACK TRANSACTION`
- `SAVEPOINT <name>` for savepoints

### MySQL
- `START TRANSACTION`
- `COMMIT`
- `ROLLBACK`
- `SAVEPOINT <name>` for savepoints

The rule should detect all these variants using flexible regex patterns.
