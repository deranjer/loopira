package dto

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/deranjer/loopira/internal/db"
)

func mustUUID(t *testing.T, s string) pgtype.UUID {
	t.Helper()
	var u pgtype.UUID
	if err := u.Scan(s); err != nil {
		t.Fatalf("invalid test uuid %q: %v", s, err)
	}
	return u
}

func TestIssueFromGetRowWithoutLabel(t *testing.T) {
	now := time.Date(2026, 8, 19, 12, 0, 0, 0, time.UTC)
	row := db.GetIssueRow{
		ID:          mustUUID(t, "11111111-1111-1111-1111-111111111111"),
		TeamKey:     "ENG",
		Number:      42,
		Title:       "Fix the thing",
		Description: "It's broken",
		Status:      "todo",
		Priority:    2,
		CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
		// AssigneeID, ProjectID, CycleID, AssigneeName, ProjectName, LabelID
		// all left zero-value (Valid: false) — the "nothing set" case.
		LabelName:  "",
		LabelColor: "",
	}

	issue := IssueFromGetRow(row)

	if issue.Identifier != "ENG-42" {
		t.Errorf("Identifier = %q, want %q", issue.Identifier, "ENG-42")
	}
	if issue.AssigneeID != nil {
		t.Errorf("AssigneeID = %v, want nil", *issue.AssigneeID)
	}
	if issue.ProjectID != nil {
		t.Errorf("ProjectID = %v, want nil", *issue.ProjectID)
	}
	if issue.CycleID != nil {
		t.Errorf("CycleID = %v, want nil", *issue.CycleID)
	}
	if issue.Label != nil {
		t.Errorf("Label = %+v, want nil", *issue.Label)
	}
	if issue.CreatedAt != now.Format(TimeFormat) {
		t.Errorf("CreatedAt = %q, want %q", issue.CreatedAt, now.Format(TimeFormat))
	}
}

func TestIssueFromGetRowWithLabel(t *testing.T) {
	row := db.GetIssueRow{
		ID:         mustUUID(t, "11111111-1111-1111-1111-111111111111"),
		TeamKey:    "ENG",
		Number:     7,
		Status:     "backlog",
		LabelID:    mustUUID(t, "22222222-2222-2222-2222-222222222222"),
		LabelName:  "Bug",
		LabelColor: "#eb5757",
	}

	issue := IssueFromGetRow(row)

	if issue.Label == nil {
		t.Fatal("Label = nil, want non-nil")
	}
	if issue.Label.ID != "22222222-2222-2222-2222-222222222222" {
		t.Errorf("Label.ID = %q, want the label's uuid", issue.Label.ID)
	}
	if issue.Label.Name != "Bug" || issue.Label.Color != "#eb5757" {
		t.Errorf("Label = %+v, want {Bug #eb5757}", *issue.Label)
	}
}

func TestIssueFromGetRowNullableAssigneeAndProject(t *testing.T) {
	row := db.GetIssueRow{
		ID:           mustUUID(t, "11111111-1111-1111-1111-111111111111"),
		TeamKey:      "ENG",
		Number:       1,
		AssigneeID:   mustUUID(t, "33333333-3333-3333-3333-333333333333"),
		AssigneeName: pgtype.Text{String: "Jamie Doe", Valid: true},
		ProjectID:    mustUUID(t, "44444444-4444-4444-4444-444444444444"),
		ProjectName:  pgtype.Text{String: "Mobile App", Valid: true},
	}

	issue := IssueFromGetRow(row)

	if issue.AssigneeID == nil || *issue.AssigneeID != "33333333-3333-3333-3333-333333333333" {
		t.Errorf("AssigneeID = %v, want the assignee's uuid", issue.AssigneeID)
	}
	if issue.AssigneeName == nil || *issue.AssigneeName != "Jamie Doe" {
		t.Errorf("AssigneeName = %v, want Jamie Doe", issue.AssigneeName)
	}
	if issue.ProjectID == nil || *issue.ProjectID != "44444444-4444-4444-4444-444444444444" {
		t.Errorf("ProjectID = %v, want the project's uuid", issue.ProjectID)
	}
	if issue.ProjectName == nil || *issue.ProjectName != "Mobile App" {
		t.Errorf("ProjectName = %v, want Mobile App", issue.ProjectName)
	}
}

func TestProgressPct(t *testing.T) {
	tests := []struct {
		done, total int32
		want        int
	}{
		{done: 0, total: 0, want: 0}, // must not divide by zero
		{done: 0, total: 5, want: 0},
		{done: 5, total: 5, want: 100},
		{done: 1, total: 4, want: 25},
		{done: 1, total: 3, want: 33}, // truncates, doesn't round
	}
	for _, tt := range tests {
		if got := progressPct(tt.done, tt.total); got != tt.want {
			t.Errorf("progressPct(%d, %d) = %d, want %d", tt.done, tt.total, got, tt.want)
		}
	}
}
