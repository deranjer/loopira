// Package dto holds the response shapes shared between the REST API
// (internal/api) and the MCP tool surface (internal/mcpserver), plus the
// sqlc-row -> DTO conversions that build them. This lives in its own
// package rather than being exported straight from internal/api because
// internal/api mounts internal/mcpserver's HTTP handler, and
// internal/mcpserver needs these same shapes to answer tool calls — an
// api <-> mcpserver import in either direction would cycle.
package dto

import (
	"encoding/json"
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

func nullableInt(i pgtype.Int4) *int {
	if !i.Valid {
		return nil
	}
	v := int(i.Int32)
	return &v
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
	ParentID     *string     `json:"parentId"`
	Blocked      bool        `json:"blocked"` // has an unresolved blocking issue
	Label        *IssueLabel `json:"label"`
	CreatedAt    string      `json:"createdAt"`
	UpdatedAt    string      `json:"updatedAt"`
}

type IssueHistoryEntry struct {
	ID        string          `json:"id"`
	IssueID   string          `json:"issueId"`
	ActorID   *string         `json:"actorId"`
	ActorName string          `json:"actorName"`
	Action    string          `json:"action"`
	Changes   json.RawMessage `json:"changes"`
	CreatedAt string          `json:"createdAt"`
}

func IssueHistoryFromRow(r db.ListIssueHistoryRow) IssueHistoryEntry {
	return IssueHistoryEntry{
		ID:        uid(r.ID),
		IssueID:   uid(r.IssueID),
		ActorID:   nullableUID(r.ActorID),
		ActorName: r.ActorName,
		Action:    r.Action,
		Changes:   json.RawMessage(r.Changes),
		CreatedAt: ts(r.CreatedAt).Format(TimeFormat),
	}
}

// IssueRef is a compact pointer to another issue, used for hierarchy and
// dependency links so callers don't pay for full issue bodies.
type IssueRef struct {
	ID         string `json:"id"`
	Identifier string `json:"identifier"`
	Title      string `json:"title"`
	Status     string `json:"status"`
}

type IssueLinks struct {
	Parent    *IssueRef  `json:"parent"`
	Children  []IssueRef `json:"children"`
	BlockedBy []IssueRef `json:"blockedBy"`
	Blocks    []IssueRef `json:"blocks"`
}

func IssueLinksFromRows(rows []db.ListIssueLinksRow) IssueLinks {
	out := IssueLinks{Children: []IssueRef{}, BlockedBy: []IssueRef{}, Blocks: []IssueRef{}}
	for _, r := range rows {
		ref := IssueRef{ID: uid(r.ID), Identifier: fmt.Sprintf("%s-%d", r.TeamKey, r.Number), Title: r.Title, Status: r.Status}
		switch r.Kind {
		case "parent":
			out.Parent = &ref
		case "child":
			out.Children = append(out.Children, ref)
		case "blocked_by":
			out.BlockedBy = append(out.BlockedBy, ref)
		case "blocks":
			out.Blocks = append(out.Blocks, ref)
		}
	}
	return out
}

type Comment struct {
	ID         string `json:"id"`
	IssueID    string `json:"issueId"`
	AuthorID   string `json:"authorId"`
	AuthorName string `json:"authorName"`
	Body       string `json:"body"`
	CreatedAt  string `json:"createdAt"`
}

func CommentFromListRow(c db.ListIssueCommentsRow) Comment {
	return Comment{
		ID:         uid(c.ID),
		IssueID:    uid(c.IssueID),
		AuthorID:   uid(c.AuthorID),
		AuthorName: c.AuthorName,
		Body:       c.Body,
		CreatedAt:  ts(c.CreatedAt).Format(TimeFormat),
	}
}

func CommentFromGetRow(c db.GetCommentRow) Comment {
	return Comment{
		ID:         uid(c.ID),
		IssueID:    uid(c.IssueID),
		AuthorID:   uid(c.AuthorID),
		AuthorName: c.AuthorName,
		Body:       c.Body,
		CreatedAt:  ts(c.CreatedAt).Format(TimeFormat),
	}
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
		ParentID:     nullableUID(r.ParentID),
		Blocked:      r.Blocked,
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
		ParentID:     nullableUID(r.ParentID),
		Blocked:      r.Blocked,
		CreatedAt:    ts(r.CreatedAt).Format(TimeFormat),
		UpdatedAt:    ts(r.UpdatedAt).Format(TimeFormat),
	}
	if r.LabelName != "" {
		i.Label = &IssueLabel{ID: uid(r.LabelID), Name: r.LabelName, Color: r.LabelColor}
	}
	return i
}

