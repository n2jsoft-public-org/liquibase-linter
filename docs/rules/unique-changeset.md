# unique-changeset

**Severity**: Critical  
**Category**: Best Practices  
**Status**: ✅ Implemented

## Description

Detects duplicate changesets within the same file or across included files in a changelog tree. A changeset is uniquely identified by the combination of ID + Author + FilePath. Duplicate changesets cause Liquibase execution failures and must be prevented.

## What it checks

- Each changeset must have a unique combination of ID, Author, and FilePath
- Checks all changesets across the entire changelog tree (including include/includeAll directives)
- Allows the same ID+Author combination in different files (per Liquibase specification)
- Reports all duplicate occurrences with location information

## Example violations

### Duplicate changeset in same file

```xml
<databaseChangeLog>
    <changeSet id="001" author="john.doe">
        <createTable tableName="users">
            <column name="id" type="INT"/>
        </createTable>
    </changeSet>
    
    <changeSet id="002" author="john.doe">
        <addColumn tableName="users">
            <column name="name" type="VARCHAR(100)"/>
        </addColumn>
    </changeSet>
    
    <!-- VIOLATION: Duplicate ID + Author in same file -->
    <changeSet id="001" author="john.doe">
        <createTable tableName="orders">
            <column name="id" type="INT"/>
        </createTable>
    </changeSet>
</databaseChangeLog>
```

**Error message**: `Duplicate changeset found: id='001' author='john.doe' in file 'changelog/v1.xml'`

## Correct usage

### Option 1: Use unique IDs

```xml
<databaseChangeLog>
    <changeSet id="001" author="john.doe">
        <createTable tableName="users">
            <column name="id" type="INT"/>
        </createTable>
    </changeSet>
    
    <changeSet id="002" author="john.doe">
        <addColumn tableName="users">
            <column name="name" type="VARCHAR(100)"/>
        </addColumn>
    </changeSet>
    
    <!-- Unique ID - no violation -->
    <changeSet id="003" author="john.doe">
        <createTable tableName="orders">
            <column name="id" type="INT"/>
        </createTable>
    </changeSet>
</databaseChangeLog>
```

### Option 2: Same ID in different files (allowed)

**File: changelog/v1/init.xml**
```xml
<changeSet id="001" author="john.doe">
    <createTable tableName="users">
        <column name="id" type="INT"/>
    </createTable>
</changeSet>
```

**File: changelog/v2/migration.xml**
```xml
<!-- No violation: Same ID+Author allowed in different files -->
<changeSet id="001" author="john.doe">
    <createTable tableName="products">
        <column name="id" type="INT"/>
    </createTable>
</changeSet>
```

### Option 3: Same ID with different authors (allowed)

```xml
<databaseChangeLog>
    <changeSet id="001" author="john.doe">
        <createTable tableName="users">
            <column name="id" type="INT"/>
        </createTable>
    </changeSet>
    
    <!-- No violation: Different author -->
    <changeSet id="001" author="jane.smith">
        <createTable tableName="orders">
            <column name="id" type="INT"/>
        </createTable>
    </changeSet>
</databaseChangeLog>
```

## Configuration

```yaml
rules:
  unique-changeset:
    enabled: true
    severity: critical
```

## Why this matters

Duplicate changesets cause critical issues:

1. **Liquibase Execution Failure**: Liquibase will fail to run with duplicate changeset errors
2. **Database State Conflicts**: Multiple definitions of the same changeset create ambiguity
3. **Rollback Problems**: Cannot determine which version of a changeset to rollback
4. **Build Pipeline Failures**: Prevents deployment and blocks CI/CD pipelines
5. **Team Confusion**: Makes it unclear which changeset definition is authoritative

This rule is marked **Critical** because duplicate changesets will cause immediate Liquibase failures in any environment.

## Uniqueness Rules

According to Liquibase specification, a changeset is unique when:
- `ID` + `Author` + `FilePath` are all identical

This means:
- ✅ Same ID with different author in same file: **Allowed**
- ✅ Same ID and author in different files: **Allowed**
- ❌ Same ID and author in same file: **Violation**

## References

- [Liquibase Changeset Documentation](https://docs.liquibase.com/concepts/changelogs/changeset.html)
- [Liquibase Best Practices](https://docs.liquibase.com/workflows/liquibase-community/liquibase-best-practices.html)
