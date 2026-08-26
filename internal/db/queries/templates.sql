-- name: ListTemplates :many
SELECT t.*, u.name AS author_name
FROM templates t
LEFT JOIN users u ON u.id = t.created_by
ORDER BY t.name;

-- name: GetTemplate :one
SELECT t.*, u.name AS author_name
FROM templates t
LEFT JOIN users u ON u.id = t.created_by
WHERE t.id = $1;

-- name: ListTemplateLinks :many
SELECT
    tfl.template_id,
    tfl.fragment_id,
    tfl.position,
    f.name,
    f.category
FROM template_fragment_links tfl
JOIN template_fragments f ON f.id = tfl.fragment_id
WHERE tfl.template_id = $1
ORDER BY tfl.position;

-- name: ListTemplateFragmentsForStamp :many
-- Full fragment content for every fragment linked to a template, in
-- composition order — used when a Project is created from this template.
SELECT f.*, tfl.position
FROM template_fragment_links tfl
JOIN template_fragments f ON f.id = tfl.fragment_id
WHERE tfl.template_id = $1
ORDER BY tfl.position;

-- name: CreateTemplate :one
INSERT INTO templates (name, description, created_by)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateTemplate :one
UPDATE templates SET
    name = $2,
    description = $3,
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteTemplate :exec
DELETE FROM templates WHERE id = $1;

-- name: AddTemplateFragment :exec
INSERT INTO template_fragment_links (template_id, fragment_id, position)
VALUES (
    $1,
    $2,
    COALESCE((SELECT max(position) + 1 FROM template_fragment_links WHERE template_id = $1), 0)
)
ON CONFLICT DO NOTHING;

-- name: RemoveTemplateFragment :exec
DELETE FROM template_fragment_links WHERE template_id = $1 AND fragment_id = $2;

-- name: SetTemplateFragmentPosition :exec
UPDATE template_fragment_links SET position = $3
WHERE template_id = $1 AND fragment_id = $2;
