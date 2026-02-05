# privilege-escalation

**Severity**: Critical  
**Category**: Security  
**Status**: ✅ Implemented

## Description

Detects excessive privilege grants that could lead to security issues.

## What it detects

- GRANT ALL PRIVILEGES
- Creating superusers
- Granting permissions to wildcard users ('%')

## Example violation

**XML format:**
```xml
<changeSet id="1" author="test">
    <sql>
        GRANT ALL PRIVILEGES ON *.* TO 'app_user'@'%';
    </sql>
</changeSet>
```

**SQL format:**
```sql
--liquibase formatted sql
--changeset test:1
GRANT ALL PRIVILEGES ON *.* TO 'app_user'@'%';

--changeset test:2
CREATE USER 'admin'@'%' IDENTIFIED BY 'pass' WITH GRANT OPTION;
```

## Correct usage

Grant only the minimum necessary privileges:

**XML format:**
```xml
<changeSet id="1" author="test">
    <sql>
        GRANT SELECT, INSERT, UPDATE ON mydb.* TO 'app_user'@'localhost';
    </sql>
</changeSet>
```

**SQL format:**
```sql
--liquibase formatted sql
--changeset test:1
GRANT SELECT, INSERT, UPDATE ON mydb.* TO 'app_user'@'localhost';
```

## Configuration

```yaml
rules:
  privilege-escalation:
    enabled: true
    severity: critical
```

## Why this matters

Excessive privileges violate the principle of least privilege and can:
- Allow unauthorized access to sensitive data
- Enable malicious or accidental damage
- Make auditing and compliance more difficult
- Increase the impact of compromised accounts

Always grant the minimum privileges necessary for the application to function.