type Project struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Description  *string `json:"description"`
	Status       string  `json:"status"`
	Priority     int16   `json:"priority"`
	LeadID       *string `json:"leadId"`
	LeadName     *string `json:"leadName"`
	TargetDate   *string `json:"targetDate"`
	TemplateID   *string `json:"templateId"`
	TemplateName *string `json:"templateName"`
	IssueCount   int     `json:"issueCount"`
	Progress     int     `json:"progress"` // 0-100, percent of issues done
}

func ProjectFromRow(p db.ListProjectsRow) Project {
	return Project{
		ID:           uid(p.ID),
		Name:         p.Name,
		Description:  nullableText(p.Description),
		Status:       p.Status,
		Priority:     p.Priority,
		LeadID:       nullableUID(p.LeadID),
		LeadName:     nullableText(p.LeadName),
		TargetDate:   nullableDate(p.TargetDate),
		TemplateID:   nullableUID(p.TemplateID),
		TemplateName: nullableText(p.TemplateName),
		IssueCount:   int(p.IssueCount),
		Progress:     progressPct(p.DoneCount, p.IssueCount),
	}
}

func ProjectFromGetRow(p db.GetProjectRow) Project {
	return Project{
		ID:           uid(p.ID),
		Name:         p.Name,
		Description:  nullableText(p.Description),
		Status:       p.Status,
		Priority:     p.Priority,
		LeadID:       nullableUID(p.LeadID),
		LeadName:     nullableText(p.LeadName),
		TargetDate:   nullableDate(p.TargetDate),
		TemplateID:   nullableUID(p.TemplateID),
		TemplateName: nullableText(p.TemplateName),
		IssueCount:   int(p.IssueCount),
		Progress:     progressPct(p.DoneCount, p.IssueCount),
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

type Attachment struct {
	ID             string `json:"id"`
	Filename       string `json:"filename"`
	ContentType    string `json:"contentType"`
	SizeBytes      int64  `json:"sizeBytes"`
	UploadedBy     string `json:"uploadedBy"`
	UploadedByName string `json:"uploadedByName"`
	CreatedAt      string `json:"createdAt"`
}

type View struct {
	ID         string          `json:"id"`
	OwnerID    string          `json:"ownerId"`
	Name       string          `json:"name"`
	Definition json.RawMessage `json:"definition"`
	Shared     bool            `json:"shared"`
	CreatedAt  string          `json:"createdAt"`
}

func ViewFromRow(v db.View) View {
	return View{
		ID:         uid(v.ID),
		OwnerID:    uid(v.OwnerID),
		Name:       v.Name,
		Definition: json.RawMessage(v.Definition),
		Shared:     v.Shared,
		CreatedAt:  ts(v.CreatedAt).Format(TimeFormat),
	}
}

func AttachmentFromRow(a db.ListProjectAttachmentsRow) Attachment {
	return Attachment{
		ID:             uid(a.ID),
		Filename:       a.Filename,
		ContentType:    a.ContentType,
		SizeBytes:      a.SizeBytes,
		UploadedBy:     uid(a.UploadedBy),
		UploadedByName: a.UploadedByName,
		CreatedAt:      ts(a.CreatedAt).Format(TimeFormat),
	}
}

type WorkLog struct {
	ID          string `json:"id"`
	ProjectID   string `json:"projectId"`
	ProjectName string `json:"projectName"`
	AuthorID    string `json:"authorId"`
	AuthorName  string `json:"authorName"`
	Source      string `json:"source"` // "human" or "agent"
	Title       string `json:"title"`
	Body        string `json:"body"`
	CreatedAt   string `json:"createdAt"`
}

func WorkLogFromProjectRow(w db.ListProjectWorkLogsRow) WorkLog {
	return WorkLog{
		ID:          uid(w.ID),
		ProjectID:   uid(w.ProjectID),
		ProjectName: w.ProjectName,
		AuthorID:    uid(w.AuthorID),
		AuthorName:  w.AuthorName,
		Source:      w.Source,
		Title:       w.Title,
		Body:        w.Body,
		CreatedAt:   ts(w.CreatedAt).Format(TimeFormat),
	}
}

func WorkLogFromGlobalRow(w db.ListWorkLogsRow) WorkLog {
	return WorkLog{
		ID:          uid(w.ID),
		ProjectID:   uid(w.ProjectID),
		ProjectName: w.ProjectName,
		AuthorID:    uid(w.AuthorID),
		AuthorName:  w.AuthorName,
		Source:      w.Source,
		Title:       w.Title,
		Body:        w.Body,
		CreatedAt:   ts(w.CreatedAt).Format(TimeFormat),
	}
}

func WorkLogFromGetRow(w db.GetWorkLogRow) WorkLog {
	return WorkLog{
		ID:          uid(w.ID),
		ProjectID:   uid(w.ProjectID),
		ProjectName: w.ProjectName,
		AuthorID:    uid(w.AuthorID),
		AuthorName:  w.AuthorName,
		Source:      w.Source,
		Title:       w.Title,
		Body:        w.Body,
		CreatedAt:   ts(w.CreatedAt).Format(TimeFormat),
	}
}

type TemplateFragment struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Category   *string `json:"category"`
	Content    string  `json:"content"`
	Version    int     `json:"version"`
	AuthorID   *string `json:"authorId"`
	AuthorName *string `json:"authorName"`
	CreatedAt  string  `json:"createdAt"`
	UpdatedAt  string  `json:"updatedAt"`
}

func TemplateFragmentFromRow(f db.ListTemplateFragmentsRow) TemplateFragment {
	return TemplateFragment{
		ID:         uid(f.ID),
		Name:       f.Name,
		Category:   nullableText(f.Category),
		Content:    f.Content,
		Version:    int(f.Version),
		AuthorID:   nullableUID(f.CreatedBy),
		AuthorName: nullableText(f.AuthorName),
		CreatedAt:  ts(f.CreatedAt).Format(TimeFormat),
		UpdatedAt:  ts(f.UpdatedAt).Format(TimeFormat),
	}
}

func TemplateFragmentFromGetRow(f db.GetTemplateFragmentRow) TemplateFragment {
	return TemplateFragment{
		ID:         uid(f.ID),
		Name:       f.Name,
		Category:   nullableText(f.Category),
		Content:    f.Content,
		Version:    int(f.Version),
		AuthorID:   nullableUID(f.CreatedBy),
		AuthorName: nullableText(f.AuthorName),
		CreatedAt:  ts(f.CreatedAt).Format(TimeFormat),
		UpdatedAt:  ts(f.UpdatedAt).Format(TimeFormat),
	}
}

func TemplateFragmentFromRecord(f db.TemplateFragment) TemplateFragment {
	return TemplateFragment{
		ID:        uid(f.ID),
		Name:      f.Name,
		Category:  nullableText(f.Category),
		Content:   f.Content,
		Version:   int(f.Version),
		AuthorID:  nullableUID(f.CreatedBy),
		CreatedAt: ts(f.CreatedAt).Format(TimeFormat),
		UpdatedAt: ts(f.UpdatedAt).Format(TimeFormat),
	}
}

type FragmentUsage struct {
	ProjectGuideFragmentID string `json:"projectGuideFragmentId"`
	ProjectID              string `json:"projectId"`
	ProjectName            string `json:"projectName"`
	LocallyModified        bool   `json:"locallyModified"`
	BaseVersion            *int   `json:"baseVersion"`
}

func FragmentUsageFromRow(r db.ListFragmentUsageRow) FragmentUsage {
	return FragmentUsage{
		ProjectGuideFragmentID: uid(r.ProjectGuideFragmentID),
		ProjectID:              uid(r.ProjectID),
		ProjectName:            r.ProjectName,
		LocallyModified:        r.LocallyModified,
		BaseVersion:            nullableInt(r.BaseVersion),
	}
}

type TemplateFragmentRef struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Category *string `json:"category"`
	Position int     `json:"position"`
}

