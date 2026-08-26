-- name: ListProjectGuideFragments :many
SELECT * FROM project_guide_fragments
WHERE project_id = $1
ORDER BY position;

-- name: AddProjectGuideFragment :one
INSERT INTO project_guide_fragments (project_id, fragment_id, name, content, base_version, position)
VALUES (
    $1,
    $2,
    $3,
    $4,
    $5,
    COALESCE((SELECT max(position) + 1 FROM project_guide_fragments WHERE project_id = $1), 0)
)
RETURNING *;

-- name: UpdateProjectGuideFragment :one
UPDATE project_guide_fragments SET
    name = $2,
    content = $3,
    locally_modified = (fragment_id IS NOT NULL),
    updated_at = now()
WHERE id = $1
RETURNING *;

-- name: DeleteProjectGuideFragment :exec
DELETE FROM project_guide_fragments WHERE id = $1;

-- name: SetProjectGuideFragmentPosition :exec
UPDATE project_guide_fragments SET position = $2
WHERE id = $1;

-- name: ResetProjectGuideFragmentToBase :one
UPDATE project_guide_fragments pgf SET
    name = tf.name,
    content = tf.content,
    base_version = tf.version,
    locally_modified = false,
    updated_at = now()
FROM template_fragments tf
WHERE pgf.id = $1 AND pgf.fragment_id = tf.id
RETURNING pgf.*;

-- name: PushFragmentUpdate :many
UPDATE project_guide_fragments SET
    name = $2,
    content = $3,
    base_version = $4,
    locally_modified = false,
    updated_at = now()
WHERE fragment_id = $1 AND id = ANY(@ids::uuid[]) AND locally_modified = false
RETURNING *;
