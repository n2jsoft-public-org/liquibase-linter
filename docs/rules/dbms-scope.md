# DBMS Scope (dbms-scope)

**Category**: Reliability  
**Severity**: Critical (default)  
**Status**: 📋 Planned, see [roadmap/dbms-scope.md](../../roadmap/dbms-scope.md)

## What It Detects

For changelogs that run on several database engines, this rule checks the `dbms` attribute of
each changeset:

1. **Missing dbms**: the changeset has no `dbms` attribute (when `require_dbms` is on).
2. **Unknown engine**: a `dbms` value, with its `!` prefix removed, is not `all`, `none` or one of
   the allowed engine names.
3. **Engine-specific precondition SQL**: the changeset runs on an engine, and one of its `sqlCheck`
   preconditions matches a pattern configured as unsupported on that engine (for example `sys.*` or
   `OBJECT_ID()` on PostgreSQL).

The rule is disabled by default. Enable it for changelogs shared by more than one engine.

## Why It Matters

- **A changeset without `dbms` runs everywhere.** In a changelog shared by SQL Server and PostgreSQL,
  T-SQL written without `dbms:mssql` is applied to PostgreSQL too and fails there.
- **A misspelt engine name never fails.** `dbms:postgres` (instead of `postgresql`) matches no
  database, so Liquibase silently skips the changeset on every engine.
- **Preconditions run before the change.** A `sqlCheck` that queries `sys.tables` fails on
  PostgreSQL even when the change itself is portable.
- **Adding `dbms` is safe.** It does not change the changeset checksum, so it can be added to
  changesets that are already deployed.

A changeset runs on engine E when its `dbms` list allows E:

| `dbms` | Runs on PostgreSQL? |
|---|---|
| (absent) | yes |
| `mssql` | no |
| `postgresql`, `mssql,postgresql`, `all` | yes |
| `!mssql` | yes (only negations, and PostgreSQL is not excluded) |
| `!postgresql`, `!all`, `none` | no |

## Examples

### ❌ Incorrect

```sql
--liquibase formatted sql

-- No dbms: also runs on PostgreSQL, where NVARCHAR fails
--changeset alice:add-email
ALTER TABLE Users ADD Email NVARCHAR(255);

-- Typo: matches no database, the changeset never runs
--changeset alice:add-phone dbms:postgres
ALTER TABLE users ADD COLUMN phone VARCHAR(50);

-- Runs on PostgreSQL, but the precondition queries the SQL Server catalog
--changeset alice:add-index dbms:all
--preconditions onFail:MARK_RAN
--precondition-sql-check expectedResult:0 SELECT COUNT(*) FROM sys.indexes WHERE name = 'ix_users_email'
CREATE INDEX ix_users_email ON users (email);
```

### ✅ Correct

```sql
--liquibase formatted sql

--changeset alice:add-email dbms:mssql
ALTER TABLE Users ADD Email NVARCHAR(255);

--changeset alice:add-email-pg dbms:postgresql
ALTER TABLE users ADD COLUMN email VARCHAR(255);

--changeset alice:add-phone dbms:postgresql
ALTER TABLE users ADD COLUMN phone VARCHAR(50);

-- information_schema exists on both engines
--changeset alice:add-archive dbms:all
--preconditions onFail:MARK_RAN
--precondition-sql-check expectedResult:0 SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'users_archive'
CREATE TABLE users_archive (id INT PRIMARY KEY);
```

```xml
<changeSet id="add-email" author="alice" dbms="mssql">
    <sql>ALTER TABLE Users ADD Email NVARCHAR(255);</sql>
</changeSet>
```

## Configuration

```yaml
rules:
  dbms-scope:
    enabled: true
    severity: critical  # info, warning, critical

dbms_scope:
  enabled: true
  require_dbms: true
  allowed_values: [mssql, postgresql]
  exclude_patterns:
    - "changelog/generated/**"
  forbidden_precondition_patterns:
    postgresql:
      - '(?i)(^|[^a-z0-9_])sys\.'
      - '(?i)\bobject_id\s*\('
      - '(?i)\bcol_length\s*\('
```

### Configuration Options

- **enabled**: Enable the rule (default: `false`)
- **require_dbms**: Report changesets without a `dbms` attribute (default: `true`)
- **allowed_values**: Engine names accepted in `dbms`, in addition to `all` and `none`. Default:
  Liquibase's short names (`cockroachdb`, `db2`, `derby`, `h2`, `hsqldb`, `informix`, `mariadb`,
  `mssql`, `mysql`, `oracle`, `postgresql`, `snowflake`, `sqlite`, `sybase`). Restrict it to the
  engines you actually target to catch more typos.
- **forbidden_precondition_patterns**: Map of engine name to regular expressions (Go syntax). A
  changeset that runs on that engine must not have precondition SQL matching any of them.
  Default: none.
- **exclude_patterns**: Glob patterns for files to skip (optional)

## Limitations

- Precondition SQL is read from formatted SQL `--precondition-sql-check`, XML `<sqlCheck>` and
  YAML/JSON `sqlCheck`. An XML `<sqlCheck>` nested inside `<and>`, `<or>` or `<not>` is not checked.
- Only the changeset `dbms` attribute is read, not a `dbms` precondition.

## Suppression

```sql
--changeset alice:legacy
--comment: liquibase-linter:disable dbms-scope Applied before the PostgreSQL migration, kept as is
```

See [Rule Suppression](../suppression.md).

## See Also

- [DBMS Twin](dbms-twin.md) - Requires a counterpart changeset for a second engine
- [No IF EXISTS](no-if-exists.md) - Engine-specific existence checks in the SQL body
- [Liquibase dbms attribute](https://docs.liquibase.com/concepts/changelogs/attributes/dbms.html)
