-- name: ListLabels :many
SELECT * FROM labels
WHERE team_id = $1 OR team_id IS NULL
ORDER BY name;
