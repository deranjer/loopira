// Package dto holds the response shapes shared between the REST API
// (internal/api) and the MCP tool surface (internal/mcpserver), plus the
// sqlc-row -> DTO conversions that build them. This lives in its own
// package rather than being exported straight from internal/api because
// internal/api mounts internal/mcpserver's HTTP handler, and
// internal/mcpserver needs these same shapes to answer tool calls — an
// api <-> mcpserver import in either direction would cycle.
package dto

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/deranjer/loopira/internal/db"
)

const TimeFormat = "2006-01-02T15:04:05Z07:00"

func uid(u pgtype.UUID) string {
	return u.String()
}

func nullableUID(u pgtype.UUID) *string {
	if !u.Valid {
		return nil
	}
	s := u.String()
	return &s
}

func nullableText(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func ts(t pgtype.Timestamptz) time.Time {
	return t.Time
}

func nullableDate(d pgtype.Date) *string {
	if !d.Valid {
		return nil
	}
	s := d.Time.Format("2006-01-02")
	return &s
}

func progressPct(done, total int32) int {
	if total == 0 {
		return 0
	}
	return int(done * 100 / total)
}

type IssueLabel struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type Issue struct {
	ID           string      `json:"id"`
	Identifier   string      `json:"identifier"` // e.g. ENG-142
	Title        string      `json:"title"`
	Description  string      `json:"description"`
	Status       string      `json:"status"`
	Priority     int16       `json:"priority"`
	AssigneeID   *string     `json:"assigneeId"`
	AssigneeName *string     `json:"assigneeName"`
	ProjectID    *string     `json:"projectId"`
	ProjectName  *string     `json:"projectName"`
	CycleID      *string     `json:"cycleId"`
	Label        *IssueLabel `json:"label"`
	CreatedAt    string      `json:"createdAt"`
	UpdatedAt    string      `json:"updatedAt"`
}

func IssueFromListRow(r db.ListIssuesRow) Issue {
	i := Issue{
		ID:           uid(r.ID),
		Identifier:   fmt.Sprintf("%s-%d", r.TeamKey, r.Number),
		Title:        r.Title,
		Description:  r.Description,
		Status:       r.Status,
		Priority:     r.Priority,
		AssigneeID:   nullableUID(r.AssigneeID),
		AssigneeName: nullableText(r.AssigneeName),
		ProjectID:    nullableUID(r.ProjectID),
		ProjectName:  nullableText(r.ProjectName),
		CycleID:      nullableUID(r.CycleID),
		CreatedAt:    ts(r.CreatedAt).Format(TimeFormat),
		UpdatedAt:    ts(r.UpdatedAt).Format(TimeFormat),
	}
	if r.LabelName != "" {
		i.Label = &IssueLabel{ID: uid(r.LabelID), Name: r.LabelName, Color: r.LabelColor}
	}
	return i
}

func IssueFromGetRow(r db.GetIssueRow) Issue {
	i := Issue{
		ID:           uid(r.ID),
		Identifier:   fmt.Sprintf("%s-%d", r.TeamKey, r.Number),
		Title:        r.Title,
		Description:  r.Description,
		Status:       r.Status,
		Priority:     r.Priority,
		AssigneeID:   nullableUID(r.AssigneeID),
		AssigneeName: nullableText(r.AssigneeName),
		ProjectID:    nullableUID(r.ProjectID),
		ProjectName:  nullableText(r.ProjectName),
		CycleID:      nullableUID(r.CycleID),
		CreatedAt:    ts(r.CreatedAt).Format(TimeFormat),
		UpdatedAt:    ts(r.UpdatedAt).Format(TimeFormat),
	}
	if r.LabelName != "" {
		i.Label = &IssueLabel{ID: uid(r.LabelID), Name: r.LabelName, Color: r.LabelColor}
	}
	return i
}

type Project struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
	Status      string  `json:"status"`
	Progress    int     `json:"progress"` // 0-100, percent of issues done
}

func ProjectFromRow(p db.ListProjectsRow) Project {
	return Project{
		ID:          uid(p.ID),
		Name:        p.Name,
		Description: nullableText(p.Description),
		Status:      p.Status,
		Progress:    progressPct(p.DoneCount, p.IssueCount),
	}
}

// ProjectFromNew builds a Project from a freshly-created row (no issues
// attached yet, so no join/progress data — matches db.Project, the plain
// table row CreateProject returns).
func ProjectFromNew(p db.Project) Project {
	return Project{
		ID:          uid(p.ID),
		Name:        p.Name,
		Description: nullableText(p.Description),
		Status:      p.Status,
		Progress:    0,
	}
}

type Cycle struct {
	ID        string  `json:"id"`
	Number    int32   `json:"number"`
	StartDate *string `json:"startDate"`
	EndDate   *string `json:"endDate"`
	Active    bool    `json:"active"`
	Done      int32   `json:"done"`
	Total     int32   `json:"total"`
	Progress  int     `json:"progress"`
}

func CycleFromRow(c db.ListCyclesRow) Cycle {
	return Cycle{
		ID:        uid(c.ID),
		Number:    c.Number,
		StartDate: nullableDate(c.StartDate),
		EndDate:   nullableDate(c.EndDate),
		Active:    c.Active,
		Done:      c.DoneCount,
		Total:     c.IssueCount,
		Progress:  progressPct(c.DoneCount, c.IssueCount),
	}
}

// CycleFromNew builds a Cycle from a freshly-created row (no issues
// attached yet — matches db.Cycle, the plain table row CreateCycle
// returns).
func CycleFromNew(c db.Cycle) Cycle {
	today := time.Now()
	return Cycle{
		ID:        uid(c.ID),
		Number:    c.Number,
		StartDate: nullableDate(c.StartDate),
		EndDate:   nullableDate(c.EndDate),
		Active:    !c.StartDate.Time.After(today) && !c.EndDate.Time.Before(today),
		Done:      0,
		Total:     0,
		Progress:  0,
	}
}

type Label struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

func LabelFromRow(l db.Label) Label {
	return Label{ID: uid(l.ID), Name: l.Name, Color: l.Color}
}

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

func UserFromRow(u db.User) User {
	return User{ID: uid(u.ID), Name: u.Name, Email: u.Email, Role: u.Role}
}
