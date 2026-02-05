# missing-indexes

**Severity**: Info  
**Category**: Performance  
**Status**: ✅ Implemented

## Description

Detects tables without proper indexes on foreign keys and other frequently queried columns.

## What it detects

- Foreign key columns without indexes
- Large tables without any indexes
- Missing indexes on frequently queried columns

## Example violation

```xml
<changeSet id="1" author="test">
    <createTable tableName="orders">
        <column name="id" type="INT">
            <constraints primaryKey="true"/>
        </column>
        <column name="user_id" type="INT"/>
    </createTable>
    <addForeignKeyConstraint
        baseTableName="orders"
        baseColumnNames="user_id"
        referencedTableName="users"
        referencedColumnNames="id"
        constraintName="fk_orders_user"/>
</changeSet>
```

## Correct usage

**XML format:**
```xml
<changeSet id="1" author="test">
    <createTable tableName="orders">
        <column name="id" type="INT">
            <constraints primaryKey="true"/>
        </column>
        <column name="user_id" type="INT"/>
    </createTable>
    <createIndex tableName="orders" indexName="idx_orders_user_id">
        <column name="user_id"/>
    </createIndex>
    <addForeignKeyConstraint
        baseTableName="orders"
        baseColumnNames="user_id"
        referencedTableName="users"
        referencedColumnNames="id"
        constraintName="fk_orders_user"/>
</changeSet>
```

**SQL format:**
```sql
--liquibase formatted sql
--changeset test:1
CREATE TABLE orders (
    id INT PRIMARY KEY,
    user_id INT
);

--changeset test:2
CREATE INDEX idx_orders_user_id ON orders(user_id);

--changeset test:3
ALTER TABLE orders 
    ADD CONSTRAINT fk_orders_user 
    FOREIGN KEY (user_id) REFERENCES users(id);
```

## Configuration

```yaml
rules:
  missing-indexes:
    enabled: true
    severity: info
```

## Why this matters

Missing indexes can cause:
- Slow query performance
- Full table scans for JOIN operations
- Degraded application responsiveness
- Increased database load

Foreign key columns are especially important to index as they're frequently used in JOIN operations.
