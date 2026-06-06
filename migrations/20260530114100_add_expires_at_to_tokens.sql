-- +goose Up
ALTER TABLE tokens ADD COLUMN expires_at TIMESTAMP NOT NULL DEFAULT (NOW() + INTERVAL '14 days');

-- +goose Down
ALTER TABLE tokens DROP COLUMN expires_at;
