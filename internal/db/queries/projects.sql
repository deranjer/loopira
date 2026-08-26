-- name: ListProjects :many
SELECT
    p.*,
    u.name AS lead_name,
    t.name AS template_name,
    count(i.id)::int AS issue_count,
    count(i.id) FILTER (WHERE i.status = 'done')::int AS done_count
FROM projects p
LEFT JOIN users u ON u.id = p.lead_id
LEFT JOIN templates t ON t.id = p.template_id
LEFT JOIN issues i ON i.project_id = p.id
WHERE p.team_id = $1
GROUP BY p.id, u.name, t.name
ORDER BY p.created_at;

-- name: CreateProject :one
INSERT INTO projects (team_id, name, description, status, lead_id, priority, target_date, template_id)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetProject :one
SELECT
    p.*,
    u.name AS lead_name,
    t.name AS template_name,
    count(i.id)::int AS issue_count,
    count(i.id) FILTER (WHERE i.status = 'done')::int AS done_count
FROM projects p
LEFT JOIN users u ON u.id = p.lead_id
LEFT JOIN templates t ON t.id = p.template_id
LEFT JOIN issues i ON i.project_id = p.id
WHERE p.id = $1
GROUP BY p.id, u.name, t.name;

-- name: UpdateProject :one
UPDATE projects SET
    name = $2,
    description = $3,
    status = $4,
    lead_id = $5,
    priority = $6,
    target_date = $7
WHERE id = $1
RETURNING *;
