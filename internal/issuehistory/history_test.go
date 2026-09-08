package issuehistory

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/deranjer/loopira/internal/db"
)

func TestDiffIncludesOnlyChangedEditableFields(t *testing.T) {
	oldAssignee := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	newAssignee := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}
	before := db.GetIssueRow{
		Title:       "Before",
		Description: "Same description",
		Status:      "todo",
		Priority:    3,
		AssigneeID:  oldAssignee,
	}
	after := before
	after.Title = "After"
	after.Status = "done"
	after.AssigneeID = newAssignee

	changes := Diff(before, after)
	if len(changes) != 3 {
		t.Fatalf("expected 3 changes, got %#v", changes)
	}
	if changes["title"].From != "Before" || changes["title"].To != "After" {
		t.Fatalf("unexpected title change: %#v", changes["title"])
	}
	if changes["status"].From != "todo" || changes["status"].To != "done" {
		t.Fatalf("unexpected status change: %#v", changes["status"])
	}
	if _, exists := changes["description"]; exists {
		t.Fatal("unchanged description was included")
	}
}

func TestDiffRepresentsClearedReferencesAsNull(t *testing.T) {
	before := db.GetIssueRow{ProjectID: pgtype.UUID{Bytes: [16]byte{3}, Valid: true}}
	changes := Diff(before, db.GetIssueRow{})

	if changes["project"].From == nil || changes["project"].To != nil {
		t.Fatalf("unexpected project change: %#v", changes["project"])
	}
}
