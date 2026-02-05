--liquibase formatted sql

--changeset developer:001
CREATE TABLE users (
    id INT PRIMARY KEY,
    username VARCHAR(50),
    email VARCHAR(100)
);

--changeset developer:002
CREATE TABLE orders (
    id INT PRIMARY KEY,
    user_id INT,
    total DECIMAL(10,2)
);

--changeset developer:001
-- VIOLATION: Duplicate changeset ID and author
CREATE TABLE products (
    id INT PRIMARY KEY,
    name VARCHAR(100)
);

--changeset developer:003
ALTER TABLE users ADD COLUMN created_at TIMESTAMP;

--changeset another-dev:001
-- OK: Same ID but different author
CREATE TABLE categories (
    id INT PRIMARY KEY,
    name VARCHAR(50)
);

--changeset developer:002
-- VIOLATION: Another duplicate changeset
CREATE TABLE inventory (
    id INT,
    product_id INT
);
