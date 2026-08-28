-- name: CreateLogEntry :one
INSERT INTO log_entries (food_id, multiplier, logged_at)
VALUES ($1, $2, COALESCE(sqlc.narg('logged_at')::date, CURRENT_DATE))
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
    COALESCE(SUM(f.calories * l.multiplier), 0)::numeric AS calories,
    COALESCE(SUM(f.protein_g * l.multiplier), 0)::numeric AS protein_g,
    COALESCE(SUM(f.carbs_g * l.multiplier), 0)::numeric AS carbs_g,
    COALESCE(SUM(f.fat_g * l.multiplier), 0)::numeric AS fat_g
FROM log_entries l
JOIN foods f ON f.id = l.food_id
WHERE l.logged_at = $1;

-- name: GetMacroTotalsByWeek :one
SELECT
    COALESCE(SUM(f.calories * l.multiplier), 0)::numeric AS calories,
    COALESCE(SUM(f.protein_g * l.multiplier), 0)::numeric AS protein_g,
    COALESCE(SUM(f.carbs_g * l.multiplier), 0)::numeric AS carbs_g,
    COALESCE(SUM(f.fat_g * l.multiplier), 0)::numeric AS fat_g
FROM log_entries l
JOIN foods f ON f.id = l.food_id
WHERE l.logged_at BETWEEN sqlc.arg('logged_at')::date - INTERVAL '6 days'
                      AND sqlc.arg('logged_at')::date;

-- name: GetDailyMacroTotalsByWeek :many
SELECT
    d.day::date AS day,
    COALESCE(SUM(f.calories * l.multiplier), 0)::numeric AS calories,
    COALESCE(SUM(f.protein_g * l.multiplier), 0)::numeric AS protein_g,
    COALESCE(SUM(f.carbs_g * l.multiplier), 0)::numeric AS carbs_g,
    COALESCE(SUM(f.fat_g * l.multiplier), 0)::numeric AS fat_g
FROM generate_series(
    sqlc.arg('logged_at')::date - INTERVAL '6 days',
    sqlc.arg('logged_at')::date,
    INTERVAL '1 day'
) AS d(day)
LEFT JOIN log_entries l ON l.logged_at = d.day::date
LEFT JOIN foods f ON f.id = l.food_id
GROUP BY d.day
ORDER BY d.day;
