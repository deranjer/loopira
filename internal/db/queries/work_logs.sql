-- name: CreateWorkLog :one
INSERT INTO work_logs (project_id, author_id, source, title, body)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetWorkLog :one
SELECT w.*, u.name AS author_name, p.name AS project_name
FROM work_logs w
JOIN users u ON u.id = w.author_id
JOIN projects p ON p.id = w.project_id
WHERE w.id = $1;

-- name: ListProjectWorkLogs :many
SELECT w.*, u.name AS author_name, p.name AS project_name
FROM work_logs w
JOIN users u ON u.id = w.author_id
JOIN projects p ON p.id = w.project_id
WHERE w.project_id = $1
ORDER BY w.created_at DESC;

-- name: ListWorkLogs :many
SELECT w.*, u.name AS author_name, p.name AS project_name
FROM work_logs w
JOIN users u ON u.id = w.author_id
JOIN projects p ON p.id = w.project_id
WHERE
    (sqlc.narg(project_id)::uuid IS NULL OR w.project_id = sqlc.narg(project_id))
    AND (sqlc.narg(author_id)::uuid IS NULL OR w.author_id = sqlc.narg(author_id))
    AND (sqlc.narg(source)::text IS NULL OR w.source = sqlc.narg(source))
    AND (sqlc.narg(created_from)::timestamptz IS NULL OR w.created_at >= sqlc.narg(created_from))
    AND (sqlc.narg(created_to)::timestamptz IS NULL OR w.created_at < sqlc.narg(created_to))
    AND (
        sqlc.narg(search)::text IS NULL
        OR to_tsvector('english', w.title || ' ' || w.body) @@ websearch_to_tsquery('english', sqlc.narg(search))
    )
ORDER BY w.created_at DESC
LIMIT sqlc.arg(limit_count) OFFSET sqlc.arg(offset_count);

-- name: CountWorkLogs :one
SELECT count(*)::int AS total
FROM work_logs w
WHERE
    (sqlc.narg(project_id)::uuid IS NULL OR w.project_id = sqlc.narg(project_id))
    AND (sqlc.narg(author_id)::uuid IS NULL OR w.author_id = sqlc.narg(author_id))
    AND (sqlc.narg(source)::text IS NULL OR w.source = sqlc.narg(source))
    AND (sqlc.narg(created_from)::timestamptz IS NULL OR w.created_at >= sqlc.narg(created_from))
    AND (sqlc.narg(created_to)::timestamptz IS NULL OR w.created_at < sqlc.narg(created_to))
    AND (
        sqlc.narg(search)::text IS NULL
        OR to_tsvector('english', w.title || ' ' || w.body) @@ websearch_to_tsquery('english', sqlc.narg(search))
    );
