-- +goose Up
ALTER TABLE foods ADD COLUMN is_frequent BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE foods DROP COLUMN is_frequent;
