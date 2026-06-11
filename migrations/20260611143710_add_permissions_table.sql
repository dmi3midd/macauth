-- +goose Up
ALTER TABLE users DROP COLUMN is_admin;
CREATE TABLE permissions (
    id VARCHAR(20) PRIMARY KEY,
    user_id VARCHAR(20) NOT NULL,
    client_id VARCHAR(64) NOT NULL,
    permission VARCHAR(64) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
ALTER TABLE users ADD COLUMN is_admin BOOLEAN NOT NULL DEFAULT false;
DROP TABLE permissions;
