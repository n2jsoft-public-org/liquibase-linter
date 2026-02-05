--liquibase formatted sql

--changeset john:1
--comment: liquibase-linter:disable sql-injection
CREATE TABLE users (
    id INT PRIMARY KEY,
    name VARCHAR(100)
);
INSERT INTO users VALUES (1, 'test' + ${username});

--changeset jane:2
--comment: liquibase-linter:disable sql-injection,missing-rollback
UPDATE users SET name = CONCAT('user_', ${newname});

--changeset bob:3
--comment: Regular comment without suppression
DELETE FROM users WHERE name = ${oldname};

--changeset alice:4
--comment: LIQUIBASE-LINTER:DISABLE SQL-INJECTION
SELECT * FROM users WHERE id = ${id};

--changeset charlie:5
--comment: Test account creation. liquibase-linter:disable hardcoded-credentials for test environment
CREATE USER 'testuser'@'localhost' IDENTIFIED BY 'testpass123';

--changeset dave:6
--comment: liquibase-linter:disable invalid-rule-id
CREATE TABLE dummy (id INT);
