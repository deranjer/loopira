-- +goose Up
CREATE TABLE issue_history (
    id         uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    issue_id   uuid NOT NULL REFERENCES issues(id) ON DELETE CASCADE,
    actor_id   uuid REFERENCES users(id) ON DELETE SET NULL,
    action     text NOT NULL CHECK (action IN ('created', 'updated')),
    changes    jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_issue_history_issue_created
    ON issue_history (issue_id, created_at DESC, id DESC);

-- Existing issues predate the audit stream. Preserve their original creator
-- and timestamp so every issue starts with a useful history entry.
INSERT INTO issue_history (issue_id, actor_id, action, changes, created_at)
SELECT id, created_by, 'created', '{}'::jsonb, created_at
FROM issues;

-- +goose Down
DROP TABLE IF EXISTS issue_history;
