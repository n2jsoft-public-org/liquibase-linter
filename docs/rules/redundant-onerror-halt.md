# redundant-onerror-halt

**Severity**: Info  
**Category**: Best Practices  
**Status**: 📋 Planned

## Description

Detects redundant `onError:HALT` configuration in preconditions. Since `HALT` is the default value for `onError`, explicitly specifying it is unnecessary and adds clutter to changelog files.

## What it detects

- Preconditions with explicit `onError="HALT"` attribute
- Preconditions with explicit `onError:HALT` in SQL format

## Example violation

**XML format:**
```xml
<changeSet id="1" author="john.doe">
    <preConditions onFail="MARK_RAN" onError="HALT">
        <tableExists tableName="users"/>
    </preConditions>
    <addColumn tableName="users">
        <column name="email" type="VARCHAR(255)"/>
    </addColumn>
</changeSet>
```

**SQL format:**
```sql
--changeset john.doe:1
--preconditions onFail:MARK_RAN onError:HALT
--precondition-sql-check expectedResult:0 SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'users'
ALTER TABLE users ADD COLUMN email VARCHAR(255);
```

**YAML format:**
```yaml
databaseChangeLog:
  - changeSet:
      id: 1
      author: john.doe
      preConditions:
        onFail: MARK_RAN
        onError: HALT  # Redundant - this is the default
        tableExists:
          tableName: users
```

## Correct usage

**XML format (without explicit onError):**
```xml
<changeSet id="1" author="john.doe">
    <preConditions onFail="MARK_RAN">
        <tableExists tableName="users"/>
    </preConditions>
    <addColumn tableName="users">
        <column name="email" type="VARCHAR(255)"/>
    </addColumn>
</changeSet>
```

**SQL format:**
```sql
--changeset john.doe:1
--preconditions onFail:MARK_RAN
--precondition-sql-check expectedResult:0 SELECT COUNT(*) FROM INFORMATION_SCHEMA.TABLES WHERE TABLE_NAME = 'users'
ALTER TABLE users ADD COLUMN email VARCHAR(255);
```

**YAML format:**
```yaml
databaseChangeLog:
  - changeSet:
      id: 1
      author: john.doe
      preConditions:
        onFail: MARK_RAN
        tableExists:
          tableName: users
```

**When onError should be specified (non-default values):**
```xml
<changeSet id="2" author="john.doe">
    <preConditions onFail="MARK_RAN" onError="WARN">
        <tableExists tableName="users"/>
    </preConditions>
    <addColumn tableName="users">
        <column name="status" type="VARCHAR(50)"/>
    </addColumn>
</changeSet>
```

## Why this matters

### Code clarity
Removing redundant default values makes changelogs cleaner and easier to read.

### Maintenance
Explicit default values can be confusing - developers may wonder if there's a specific reason for the explicit declaration.

### Best practices
Following the principle of "only specify what differs from defaults" improves code consistency across the project.

### Documentation value
When a default value is specified, it adds no informational value but increases visual noise.

## Liquibase onError values

The `onError` attribute controls what happens when a precondition check encounters an error (not a failure):

- **HALT** (default): Stop execution and throw an error
- **CONTINUE**: Continue executing the changeset
- **MARK_RAN**: Skip the changeset and mark it as executed
- **WARN**: Output a warning but continue execution

Since `HALT` is the default behavior, it never needs to be explicitly specified.

## Configuration

```yaml
rules:
  redundant-onerror-halt:
    enabled: true
    severity: info
```

## Related rules

- [mandatory-preconditions](mandatory-preconditions.md): Requires preconditions on all changesets
- [missing-preconditions](missing-preconditions.md): Detects risky operations without preconditions

## Implementation notes

### Detection strategy

1. Check if changeset has preconditions (`ChangeSet.Preconditions != nil`)
2. Check if `Precondition.OnError` is set to `"HALT"`
3. Report violation if found

### Parser support

All parsers already extract the `onError` attribute:
- **XML**: Parse `onError` attribute from `<preConditions>` element
- **YAML**: Parse `onError` key from `preConditions` object
- **SQL**: Parse `onError:HALT` from `--preconditions` comment

The value is stored in `Precondition.OnError` field.

### Testing considerations

- Test with `onError="HALT"` (violation)
- Test with `onError="WARN"` (no violation - non-default)
- Test with `onError="CONTINUE"` (no violation - non-default)
- Test with `onError="MARK_RAN"` (no violation - non-default)
- Test with no `onError` attribute (no violation)
- Test across all changelog formats (XML, YAML, SQL)

## Examples from Liquibase documentation

According to Liquibase documentation, the default `onError` is `HALT`, so these are equivalent:

```xml
<!-- Explicit (redundant) -->
<preConditions onError="HALT">
    <tableExists tableName="users"/>
</preConditions>

<!-- Implicit (preferred) -->
<preConditions>
    <tableExists tableName="users"/>
</preConditions>
```

Both will halt execution if the precondition check encounters an error.

## Auto-fix capability

This rule is a good candidate for auto-fix functionality:
- Simply remove the `onError="HALT"` attribute
- No logic changes, purely cosmetic
- Safe transformation with no behavioral impact

## Expected output

```
INFO: redundant-onerror-halt
File: changelog/sprints/v123/0 - structure/add-column.xml
Changeset: john.doe:456
Message: Redundant 'onError="HALT"' - this is the default behavior and can be omitted
```

---

**References**:
- [Liquibase Preconditions Documentation](https://docs.liquibase.com/concepts/changelogs/preconditions.html)
