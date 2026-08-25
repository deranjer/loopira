-- name: ListProjectAttachments :many
SELECT a.*, u.name AS uploaded_by_name
FROM attachments a
JOIN users u ON u.id = a.uploaded_by
WHERE a.project_id = $1
ORDER BY a.created_at DESC;

-- name: CreateAttachment :one
INSERT INTO attachments (project_id, filename, content_type, size_bytes, storage_key, uploaded_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetAttachment :one
SELECT * FROM attachments WHERE id = $1;

-- name: DeleteAttachment :exec
DELETE FROM attachments WHERE id = $1;
