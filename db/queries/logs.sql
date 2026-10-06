-- name: CreateLogEntry :one
INSERT INTO log_entries (food_id, food_name, calories, protein_g, carbs_g, fat_g, multiplier, logged_at)
SELECT $1, f.name, f.calories, f.protein_g, f.carbs_g, f.fat_g, $2,
       COALESCE(sqlc.narg('logged_at')::date, CURRENT_DATE)
FROM foods f
WHERE f.id = $1
RETURNING *;

-- name: GetLogEntry :one
SELECT * FROM log_entries where id = $1;

-- name: DeleteLogEntry :exec
DELETE FROM log_entries WHERE id = $1;

-- name: ListLogEntries :many
SELECT * FROM log_entries ORDER BY logged_at DESC, id;

-- name: ListLogEntriesByDate :many
SELECT * FROM log_entries
WHERE logged_at = $1
ORDER BY id;

-- name: GetMacroTotalsByDate :one
SELECT
    COALESCE(SUM(calories * multiplier), 0)::numeric AS calories,
    COALESCE(SUM(protein_g * multiplier), 0)::numeric AS protein_g,
    COALESCE(SUM(carbs_g * multiplier), 0)::numeric AS carbs_g,
    COALESCE(SUM(fat_g * multiplier), 0)::numeric AS fat_g
FROM log_entries
WHERE logged_at = $1;
