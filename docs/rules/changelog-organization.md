# changelog-organization

**Severity**: Info  
**Category**: Best Practices  
**Status**: ✅ Implemented

## Description

Ensures proper organization of changelog files and changesets.

## What it checks

- ChangeSet IDs are sequential and meaningful
- Author information is present
- Contexts are used appropriately
- Labels are meaningful
- File organization follows a logical structure

## Example violation

```xml
<changeSet id="random123" author="">
    <createTable tableName="users">
        <column name="id" type="INT"/>
    </createTable>
</changeSet>
```

## Correct usage

```xml
<changeSet id="001-create-users-table" author="john.doe">
    <createTable tableName="users">
        <column name="id" type="INT"/>
    </createTable>
</changeSet>
```

## Configuration

```yaml
rules:
  changelog-organization:
    enabled: true
    severity: info
```

## Why this matters

Well-organized changelogs:
- Make it easier to find and review specific changes
- Improve team collaboration
- Enable better auditing and compliance
- Simplify troubleshooting
- Provide a clear history of database evolution

Good organization also makes it easier to roll back changes or understand the database schema evolution over time.
