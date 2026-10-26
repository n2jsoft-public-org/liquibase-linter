--liquibase formatted sql

--changeset alice:1 dbms:mssql
ALTER TABLE Users ADD Email NVARCHAR(255);

--changeset alice:1-pg dbms:postgresql
ALTER TABLE users ADD COLUMN email VARCHAR(255);

--changeset alice:2 dbms:mssql
ALTER TABLE Users ADD Phone NVARCHAR(50);

--changeset alice:3 dbms:mssql
--comment: liquibase-linter:disable dbms-twin SQL Server full-text index, no PostgreSQL equivalent
CREATE FULLTEXT INDEX ON Users(Name) KEY INDEX PK_Users;

--changeset alice:4 dbms:mssql
ALTER TABLE Users ADD Age INT;

--changeset alice:4-pg dbms:mssql
ALTER TABLE users ADD COLUMN age INT;
