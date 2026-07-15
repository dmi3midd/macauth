-- +goose Up
CREATE TABLE tokens (
    id VARCHAR(20) PRIMARY KEY,
    refresh_token TEXT NOT NULL,
    user_id VARCHAR(20) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE tokens;