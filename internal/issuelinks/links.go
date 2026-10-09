// Package issuelinks holds the validation rules for issue hierarchy
// (parent/child) and blocking dependencies, shared by the REST API and the
// MCP tools so both enforce the same invariants.
package issuelinks

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/deranjer/loopira/internal/db"
)

// ErrInvalid marks a request that violates a link rule (self-reference,
// cross-team link, cycle). Its message is safe to show to callers.
var ErrInvalid = errors.New("invalid issue link")

type invalidError string

func (e invalidError) Error() string { return string(e) }
func (e invalidError) Unwrap() error { return ErrInvalid }

// SetParent re-parents issue under parent, or detaches it when parent is
// not valid. The parent must be in the same team and must not be the issue
// itself or one of its descendants.
func SetParent(ctx context.Context, q *db.Queries, issue db.GetIssueRow, parent pgtype.UUID) (db.Issue, error) {
	if parent.Valid {
		if parent == issue.ID {
			return db.Issue{}, invalidError("an issue can't be its own parent")
		}
		p, err := q.GetIssue(ctx, parent)
		if err != nil {
			return db.Issue{}, invalidError("parent issue not found")
		}
		if p.TeamID != issue.TeamID {
			return db.Issue{}, invalidError("parent must be in the same team")
		}
		cycle, err := q.IsIssueInAncestry(ctx, db.IsIssueInAncestryParams{StartID: parent, IssueID: issue.ID})
		if err != nil {
			return db.Issue{}, err
		}
		if cycle {
			return db.Issue{}, invalidError("that would make an issue a descendant of itself")
		}
	}
	return q.SetIssueParent(ctx, db.SetIssueParentParams{ID: issue.ID, ParentID: parent})
}

// AddBlocker records that blocker blocks blocked. It reports whether a new
// link was created (false if it already existed).
func AddBlocker(ctx context.Context, q *db.Queries, blocker, blocked db.GetIssueRow) (bool, error) {
	if blocker.ID == blocked.ID {
		return false, invalidError("an issue can't block itself")
	}
	if blocker.TeamID != blocked.TeamID {
		return false, invalidError("blocking issues must be in the same team")
	}
	cycle, err := q.DoesIssueBlockTransitively(ctx, db.DoesIssueBlockTransitivelyParams{FromID: blocked.ID, ToID: blocker.ID})
	if err != nil {
		return false, err
	}
	if cycle {
		return false, invalidError("that would create a circular dependency")
	}
	n, err := q.AddIssueBlocker(ctx, db.AddIssueBlockerParams{BlockerID: blocker.ID, BlockedID: blocked.ID})
	return n > 0, err
}

// RemoveBlocker deletes a blocking link. It reports whether one existed.
func RemoveBlocker(ctx context.Context, q *db.Queries, blocker, blocked pgtype.UUID) (bool, error) {
	n, err := q.RemoveIssueBlocker(ctx, db.RemoveIssueBlockerParams{BlockerID: blocker, BlockedID: blocked})
	return n > 0, err
}
