--liquibase formatted sql

--changeset john:1
--comment: liquibase-linter:disable sql-injection
--comment: This creates the users table with basic fields
CREATE TABLE users (
    id INT PRIMARY KEY,
    username VARCHAR(100)
);
INSERT INTO users VALUES (1, ${username});

--changeset jane:2
--comment: Update email field to support longer addresses
--comment: liquibase-linter:disable missing-rollback
ALTER TABLE users ADD COLUMN email VARCHAR(255);

--changeset bob:3
--comment: liquibase-linter:disable sql-injection
--comment: liquibase-linter:disable hardcoded-credentials
--comment: This setup is for test environment only
CREATE USER 'testuser'@'localhost' IDENTIFIED BY 'password123';
GRANT ALL ON *.* TO 'testuser'@'localhost';
INSERT INTO users VALUES (2, ${newuser});

--changeset alice:4
--comment: First comment line
--comment: Second comment line
--comment: Third comment line
CREATE TABLE orders (id INT PRIMARY KEY);

--changeset charlie:5
--comment: liquibase-linter:disable sql-injection,missing-rollback
--comment: Batch update for legacy data migration
UPDATE users SET username = CONCAT('user_', ${suffix}) WHERE id > 100;

--changeset dave:6
--comment: Regular single comment
CREATE TABLE products (id INT PRIMARY KEY, name VARCHAR(100));

--changeset eve:7
--comment: liquibase-linter:disable sql-injection
--comment: Additional context about this change
--comment: liquibase-linter:disable missing-rollback
--comment: Even more documentation here
DELETE FROM users WHERE username = ${oldname};
