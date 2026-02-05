--liquibase formatted sql

--changeset john:1
CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
--rollback DROP TABLE users;

--changeset john:2 labels:index,performance
CREATE INDEX idx_users_email ON users(email);
--rollback DROP INDEX idx_users_email ON users;

--changeset jane:3 context:dev
--comment: Add test data for development
INSERT INTO users (username, email) VALUES ('testuser', 'test@example.com');
--rollback DELETE FROM users WHERE username = 'testuser';
