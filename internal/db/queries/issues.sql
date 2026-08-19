-- name: ListIssues :many
SELECT
    i.*,
    t.key AS team_key,
    u.name AS assignee_name,
    p.name AS project_name,
    lbl.id AS label_id,
    COALESCE(lbl.name, '') AS label_name,
    COALESCE(lbl.color, '') AS label_color
FROM issues i
JOIN teams t ON t.id = i.team_id
LEFT JOIN users u ON u.id = i.assignee_id
LEFT JOIN projects p ON p.id = i.project_id
LEFT JOIN LATERAL (
    SELECT l.id, l.name, l.color
    FROM issue_labels il
    JOIN labels l ON l.id = il.label_id
    WHERE il.issue_id = i.id
    ORDER BY l.name
    LIMIT 1
) lbl ON true
WHERE i.team_id = $1
  AND (sqlc.narg(status)::text IS NULL OR i.status = sqlc.narg(status))
  AND (sqlc.narg(project_id)::uuid IS NULL OR i.project_id = sqlc.narg(project_id))
  AND (sqlc.narg(cycle_id)::uuid IS NULL OR i.cycle_id = sqlc.narg(cycle_id))
ORDER BY i.created_at DESC;

-- name: GetIssue :one
SELECT
    i.*,
    t.key AS team_key,
    u.name AS assignee_name,
    p.name AS project_name,
    lbl.id AS label_id,
    COALESCE(lbl.name, '') AS label_name,
    COALESCE(lbl.color, '') AS label_color
FROM issues i
JOIN teams t ON t.id = i.team_id
LEFT JOIN users u ON u.id = i.assignee_id
LEFT JOIN projects p ON p.id = i.project_id
LEFT JOIN LATERAL (
    SELECT l.id, l.name, l.color
    FROM issue_labels il
    JOIN labels l ON l.id = il.label_id
    WHERE il.issue_id = i.id
    ORDER BY l.name
    LIMIT 1
) lbl ON true
WHERE i.id = $1;

-- name: GetIssueByNumber :one
SELECT
    i.*,
    t.key AS team_key,
    u.name AS assignee_name,
    p.name AS project_name,
    lbl.id AS label_id,
    COALESCE(lbl.name, '') AS label_name,
    COALESCE(lbl.color, '') AS label_color
FROM issues i
JOIN teams t ON t.id = i.team_id
LEFT JOIN users u ON u.id = i.assignee_id
LEFT JOIN projects p ON p.id = i.project_id
LEFT JOIN LATERAL (
    SELECT l.id, l.name, l.color
    FROM issue_labels il
    JOIN labels l ON l.id = il.label_id
    WHERE il.issue_id = i.id
    ORDER BY l.name
    LIMIT 1
) lbl ON true
WHERE i.team_id = $1 AND i.number = $2;

-- name: CreateIssue :one
WITH next_num AS (
    UPDATE teams SET next_issue_number = next_issue_number + 1
    WHERE id = $1
    RETURNING next_issue_number - 1 AS number
)
INSERT INTO issues (team_id, number, title, description, priority, assignee_id, project_id, created_by)
SELECT $1, next_num.number, $2, $3, $4, $5, $6, $7
FROM next_num
RETURNING *;

-- name: UpdateIssueStatus :one
UPDATE issues SET status = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateIssueDetails :one
UPDATE issues SET
    title = $2,
    description = $3,
    priority = $4,
    assignee_id = $5,
    project_id = $6,
    cycle_id = $7,
    updated_at = now()
WHERE id = $1
RETURNING *;
