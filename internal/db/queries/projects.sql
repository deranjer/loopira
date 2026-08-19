-- name: ListProjects :many
SELECT
    p.*,
    count(i.id)::int AS issue_count,
    count(i.id) FILTER (WHERE i.status = 'done')::int AS done_count
FROM projects p
LEFT JOIN issues i ON i.project_id = p.id
WHERE p.team_id = $1
GROUP BY p.id
ORDER BY p.created_at;

-- name: CreateProject :one
INSERT INTO projects (team_id, name, description, status)
VALUES ($1, $2, $3, 'planned')
RETURNING *;

-- name: GetProject :one
SELECT
    p.*,
    count(i.id)::int AS issue_count,
    count(i.id) FILTER (WHERE i.status = 'done')::int AS done_count
FROM projects p
LEFT JOIN issues i ON i.project_id = p.id
WHERE p.id = $1
GROUP BY p.id;
