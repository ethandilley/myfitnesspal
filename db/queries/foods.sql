-- name: CreateFood :one
INSERT INTO foods (name, calories, protein_g, carbs_g, fat_g, is_frequent)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetFood :one
SELECT * FROM foods WHERE id = $1;

-- name: ListFoods :many
SELECT * FROM foods
WHERE sqlc.narg('is_frequent')::boolean IS NULL
   OR is_frequent = sqlc.narg('is_frequent')
ORDER BY id;

-- name: SetFoodFrequent :one
UPDATE foods SET is_frequent = $2 WHERE id = $1
RETURNING *;

-- name: DeleteFood :exec
DELETE FROM foods WHERE id = $1;
