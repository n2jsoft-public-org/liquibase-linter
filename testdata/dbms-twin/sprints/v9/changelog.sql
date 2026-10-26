--liquibase formatted sql

--changeset alice:1 dbms:mssql
ALTER TABLE Users ADD Nickname NVARCHAR(50);
