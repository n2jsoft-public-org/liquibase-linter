--liquibase formatted sql

--changeset test:1 labels:v1 context:local
--comment: liquibase-linter:disable sql-injection
UPDATE users
SET name = CONCAT('prefix_', name)
WHERE id = 1;
