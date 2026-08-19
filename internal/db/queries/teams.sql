-- name: ListTeams :many
SELECT * FROM teams ORDER BY name;

-- name: GetTeam :one
SELECT * FROM teams WHERE id = $1;

-- name: CreateTeam :one
INSERT INTO teams (name, key)
VALUES ($1, $2)
RETURNING *;

-- name: UpdateTeamName :one
UPDATE teams SET name = $2 WHERE id = $1
RETURNING *;

-- name: AddTeamMember :exec
INSERT INTO team_members (team_id, user_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;
