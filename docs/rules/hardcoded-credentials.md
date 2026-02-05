# hardcoded-credentials

**Severity**: Critical  
**Category**: Security  
**Status**: ✅ Implemented

## Description

Finds hardcoded passwords, API keys, and other sensitive credentials in changelog files.

## What it detects

- Hardcoded passwords in CREATE USER statements
- API keys and tokens in plain text
- Connection strings with embedded credentials

## Example violation

**XML format:**
```xml
<changeSet id="1" author="test">
    <sql>
        CREATE USER 'app_user'@'localhost' IDENTIFIED BY 'password123';
    </sql>
</changeSet>
```

**SQL format:**
```sql
--liquibase formatted sql
--changeset test:1
CREATE USER 'app_user'@'localhost' IDENTIFIED BY 'password123';
```

## Correct usage

Use environment variables or secret management systems:

**XML format:**
```xml
<changeSet id="1" author="test">
    <sql>
        CREATE USER 'app_user'@'localhost' IDENTIFIED BY '${env.DB_PASSWORD}';
    </sql>
</changeSet>
```

**SQL format:**
```sql
--liquibase formatted sql
--changeset test:1
CREATE USER 'app_user'@'localhost' IDENTIFIED BY '${env.DB_PASSWORD}';
```

Or use Liquibase properties loaded from secure sources.

## Configuration

```yaml
rules:
  hardcoded-credentials:
    enabled: true
    severity: critical
```

## Why this matters

Hardcoded credentials in source code:
- Can be exposed through version control systems
- Make credential rotation difficult
- Violate security best practices
- May lead to unauthorized access if the repository is compromised
