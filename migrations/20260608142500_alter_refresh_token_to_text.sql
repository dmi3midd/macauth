-- +goose Up
ALTER TABLE tokens ALTER COLUMN refresh_token TYPE TEXT;

-- +goose Down
ALTER TABLE tokens ALTER COLUMN refresh_token TYPE VARCHAR(255);
