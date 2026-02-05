--liquibase formatted sql

--changeset admin:1
-- Missing rollback for this dangerous operation
DROP TABLE important_data;

--changeset admin:2
-- SQL injection vulnerability
DELETE FROM users WHERE username = '${username}';

--changeset admin:3 runAlways:true
-- Dangerous operation without preconditions
TRUNCATE TABLE logs;
