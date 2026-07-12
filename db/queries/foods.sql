-- name: CreateFood :one
INSERT INTO foods (name, calories, protein_g, carbs_g, fat_g)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetFood :one
SELECT * FROM foods WHERE id = $1;

-- name: ListFoods :many
SELECT * FROM foods ORDER BY id;

-- name: DeleteFood :exec
DELETE FROM foods WHERE id = $1;
