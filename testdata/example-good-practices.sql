--liquibase formatted sql

--changeset john:1-create-users-table
--comment: Create the users table for authentication
CREATE TABLE users (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
--rollback DROP TABLE users;

--changeset john:2-create-roles-table
--comment: Create roles table for RBAC
CREATE TABLE roles (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
--rollback DROP TABLE roles;

--changeset john:3-create-user-roles-table
--comment: Many-to-many relationship between users and roles
CREATE TABLE user_roles (
    user_id BIGINT NOT NULL,
    role_id BIGINT NOT NULL,
    assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);
--rollback DROP TABLE user_roles;

--changeset john:4-add-user-roles-indexes
--comment: Add indexes for better query performance
CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);
--rollback DROP INDEX idx_user_roles_role_id; DROP INDEX idx_user_roles_user_id;

--changeset john:5-add-default-roles
--comment: Insert default system roles
INSERT INTO roles (name, description) VALUES 
    ('ADMIN', 'System administrator with full access'),
    ('USER', 'Standard user with basic permissions'),
    ('GUEST', 'Guest user with read-only access');
--rollback DELETE FROM roles WHERE name IN ('ADMIN', 'USER', 'GUEST');

--changeset jane:6-add-user-status-column
--comment: Add status column to track active/inactive users
--preconditions onFail:MARK_RAN
--precondition-sql-check expectedResult:0 SELECT COUNT(*) FROM information_schema.columns WHERE table_name='users' AND column_name='status'
ALTER TABLE users ADD COLUMN status VARCHAR(20) DEFAULT 'active' NOT NULL;
--rollback ALTER TABLE users DROP COLUMN status;

--changeset jane:7-add-email-verification
--comment: Add email verification tracking
ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT FALSE NOT NULL;
ALTER TABLE users ADD COLUMN verification_token VARCHAR(100);
ALTER TABLE users ADD COLUMN verification_sent_at TIMESTAMP;
--rollback ALTER TABLE users DROP COLUMN verification_sent_at; ALTER TABLE users DROP COLUMN verification_token; ALTER TABLE users DROP COLUMN email_verified;

--changeset jane:8-create-audit-log-table
--comment: Create audit log for tracking user actions
CREATE TABLE audit_log (
    id BIGINT PRIMARY KEY AUTO_INCREMENT,
    user_id BIGINT,
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(50),
    entity_id BIGINT,
    details TEXT,
    ip_address VARCHAR(45),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL,
    INDEX idx_audit_log_user_id (user_id),
    INDEX idx_audit_log_created_at (created_at),
    INDEX idx_audit_log_action (action)
);
--rollback DROP TABLE audit_log;

--changeset bob:9-add-password-reset
--comment: Add password reset token functionality
ALTER TABLE users ADD COLUMN reset_token VARCHAR(100);
ALTER TABLE users ADD COLUMN reset_token_expires_at TIMESTAMP;
CREATE INDEX idx_users_reset_token ON users(reset_token);
--rollback DROP INDEX idx_users_reset_token; ALTER TABLE users DROP COLUMN reset_token_expires_at; ALTER TABLE users DROP COLUMN reset_token;

--changeset bob:10-create-sessions-table
--comment: Track active user sessions
CREATE TABLE user_sessions (
    id VARCHAR(100) PRIMARY KEY,
    user_id BIGINT NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_sessions_user_id (user_id),
    INDEX idx_sessions_expires_at (expires_at)
);
--rollback DROP TABLE user_sessions;
