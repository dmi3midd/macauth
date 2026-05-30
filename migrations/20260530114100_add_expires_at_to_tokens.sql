-- +goose Up
ALTER TABLE tokens ADD COLUMN expires_at TIMESTAMP NOT NULL DEFAULT (datetime('now', '+14 days'));

-- +goose Down
ALTER TABLE tokens DROP COLUMN expires_at;
