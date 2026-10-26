# DBMS Twin (dbms-twin)

**Category**: Reliability  
**Severity**: Critical (default)  
**Status**: 📋 Planned, see [roadmap/dbms-scope.md](../../roadmap/dbms-scope.md)

## What It Detects

When a changelog is migrated from one engine (the *source*) to another (the *target*), every change
written for the source must also exist for the target. This rule checks that:

1. A changeset that runs on the source engine only has a **twin** in the same file: same author, ID
   followed by the configured suffix (e.g. `add-email` → `add-email-pg`), and a `dbms` that reaches
   the target engine.
2. A changeset whose ID ends with the suffix declares `dbms` equal to the target engine alone.

A changeset that already reaches the target (`dbms:mssql,postgresql`, `dbms:all`, no `dbms`) needs
no twin. The rule can be limited to sprint folders from a given number on.

The rule is disabled by default.

## Why It Matters

- **Schema drift.** A SQL Server change with no PostgreSQL counterpart leaves the PostgreSQL schema
  behind, and the gap only shows when an application query fails on PostgreSQL.
- **Twin on the wrong engine.** A `-pg` changeset copied from its source and left with `dbms:mssql`
  runs twice on SQL Server and never on PostgreSQL.
- **Explicit exemptions.** Some changes have no counterpart on purpose (a data fix for rows that only
  exist in production, a SQL Server-only feature). Suppressing the rule on that changeset documents
  why.

## Examples

### ❌ Incorrect

```sql
--liquibase formatted sql

-- No PostgreSQL twin
--changeset alice:add-phone dbms:mssql
ALTER TABLE Users ADD Phone NVARCHAR(50);

-- Twin declared for the wrong engine
--changeset alice:add-email dbms:mssql
ALTER TABLE Users ADD Email NVARCHAR(255);

--changeset alice:add-email-pg dbms:mssql
ALTER TABLE users ADD COLUMN email VARCHAR(255);
```

### ✅ Correct

```sql
--liquibase formatted sql

--changeset alice:add-email dbms:mssql
ALTER TABLE Users ADD Email NVARCHAR(255);

--changeset alice:add-email-pg dbms:postgresql
ALTER TABLE users ADD COLUMN email VARCHAR(255);

--changeset alice:fulltext dbms:mssql
--comment: liquibase-linter:disable dbms-twin SQL Server full-text index, no PostgreSQL equivalent
CREATE FULLTEXT INDEX ON Users(Name) KEY INDEX PK_Users;
```

## Configuration

```yaml
rules:
  dbms-twin:
    enabled: true
    severity: critical  # info, warning, critical

dbms_twin:
  enabled: true
  source_dbms: mssql
  target_dbms: postgresql
  id_suffix: "-pg"
  # Only check sprint folders v10 and later
  since_path_pattern: 'changelog/sprints/v(\d+)/'
  since: 10
  exclude_patterns:
    - "changelog/generated/**"
```

### Configuration Options

- **enabled**: Enable the rule (default: `false`)
- **source_dbms**: Engine the changelog is migrated from (required)
- **target_dbms**: Engine the changelog is migrated to (required)
- **id_suffix**: Suffix added to the source changeset ID to name its twin (default: `-pg`)
- **since_path_pattern**: Regular expression whose first capture group is a number; files that do
  not match it are not checked (optional)
- **since**: Lowest number accepted from `since_path_pattern` (optional, requires `since_path_pattern`)
- **exclude_patterns**: Glob patterns for files to skip (optional)

## Suppression

To state that a source changeset needs no twin, suppress the rule on it and give the reason in the
same comment:

```sql
--changeset alice:fix-prod-rows dbms:mssql
--comment: liquibase-linter:disable dbms-twin Data fix for rows that only exist in production
UPDATE Orders SET Status = 3 WHERE Id IN (1, 2);
```

See [Rule Suppression](../suppression.md).

## See Also

- [DBMS Scope](dbms-scope.md) - Requires a valid `dbms` on every changeset
- [Liquibase dbms attribute](https://docs.liquibase.com/concepts/changelogs/attributes/dbms.html)
