# sql-injection

**Severity**: Critical  
**Category**: Security  
**Status**: ✅ Implemented

## Description

Detects potential SQL injection vulnerabilities in Liquibase changesets.

## What it detects

- String concatenation in SQL statements
- Unescaped variable substitution
- Dynamic SQL construction without parameterization

## Example violation

**XML format:**
```xml
<changeSet id="1" author="test">
    <sql>
        DELETE FROM users WHERE username = '${username}';
    </sql>
</changeSet>
```

**SQL format:**
```sql
--liquibase formatted sql
--changeset test:1
DELETE FROM users WHERE username = '${username}';
```

## Correct usage

**XML format:**
```xml
<changeSet id="1" author="test">
    <delete tableName="users">
        <where>username = :username</where>
    </delete>
</changeSet>
```

**SQL format (using Liquibase change types):**
```xml
<changeSet id="1" author="test">
    <delete tableName="users">
        <where>username = :username</where>
    </delete>
</changeSet>
```

Or use parameterized queries in your application layer instead of embedding dynamic values in migrations.

## Configuration

```yaml
rules:
  sql-injection:
    enabled: true
    severity: critical
```

## Why this matters

SQL injection vulnerabilities allow attackers to manipulate database queries, potentially leading to:
- Unauthorized data access
- Data modification or deletion
- Bypassing authentication
- Complete database compromise

Always use parameterized queries or Liquibase's built-in change types instead of raw SQL with string interpolation.
