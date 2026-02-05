# missing-preconditions

**Severity**: Warning  
**Category**: Reliability  
**Status**: ✅ Implemented

## Description

Ensures risky operations have appropriate preconditions.

## What it detects

- ALTER TABLE without column existence checks
- DROP operations without existence checks
- Data manipulation without validation

## Example violation

```xml
<changeSet id="1" author="test">
    <addColumn tableName="users">
        <column name="email" type="VARCHAR(255)"/>
    </addColumn>
</changeSet>
```

## Correct usage

```xml
<changeSet id="1" author="test">
    <preConditions onFail="MARK_RAN">
        <tableExists tableName="users"/>
        <not>
            <columnExists tableName="users" columnName="email"/>
        </not>
    </preConditions>
    <addColumn tableName="users">
        <column name="email" type="VARCHAR(255)"/>
    </addColumn>
</changeSet>
```

## Configuration

```yaml
rules:
  missing-preconditions:
    enabled: true
    severity: warning
```

## Why this matters

Preconditions help:
- Prevent errors from changes applied to unexpected database states
- Make changesets safer to run in different environments
- Document assumptions about database state
- Enable better error handling with onFail attributes

Preconditions act as a safety net, ensuring changes only execute when the database is in the expected state.
