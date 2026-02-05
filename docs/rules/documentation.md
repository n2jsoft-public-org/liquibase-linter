# documentation

**Severity**: Info  
**Category**: Best Practices  
**Status**: ✅ Implemented

## Description

Ensures changesets are properly documented with comments and context.

## What it checks

- Comments describing the changeset purpose
- JIRA/ticket references
- Complex changes have explanations
- Business context is provided

## Example violation

```xml
<changeSet id="1" author="test">
    <sql>
        ALTER TABLE users ADD COLUMN temp_flag BOOLEAN DEFAULT FALSE;
        UPDATE users SET temp_flag = TRUE WHERE status = 'pending';
        DELETE FROM users WHERE temp_flag = FALSE;
        ALTER TABLE users DROP COLUMN temp_flag;
    </sql>
</changeSet>
```

## Correct usage

```xml
<changeSet id="1" author="test">
    <comment>
        TICKET-123: Clean up pending user accounts.
        Removes users with 'pending' status as part of the account verification redesign.
        This is safe to run as verified users will not be affected.
    </comment>
    <sql>
        ALTER TABLE users ADD COLUMN temp_flag BOOLEAN DEFAULT FALSE;
        UPDATE users SET temp_flag = TRUE WHERE status = 'pending';
        DELETE FROM users WHERE temp_flag = FALSE;
        ALTER TABLE users DROP COLUMN temp_flag;
    </sql>
</changeSet>
```

## Configuration

```yaml
rules:
  documentation:
    enabled: true
    severity: info
```

## Why this matters

Good documentation:
- Helps team members understand the purpose of changes
- Makes code reviews more effective
- Aids in troubleshooting issues
- Provides context for future maintenance
- Links database changes to business requirements

Always document complex changes, data migrations, and anything that might not be immediately obvious to another developer.
