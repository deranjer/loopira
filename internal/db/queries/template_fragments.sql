-- name: ListTemplateFragments :many
SELECT f.*, u.name AS author_name
FROM template_fragments f
LEFT JOIN users u ON u.id = f.created_by
ORDER BY f.name;

-- name: GetTemplateFragment :one
SELECT f.*, u.name AS author_name
FROM template_fragments f
LEFT JOIN users u ON u.id = f.created_by
WHERE f.id = $1;

-- name: CreateTemplateFragment :one
INSERT INTO template_fragments (name, category, content, created_by)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateTemplateFragment :one
UPDATE template_fragments SET
    name = $2,
    category = $3,
    content = $4,
    version = version + 1,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteTemplateFragment :exec
DELETE FROM template_fragments WHERE id = $1;

-- name: ListFragmentUsage :many
SELECT
    pgf.id AS project_guide_fragment_id,
    p.id AS project_id,
    p.name AS project_name,
    pgf.locally_modified,
    pgf.base_version
FROM project_guide_fragments pgf
JOIN projects p ON p.id = pgf.project_id
WHERE pgf.fragment_id = $1
ORDER BY p.name;
