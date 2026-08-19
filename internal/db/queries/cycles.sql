-- name: ListCycles :many
SELECT
    c.*,
    count(i.id)::int AS issue_count,
    count(i.id) FILTER (WHERE i.status = 'done')::int AS done_count,
    (current_date BETWEEN c.start_date AND c.end_date)::boolean AS active
FROM cycles c
LEFT JOIN issues i ON i.cycle_id = c.id
WHERE c.team_id = $1
GROUP BY c.id
ORDER BY c.number DESC;

-- name: CreateCycle :one
INSERT INTO cycles (team_id, number, start_date, end_date)
SELECT $1, COALESCE(MAX(number), 0) + 1, $2, $3
FROM cycles WHERE team_id = $1
RETURNING *;
