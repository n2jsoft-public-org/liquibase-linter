# Examples

This guide provides practical examples of common Liquibase patterns and how the linter helps improve them.

## Table of Contents

- [Security Examples](#security-examples)
- [Reliability Examples](#reliability-examples)
- [Performance Examples](#performance-examples)
- [Best Practice Examples](#best-practice-examples)
- [Common Scenarios](#common-scenarios)

## Security Examples

### SQL Injection Prevention

❌ **Bad**: String concatenation with variables

```xml
<changeSet id="1" author="dev">
    <sql>
        DELETE FROM users WHERE username = '${username}';
    </sql>
</changeSet>
```

**Linter output**: `[CRITICAL] sql-injection: String concatenation detected in SQL statement`

✅ **Good**: Use Liquibase's built-in change types

```xml
<changeSet id="1" author="dev">
    <delete tableName="users">
        <where>username = :username</where>
    </delete>
</changeSet>
```

### Hardcoded Credentials

❌ **Bad**: Passwords in plain text

```xml
<changeSet id="1" author="dev">
    <sql>
        CREATE USER 'app_user'@'localhost' IDENTIFIED BY 'SecretPassword123!';
    </sql>
</changeSet>
```

**Linter output**: `[CRITICAL] hardcoded-credentials: Hardcoded password detected`

✅ **Good**: Use environment variables or secret management

```xml
<changeSet id="1" author="dev">
    <sql>
        CREATE USER 'app_user'@'localhost' IDENTIFIED BY '${db.app_password}';
    </sql>
</changeSet>
```

Better yet, manage users outside of Liquibase using proper secret management tools.

### Dangerous Operations

❌ **Bad**: Drop table without preconditions

```xml
<changeSet id="1" author="dev">
    <dropTable tableName="customer_data"/>
</changeSet>
```

**Linter output**: `[CRITICAL] dangerous-operations: DROP TABLE without preconditions`

✅ **Good**: Add preconditions and context

```xml
<changeSet id="1" author="dev" context="test">
    <preConditions onFail="MARK_RAN">
        <tableExists tableName="customer_data"/>
        <sqlCheck expectedResult="0">
            SELECT COUNT(*) FROM customer_data
        </sqlCheck>
    </preConditions>
    <dropTable tableName="customer_data"/>
    <rollback>
        <!-- Recreation script would go here -->
    </rollback>
</changeSet>
```

### Privilege Escalation

❌ **Bad**: Granting excessive privileges

```xml
<changeSet id="1" author="dev">
    <sql>
        GRANT ALL PRIVILEGES ON *.* TO 'app_user'@'%';
    </sql>
</changeSet>
```

**Linter output**: `[CRITICAL] privilege-escalation: Excessive privileges granted`

✅ **Good**: Grant minimal necessary privileges

```xml
<changeSet id="1" author="dev">
    <sql>
        GRANT SELECT, INSERT, UPDATE, DELETE ON mydb.* TO 'app_user'@'localhost';
    </sql>
</changeSet>
```

## Reliability Examples

### Missing Rollback

❌ **Bad**: No rollback script

```xml
<changeSet id="1" author="dev">
    <addColumn tableName="users">
        <column name="phone" type="VARCHAR(20)"/>
    </addColumn>
</changeSet>
```

**Linter output**: `[WARNING] missing-rollback: Changeset does not include rollback script`

✅ **Good**: Include explicit rollback

```xml
<changeSet id="1" author="dev">
    <addColumn tableName="users">
        <column name="phone" type="VARCHAR(20)"/>
    </addColumn>
    <rollback>
        <dropColumn tableName="users" columnName="phone"/>
    </rollback>
</changeSet>
```

### Non-Idempotent Changes

❌ **Bad**: Insert without precondition

```xml
<changeSet id="1" author="dev">
    <insert tableName="config">
        <column name="key" value="app_version"/>
        <column name="value" value="1.0.0"/>
    </insert>
</changeSet>
```

**Linter output**: `[WARNING] non-idempotent: Insert without precondition may fail on re-run`

✅ **Good**: Add precondition to make it idempotent

```xml
<changeSet id="1" author="dev">
    <preConditions onFail="MARK_RAN">
        <sqlCheck expectedResult="0">
            SELECT COUNT(*) FROM config WHERE key = 'app_version'
        </sqlCheck>
    </preConditions>
    <insert tableName="config">
        <column name="key" value="app_version"/>
        <column name="value" value="1.0.0"/>
    </insert>
</changeSet>
```

### Missing Preconditions

❌ **Bad**: Alter without checking existence

```xml
<changeSet id="1" author="dev">
    <addNotNullConstraint tableName="users" columnName="email"/>
</changeSet>
```

**Linter output**: `[WARNING] missing-preconditions: Adding NOT NULL without existence check`

✅ **Good**: Check before altering

```xml
<changeSet id="1" author="dev">
    <preConditions onFail="MARK_RAN">
        <columnExists tableName="users" columnName="email"/>
        <sqlCheck expectedResult="0">
            SELECT COUNT(*) FROM users WHERE email IS NULL
        </sqlCheck>
    </preConditions>
    <addNotNullConstraint tableName="users" columnName="email"/>
    <rollback>
        <dropNotNullConstraint tableName="users" columnName="email"/>
    </rollback>
</changeSet>
```

## Performance Examples

### Missing Indexes

❌ **Bad**: Foreign key without index

```xml
<changeSet id="1" author="dev">
    <addForeignKeyConstraint
        constraintName="fk_orders_user"
        baseTableName="orders"
        baseColumnNames="user_id"
        referencedTableName="users"
        referencedColumnNames="id"/>
</changeSet>
```

**Linter output**: `[INFO] missing-indexes: Foreign key without index on user_id`

✅ **Good**: Add index for foreign key

```xml
<changeSet id="1" author="dev">
    <createIndex indexName="idx_orders_user_id" tableName="orders">
        <column name="user_id"/>
    </createIndex>
    <addForeignKeyConstraint
        constraintName="fk_orders_user"
        baseTableName="orders"
        baseColumnNames="user_id"
        referencedTableName="users"
        referencedColumnNames="id"/>
</changeSet>
```

### Table Locks

❌ **Bad**: Adding column with NOT NULL to large table

```xml
<changeSet id="1" author="dev">
    <addColumn tableName="large_events_table">
        <column name="processed" type="BOOLEAN" defaultValueBoolean="false">
            <constraints nullable="false"/>
        </column>
    </addColumn>
</changeSet>
```

**Linter output**: `[WARNING] table-locks: Adding NOT NULL column may lock table`

✅ **Good**: Add nullable first, backfill, then add constraint

```xml
<changeSet id="1" author="dev">
    <addColumn tableName="large_events_table">
        <column name="processed" type="BOOLEAN"/>
    </addColumn>
</changeSet>

<changeSet id="2" author="dev">
    <update tableName="large_events_table">
        <column name="processed" valueBoolean="false"/>
        <where>processed IS NULL</where>
    </update>
</changeSet>

<changeSet id="3" author="dev">
    <addNotNullConstraint tableName="large_events_table" 
                          columnName="processed" 
                          defaultNullValue="false"/>
</changeSet>
```

### Large Data Operations

❌ **Bad**: Unbounded UPDATE

```xml
<changeSet id="1" author="dev">
    <sql>
        UPDATE users SET status = 'inactive';
    </sql>
</changeSet>
```

**Linter output**: `[WARNING] large-data-operations: Unbounded UPDATE detected`

✅ **Good**: Add WHERE clause or batch the operation

```xml
<changeSet id="1" author="dev">
    <sql>
        UPDATE users 
        SET status = 'inactive' 
        WHERE last_login < DATE_SUB(NOW(), INTERVAL 1 YEAR);
    </sql>
</changeSet>
```

## Best Practice Examples

### Naming Conventions

❌ **Bad**: Inconsistent naming

```xml
<changeSet id="1" author="dev">
    <createTable tableName="UserData">
        <column name="ID" type="BIGINT"/>
        <column name="user-name" type="VARCHAR(50)"/>
        <column name="Email_Address" type="VARCHAR(100)"/>
    </createTable>
</changeSet>
```

**Linter output**: Multiple naming convention violations

✅ **Good**: Consistent snake_case naming

```xml
<changeSet id="1" author="dev">
    <createTable tableName="user_data">
        <column name="id" type="BIGINT"/>
        <column name="user_name" type="VARCHAR(50)"/>
        <column name="email_address" type="VARCHAR(100)"/>
    </createTable>
</changeSet>
```

### Changelog Organization

❌ **Bad**: Poor organization

```xml
<changeSet id="1" author="dev">
    <createTable tableName="users">
        <!-- ... -->
    </createTable>
</changeSet>

<changeSet id="1" author="dev">  <!-- Duplicate ID! -->
    <createTable tableName="orders">
        <!-- ... -->
    </createTable>
</changeSet>
```

**Linter output**: `[INFO] changelog-organization: Duplicate changeset ID`

✅ **Good**: Unique, sequential IDs with meaningful author

```xml
<changeSet id="2024-001-create-users-table" author="john.doe">
    <comment>JIRA-123: Create users table for authentication</comment>
    <createTable tableName="users">
        <!-- ... -->
    </createTable>
</changeSet>

<changeSet id="2024-002-create-orders-table" author="john.doe">
    <comment>JIRA-124: Create orders table for e-commerce</comment>
    <createTable tableName="orders">
        <!-- ... -->
    </createTable>
</changeSet>
```

### Documentation

❌ **Bad**: No documentation

```xml
<changeSet id="1" author="dev">
    <sql>
        ALTER TABLE users ADD COLUMN flags BIGINT DEFAULT 0;
    </sql>
</changeSet>
```

**Linter output**: `[INFO] documentation: Changeset lacks descriptive comment`

✅ **Good**: Clear documentation

```xml
<changeSet id="1" author="dev">
    <comment>
        JIRA-567: Add bitwise flags column for user preferences
        
        This column stores user preferences as bit flags:
        - Bit 0: Email notifications enabled
        - Bit 1: SMS notifications enabled
        - Bit 2: Push notifications enabled
        - Bit 3: Marketing emails enabled
    </comment>
    <addColumn tableName="users">
        <column name="notification_flags" type="BIGINT" defaultValueNumeric="0">
            <constraints nullable="false"/>
        </column>
    </addColumn>
    <rollback>
        <dropColumn tableName="users" columnName="notification_flags"/>
    </rollback>
</changeSet>
```

## Common Scenarios

### Scenario 1: Adding a New Table

Complete example with all best practices:

```xml
<changeSet id="2024-010-create-products-table" author="alice.smith">
    <comment>JIRA-789: Create products table for catalog management</comment>
    
    <preConditions onFail="MARK_RAN">
        <not>
            <tableExists tableName="products"/>
        </not>
    </preConditions>
    
    <createTable tableName="products">
        <column name="id" type="BIGINT" autoIncrement="true">
            <constraints primaryKey="true" nullable="false"/>
        </column>
        <column name="sku" type="VARCHAR(50)">
            <constraints nullable="false" unique="true"/>
        </column>
        <column name="name" type="VARCHAR(200)">
            <constraints nullable="false"/>
        </column>
        <column name="description" type="TEXT"/>
        <column name="price" type="DECIMAL(10,2)">
            <constraints nullable="false"/>
        </column>
        <column name="category_id" type="BIGINT"/>
        <column name="created_at" type="TIMESTAMP" defaultValueComputed="CURRENT_TIMESTAMP">
            <constraints nullable="false"/>
        </column>
        <column name="updated_at" type="TIMESTAMP" defaultValueComputed="CURRENT_TIMESTAMP"/>
    </createTable>
    
    <createIndex indexName="idx_products_sku" tableName="products">
        <column name="sku"/>
    </createIndex>
    
    <createIndex indexName="idx_products_category_id" tableName="products">
        <column name="category_id"/>
    </createIndex>
    
    <rollback>
        <dropTable tableName="products"/>
    </rollback>
</changeSet>
```

### Scenario 2: Modifying Existing Data

Safe data migration example:

```xml
<changeSet id="2024-011-normalize-email-addresses" author="bob.jones">
    <comment>JIRA-790: Normalize all email addresses to lowercase</comment>
    
    <preConditions onFail="WARN">
        <tableExists tableName="users"/>
        <columnExists tableName="users" columnName="email"/>
    </preConditions>
    
    <sql>
        UPDATE users 
        SET email = LOWER(email)
        WHERE email != LOWER(email);
    </sql>
    
    <rollback>
        <comment>Cannot rollback email normalization</comment>
    </rollback>
</changeSet>
```

### Scenario 3: Refactoring Schema

Multi-step refactoring with backward compatibility:

```xml
<!-- Step 1: Add new column -->
<changeSet id="2024-012-add-full-name-column" author="charlie.brown">
    <comment>JIRA-791: Add full_name column (step 1 of 3)</comment>
    
    <addColumn tableName="users">
        <column name="full_name" type="VARCHAR(200)"/>
    </addColumn>
    
    <rollback>
        <dropColumn tableName="users" columnName="full_name"/>
    </rollback>
</changeSet>

<!-- Step 2: Populate new column -->
<changeSet id="2024-013-populate-full-name" author="charlie.brown">
    <comment>JIRA-791: Populate full_name from first_name and last_name (step 2 of 3)</comment>
    
    <sql>
        UPDATE users 
        SET full_name = CONCAT(first_name, ' ', last_name)
        WHERE full_name IS NULL;
    </sql>
    
    <rollback>
        <sql>UPDATE users SET full_name = NULL;</sql>
    </rollback>
</changeSet>

<!-- Step 3: Add NOT NULL constraint -->
<changeSet id="2024-014-add-full-name-constraint" author="charlie.brown">
    <comment>JIRA-791: Make full_name NOT NULL (step 3 of 3)</comment>
    
    <addNotNullConstraint tableName="users" 
                          columnName="full_name"
                          defaultNullValue="Unknown"/>
    
    <rollback>
        <dropNotNullConstraint tableName="users" columnName="full_name"/>
    </rollback>
</changeSet>

<!-- Optional Step 4: Remove old columns (after app update) -->
<changeSet id="2024-015-drop-old-name-columns" author="charlie.brown" context="cleanup">
    <comment>JIRA-791: Drop first_name and last_name columns after app migration</comment>
    
    <preConditions onFail="WARN">
        <columnExists tableName="users" columnName="first_name"/>
        <columnExists tableName="users" columnName="last_name"/>
    </preConditions>
    
    <dropColumn tableName="users" columnName="first_name"/>
    <dropColumn tableName="users" columnName="last_name"/>
    
    <rollback>
        <addColumn tableName="users">
            <column name="first_name" type="VARCHAR(100)"/>
            <column name="last_name" type="VARCHAR(100)"/>
        </addColumn>
    </rollback>
</changeSet>
```

## Configuration for Progressive Adoption

If you have an existing project with violations, adopt the linter progressively:

```yaml
# .liquibase-linter.yaml

# Start with critical issues only
severity_threshold: critical
fail_on: critical

rules:
  # Enable critical security rules first
  sql-injection:
    enabled: true
    severity: critical
  hardcoded-credentials:
    enabled: true
    severity: critical
  dangerous-operations:
    enabled: true
    severity: critical
  
  # Disable other rules initially
  missing-rollback:
    enabled: false
  naming-conventions:
    enabled: false

# Ignore legacy files you'll refactor later
ignore:
  - "db/changelog/legacy/**"
  - "db/changelog/2020-*.xml"
  - "db/changelog/2021-*.xml"
```

Then gradually:
1. Fix critical issues
2. Enable warning-level rules
3. Reduce ignore patterns
4. Enable info-level rules

## Resources

- [Usage Guide](usage.md) - Comprehensive usage documentation
- [Rules Reference](rules.md) - Complete rule documentation
- [Configuration Guide](configuration.md) - Configuration options
- [CI/CD Integration](cicd.md) - CI/CD setup examples
