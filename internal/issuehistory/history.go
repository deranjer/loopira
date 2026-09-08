package issuehistory

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/deranjer/loopira/internal/db"
)

type Change struct {
	From any `json:"from"`
	To   any `json:"to"`
}

func uuidValue(value pgtype.UUID) any {
	if !value.Valid {
		return nil
	}
	return value.String()
}

// Diff returns only user-editable issue fields whose values changed.
func Diff(before, after db.GetIssueRow) map[string]Change {
	changes := make(map[string]Change)
	if before.Title != after.Title {
		changes["title"] = Change{From: before.Title, To: after.Title}
	}
	if before.Description != after.Description {
		changes["description"] = Change{From: before.Description, To: after.Description}
	}
	if before.Status != after.Status {
		changes["status"] = Change{From: before.Status, To: after.Status}
	}
	if before.Priority != after.Priority {
		changes["priority"] = Change{From: before.Priority, To: after.Priority}
	}
	if before.AssigneeID != after.AssigneeID {
		changes["assignee"] = Change{From: uuidValue(before.AssigneeID), To: uuidValue(after.AssigneeID)}
	}
	if before.ProjectID != after.ProjectID {
		changes["project"] = Change{From: uuidValue(before.ProjectID), To: uuidValue(after.ProjectID)}
	}
	if before.CycleID != after.CycleID {
		changes["cycle"] = Change{From: uuidValue(before.CycleID), To: uuidValue(after.CycleID)}
	}
	if before.LabelID != after.LabelID {
		changes["label"] = Change{From: uuidValue(before.LabelID), To: uuidValue(after.LabelID)}
	}
	return changes
}

func Record(ctx context.Context, q *db.Queries, issueID, actorID pgtype.UUID, action string, changes map[string]Change) error {
	if changes == nil {
		changes = map[string]Change{}
	}
	encoded, err := json.Marshal(changes)
	if err != nil {
		return err
	}
	_, err = q.CreateIssueHistory(ctx, db.CreateIssueHistoryParams{
		IssueID: issueID,
		ActorID: actorID,
		Action:  action,
		Changes: encoded,
	})
	return err
}
