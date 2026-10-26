--liquibase formatted sql

--changeset alice:no-dbms
CREATE TABLE a (id INT);

--changeset alice:typo dbms:postgres
CREATE TABLE b (id INT);

--changeset alice:catalog-on-pg dbms:postgresql
--preconditions onFail:MARK_RAN
--precondition-sql-check expectedResult:0 SELECT COUNT(*) FROM sys.tables WHERE name = 'c'
CREATE TABLE c (id INT);

--changeset alice:catalog-on-mssql dbms:mssql
--precondition-sql-check expectedResult:0 SELECT COUNT(*) FROM sys.tables WHERE name = 'd'
CREATE TABLE d (id INT);

--changeset alice:portable dbms:mssql,postgresql
--precondition-sql-check expectedResult:0 SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'e'
CREATE TABLE e (id INT);