type Template struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description *string               `json:"description"`
	AuthorID    *string               `json:"authorId"`
	AuthorName  *string               `json:"authorName"`
	CreatedAt   string                `json:"createdAt"`
	UpdatedAt   string                `json:"updatedAt"`
	Fragments   []TemplateFragmentRef `json:"fragments"`
}

func TemplateFromRow(t db.ListTemplatesRow) Template {
	return Template{
		ID:          uid(t.ID),
		Name:        t.Name,
		Description: nullableText(t.Description),
		AuthorID:    nullableUID(t.CreatedBy),
		AuthorName:  nullableText(t.AuthorName),
		CreatedAt:   ts(t.CreatedAt).Format(TimeFormat),
		UpdatedAt:   ts(t.UpdatedAt).Format(TimeFormat),
	}
}

func TemplateFromRecord(t db.Template) Template {
	return Template{
		ID:          uid(t.ID),
		Name:        t.Name,
		Description: nullableText(t.Description),
		AuthorID:    nullableUID(t.CreatedBy),
		CreatedAt:   ts(t.CreatedAt).Format(TimeFormat),
		UpdatedAt:   ts(t.UpdatedAt).Format(TimeFormat),
		Fragments:   []TemplateFragmentRef{},
	}
}

func TemplateFromGetRow(t db.GetTemplateRow) Template {
	return Template{
		ID:          uid(t.ID),
		Name:        t.Name,
		Description: nullableText(t.Description),
		AuthorID:    nullableUID(t.CreatedBy),
		AuthorName:  nullableText(t.AuthorName),
		CreatedAt:   ts(t.CreatedAt).Format(TimeFormat),
		UpdatedAt:   ts(t.UpdatedAt).Format(TimeFormat),
	}
}

