-- +goose Up
CREATE TABLE log_entries (
    id         SERIAL PRIMARY KEY,
    food_id    INTEGER NOT NULL REFERENCES foods(id) ON DELETE CASCADE,
    multiplier NUMERIC NOT NULL DEFAULT 1,
    logged_at  DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE log_entries;
