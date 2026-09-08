-- name: CreateIssueHistory :one
INSERT INTO issue_history (issue_id, actor_id, action, changes)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListIssueHistory :many
SELECT
    h.id,
    h.issue_id,
    h.actor_id,
    COALESCE(u.name, 'Deleted user') AS actor_name,
    h.action,
    h.changes,
    h.created_at
FROM issue_history h
LEFT JOIN users u ON u.id = h.actor_id
WHERE h.issue_id = $1
ORDER BY h.created_at DESC, h.id DESC;
