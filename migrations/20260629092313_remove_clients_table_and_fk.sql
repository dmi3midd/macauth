-- +goose Up
ALTER TABLE tokens DROP CONSTRAINT tokens_client_id_fkey;
DROP TABLE clients;

-- +goose Down
CREATE TABLE clients (
    id VARCHAR(20) PRIMARY KEY,
    name VARCHAR(255) NOT NULL UNIQUE,
    hashed_secret VARCHAR(255) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE tokens ADD CONSTRAINT tokens_client_id_fkey FOREIGN KEY (client_id) REFERENCES clients(id) ON DELETE CASCADE;

