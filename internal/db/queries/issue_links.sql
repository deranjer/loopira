-- name: SetIssueParent :one
UPDATE issues SET parent_id = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- IsIssueInAncestry reports whether sqlc.arg(issue_id) is the issue
-- sqlc.arg(start_id) or one of its ancestors. Used to reject parent cycles:
-- the proposed parent must not descend from the issue being re-parented.
-- name: IsIssueInAncestry :one
WITH RECURSIVE ancestry AS (
    SELECT issues.id, issues.parent_id FROM issues WHERE issues.id = sqlc.arg(start_id)
    UNION
    SELECT i.id, i.parent_id FROM issues i JOIN ancestry a ON i.id = a.parent_id
)
SELECT (sqlc.arg(issue_id)::uuid IN (SELECT ancestry.id FROM ancestry))::bool AS found;

-- name: ListIssueLinks :many
SELECT kind, id, team_key, number, title, status FROM (
    SELECT 'parent'::text AS kind, p.id, t.key AS team_key, p.number, p.title, p.status
    FROM issues c
    JOIN issues p ON p.id = c.parent_id
    JOIN teams t ON t.id = p.team_id
    WHERE c.id = $1
  UNION ALL
    SELECT 'child'::text, c.id, t.key, c.number, c.title, c.status
    FROM issues c
    JOIN teams t ON t.id = c.team_id
    WHERE c.parent_id = $1
  UNION ALL
    SELECT 'blocked_by'::text, b.id, t.key, b.number, b.title, b.status
    FROM issue_relations r
    JOIN issues b ON b.id = r.issue_id
    JOIN teams t ON t.id = b.team_id
    WHERE r.related_issue_id = $1 AND r.type = 'blocks'
  UNION ALL
    SELECT 'blocks'::text, b.id, t.key, b.number, b.title, b.status
    FROM issue_relations r
    JOIN issues b ON b.id = r.related_issue_id
    JOIN teams t ON t.id = b.team_id
    WHERE r.issue_id = $1 AND r.type = 'blocks'
) links
ORDER BY kind, number;

-- name: AddIssueBlocker :execrows
INSERT INTO issue_relations (issue_id, related_issue_id, type)
VALUES (sqlc.arg(blocker_id), sqlc.arg(blocked_id), 'blocks')
ON CONFLICT (issue_id, related_issue_id, type) DO NOTHING;

-- name: RemoveIssueBlocker :execrows
DELETE FROM issue_relations
WHERE issue_id = sqlc.arg(blocker_id) AND related_issue_id = sqlc.arg(blocked_id) AND type = 'blocks';

-- DoesIssueBlockTransitively reports whether sqlc.arg(from_id) already blocks
-- sqlc.arg(to_id), directly or through a chain. Adding the reverse edge
-- would create a dependency cycle.
-- name: DoesIssueBlockTransitively :one
WITH RECURSIVE chain AS (
    SELECT ir.related_issue_id AS id FROM issue_relations ir WHERE ir.issue_id = sqlc.arg(from_id) AND ir.type = 'blocks'
    UNION
    SELECT r.related_issue_id FROM issue_relations r JOIN chain c ON r.issue_id = c.id WHERE r.type = 'blocks'
)
SELECT (sqlc.arg(to_id)::uuid IN (SELECT chain.id FROM chain))::bool AS found;
