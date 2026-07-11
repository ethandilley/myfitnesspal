-- +goose Up
CREATE TABLE foods (
    id         SERIAL PRIMARY KEY,
    name       TEXT NOT NULL UNIQUE,
    calories   NUMERIC NOT NULL,
    protein_g  NUMERIC NOT NULL,
    carbs_g    NUMERIC NOT NULL,
    fat_g      NUMERIC NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE foods;
