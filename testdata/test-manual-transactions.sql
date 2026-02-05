--liquibase formatted sql

--changeset test:1
INSERT INTO users (name, email) VALUES ('John', 'john@example.com');

--changeset test:2
BEGIN TRANSACTION;
UPDATE users SET active = 1;
COMMIT;

--changeset test:3
START TRANSACTION;
INSERT INTO logs (message) VALUES ('test');
COMMIT TRANSACTION;

--changeset test:4
UPDATE accounts SET balance = 0;
ROLLBACK;

--changeset test:5 context:procedure
--comment: This is excluded because it's a procedure
CREATE PROCEDURE UpdateUsers
AS
BEGIN
    BEGIN TRANSACTION;
    UPDATE users SET status = 'active';
    COMMIT TRANSACTION;
END;

--changeset test:6
-- Comment with BEGIN TRANSACTION should not trigger
INSERT INTO users (name) VALUES ('Alice');