func TemplateFragmentRefFromRow(l db.ListTemplateLinksRow) TemplateFragmentRef {
	return TemplateFragmentRef{
		ID:       uid(l.FragmentID),
		Name:     l.Name,
		Category: nullableText(l.Category),
		Position: int(l.Position),
	}
}

type ProjectGuideFragment struct {
	ID              string  `json:"id"`
	ProjectID       string  `json:"projectId"`
	FragmentID      *string `json:"fragmentId"`
	Name            string  `json:"name"`
	Content         string  `json:"content"`
	BaseVersion     *int    `json:"baseVersion"`
	LocallyModified bool    `json:"locallyModified"`
	Position        int     `json:"position"`
	CreatedAt       string  `json:"createdAt"`
	UpdatedAt       string  `json:"updatedAt"`
}

func ProjectGuideFragmentFromRow(f db.ProjectGuideFragment) ProjectGuideFragment {
	return ProjectGuideFragment{
		ID:              uid(f.ID),
		ProjectID:       uid(f.ProjectID),
		FragmentID:      nullableUID(f.FragmentID),
		Name:            f.Name,
		Content:         f.Content,
		BaseVersion:     nullableInt(f.BaseVersion),
		LocallyModified: f.LocallyModified,
		Position:        int(f.Position),
		CreatedAt:       ts(f.CreatedAt).Format(TimeFormat),
		UpdatedAt:       ts(f.UpdatedAt).Format(TimeFormat),
	}
}

// IssueSummary is the compact list form of an issue for agent clients:
// no description, timestamps, or ids beyond the human identifier (which
// every issue-taking tool accepts).
type IssueSummary struct {
	Identifier string `json:"identifier"`
	Title      string `json:"title"`
	Status     string `json:"status"`
	Priority   int16  `json:"priority"`
	Assignee   string `json:"assignee,omitempty"`
	Project    string `json:"project,omitempty"`
	Label      string `json:"label,omitempty"`
	Blocked    bool   `json:"blocked,omitempty"`
}

func IssueSummaryFromListRow(r db.ListIssuesRow) IssueSummary {
	s := IssueSummary{
		Identifier: fmt.Sprintf("%s-%d", r.TeamKey, r.Number),
		Title:      r.Title,
		Status:     r.Status,
		Priority:   r.Priority,
		Label:      r.LabelName,
		Blocked:    r.Blocked,
	}
	if r.AssigneeName.Valid {
		s.Assignee = r.AssigneeName.String
	}
	if r.ProjectName.Valid {
		s.Project = r.ProjectName.String
	}
	return s
}

// WorkLogPreviewChars caps the body excerpt in WorkLogSummary.
const WorkLogPreviewChars = 240

// WorkLogSummary is the compact list form of a work log entry.
type WorkLogSummary struct {
	ID        string `json:"id"`
	Project   string `json:"project"`
	Author    string `json:"author"`
	Source    string `json:"source"`
	Title     string `json:"title"`
	Preview   string `json:"preview,omitempty"`
	Truncated bool   `json:"truncated,omitempty"`
	CreatedAt string `json:"createdAt"`
}

func WorkLogSummaryFromGlobalRow(w db.ListWorkLogsRow) WorkLogSummary {
	preview, truncated := w.Body, false
	if r := []rune(preview); len(r) > WorkLogPreviewChars {
		preview, truncated = string(r[:WorkLogPreviewChars])+"…", true
	}
	return WorkLogSummary{
		ID:        uid(w.ID),
		Project:   w.ProjectName,
		Author:    w.AuthorName,
		Source:    w.Source,
		Title:     w.Title,
		Preview:   preview,
		Truncated: truncated,
		CreatedAt: ts(w.CreatedAt).Format(TimeFormat),
	}
}

// TemplateFragmentSummary is the catalog form of a fragment: enough to
// pick one, without its (often multi-KB) content.
type TemplateFragmentSummary struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Category *string `json:"category"`
	Version  int     `json:"version"`
}

func TemplateFragmentSummaryFromRow(f db.ListTemplateFragmentsRow) TemplateFragmentSummary {
	return TemplateFragmentSummary{ID: uid(f.ID), Name: f.Name, Category: nullableText(f.Category), Version: int(f.Version)}
}
