-- name: ListLabels :many
SELECT * FROM labels
WHERE team_id = $1 OR team_id IS NULL
ORDER BY name;

-- name: CreateLabel :one
INSERT INTO labels (team_id, name, color)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ClearIssueLabels :exec
DELETE FROM issue_labels WHERE issue_id = $1;

-- name: AddIssueLabel :exec
INSERT INTO issue_labels (issue_id, label_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;
