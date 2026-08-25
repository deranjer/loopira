-- name: ListViews :many
SELECT * FROM views WHERE owner_id = $1 OR shared = true ORDER BY created_at;

-- name: CreateView :one
INSERT INTO views (owner_id, name, definition, shared)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateView :one
UPDATE views SET name = $3, definition = $4, shared = $5
WHERE id = $1 AND owner_id = $2
RETURNING *;

-- name: DeleteView :exec
DELETE FROM views WHERE id = $1 AND owner_id = $2;
