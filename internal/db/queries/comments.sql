-- name: CreateComment :one
INSERT INTO comments (issue_id, author_id, body)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListIssueComments :many
SELECT c.*, u.name AS author_name
FROM comments c
JOIN users u ON u.id = c.author_id
WHERE c.issue_id = $1
ORDER BY c.created_at ASC, c.id ASC;

-- name: GetComment :one
SELECT c.*, u.name AS author_name
FROM comments c
JOIN users u ON u.id = c.author_id
WHERE c.id = $1;
