-- +goose Up
CREATE TABLE tokens (
    id VARCHAR(20) PRIMARY KEY,
    refresh_token VARCHAR(255) NOT NULL,
    user_id VARCHAR(20) NOT NULL,
    client_id VARCHAR(20) NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE tokens;
