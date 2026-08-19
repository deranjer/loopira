package mcpserver

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/deranjer/loopira/internal/auth"
	"github.com/deranjer/loopira/internal/db"
	"github.com/deranjer/loopira/internal/dto"
	"github.com/deranjer/loopira/internal/ws"
)

func parseUUID(s string) (pgtype.UUID, error) {
	var u pgtype.UUID
	err := u.Scan(s)
	return u, err
}

// errorResult reports a tool-level failure. Per the MCP spec this belongs
// in the result content with IsError set, not as an RPC-level error, so
// the calling agent can see what went wrong and try something else.
func errorResult(format string, args ...any) *mcp.CallToolResult {
	return &mcp.CallToolResult{
		IsError: true,
		Content: []mcp.Content{&mcp.TextContent{Text: fmt.Sprintf(format, args...)}},
	}
}

var errNoWrite = errorResult("this API key is read-only and can't make changes")

// currentTeam resolves "the" workspace team — v1 is single-workspace, so
// every tool operates on the first (only) seeded team rather than taking
// a teamId argument.
func (s *toolServer) currentTeam(ctx context.Context) (db.Team, error) {
	teams, err := s.q.ListTeams(ctx)
	if err != nil {
		return db.Team{}, err
	}
	if len(teams) == 0 {
		return db.Team{}, fmt.Errorf("no team configured")
	}
	return teams[0], nil
}

// resolveIssue accepts either a raw uuid or a human identifier like
// "ENG-3", matching how someone would actually refer to an issue in chat.
func (s *toolServer) resolveIssue(ctx context.Context, teamID pgtype.UUID, ref string) (db.GetIssueRow, error) {
	if id, err := parseUUID(ref); err == nil {
		return s.q.GetIssue(ctx, id)
	}
	if _, numStr, ok := strings.Cut(ref, "-"); ok {
		if n, err := strconv.Atoi(numStr); err == nil {
			row, err := s.q.GetIssueByNumber(ctx, db.GetIssueByNumberParams{TeamID: teamID, Number: int32(n)})
			// GetIssueByNumberRow and GetIssueRow are the same joined shape
			// under different sqlc-generated names — safe to convert directly.
			return db.GetIssueRow(row), err
		}
	}
	return db.GetIssueRow{}, fmt.Errorf("could not resolve issue %q", ref)
}

type listIssuesArgs struct {
	Status    string `json:"status,omitempty" jsonschema:"filter by status: backlog, todo, in_progress, done, or canceled"`
	ProjectID string `json:"projectId,omitempty" jsonschema:"filter by project id, from list_projects"`
	CycleID   string `json:"cycleId,omitempty" jsonschema:"filter by cycle id, from list_cycles"`
}

func (s *toolServer) listIssues(ctx context.Context, _ *mcp.CallToolRequest, args listIssuesArgs) (*mcp.CallToolResult, []dto.Issue, error) {
	team, err := s.currentTeam(ctx)
	if err != nil {
		return errorResult("%s", err), nil, nil
	}
	params := db.ListIssuesParams{TeamID: team.ID}
	if args.Status != "" {
		params.Status = pgtype.Text{String: args.Status, Valid: true}
	}
	if args.ProjectID != "" {
		id, err := parseUUID(args.ProjectID)
		if err != nil {
			return errorResult("invalid projectId %q", args.ProjectID), nil, nil
		}
		params.ProjectID = id
	}
	if args.CycleID != "" {
		id, err := parseUUID(args.CycleID)
		if err != nil {
			return errorResult("invalid cycleId %q", args.CycleID), nil, nil
		}
		params.CycleID = id
	}
	rows, err := s.q.ListIssues(ctx, params)
	if err != nil {
		return nil, nil, err
	}
	out := make([]dto.Issue, len(rows))
	for i, r := range rows {
		out[i] = dto.IssueFromListRow(r)
	}
	return nil, out, nil
}

type getIssueArgs struct {
	ID string `json:"id" jsonschema:"issue id (uuid) or identifier like ENG-3"`
}

func (s *toolServer) getIssue(ctx context.Context, _ *mcp.CallToolRequest, args getIssueArgs) (*mcp.CallToolResult, dto.Issue, error) {
	team, err := s.currentTeam(ctx)
	if err != nil {
		return errorResult("%s", err), dto.Issue{}, nil
	}
	row, err := s.resolveIssue(ctx, team.ID, args.ID)
	if err != nil {
		return errorResult("issue %q not found", args.ID), dto.Issue{}, nil
	}
	return nil, dto.IssueFromGetRow(row), nil
}

// applyIssueLabel replaces an issue's label set with zero or one label —
// the UI only ever shows a single label per issue even though the schema
// supports many-to-many via issue_labels. Mirrors internal/api's own
// applyIssueLabel; not shared directly since mcpserver can't import
// api's unexported helpers (and importing internal/api at all would
// reintroduce the cycle this package's whole existence avoids).
func (s *toolServer) applyIssueLabel(ctx context.Context, issueID pgtype.UUID, labelID string) error {
	if err := s.q.ClearIssueLabels(ctx, issueID); err != nil {
		return err
	}
	if labelID == "" {
		return nil
	}
	id, err := parseUUID(labelID)
	if err != nil {
		return fmt.Errorf("invalid labelId %q", labelID)
	}
	return s.q.AddIssueLabel(ctx, db.AddIssueLabelParams{IssueID: issueID, LabelID: id})
}

type createIssueArgs struct {
	Title       string `json:"title" jsonschema:"issue title"`
	Description string `json:"description,omitempty" jsonschema:"issue description"`
	Priority    int    `json:"priority,omitempty" jsonschema:"priority level 0 to 4: 0 none (default) 1 urgent 2 high 3 medium 4 low"`
	AssigneeID  string `json:"assigneeId,omitempty" jsonschema:"assignee user id, from list_users"`
	ProjectID   string `json:"projectId,omitempty" jsonschema:"project id, from list_projects"`
	LabelID     string `json:"labelId,omitempty" jsonschema:"label id, from list_labels"`
}

func (s *toolServer) createIssue(ctx context.Context, _ *mcp.CallToolRequest, args createIssueArgs) (*mcp.CallToolResult, dto.Issue, error) {
	if !auth.CanWrite(ctx) {
		return errNoWrite, dto.Issue{}, nil
	}
	if strings.TrimSpace(args.Title) == "" {
		return errorResult("title is required"), dto.Issue{}, nil
	}
	team, err := s.currentTeam(ctx)
	if err != nil {
		return errorResult("%s", err), dto.Issue{}, nil
	}
	userIDStr, _ := auth.UserID(ctx)
	createdBy, err := parseUUID(userIDStr)
	if err != nil {
		return errorResult("could not resolve caller"), dto.Issue{}, nil
	}
	var assigneeID, projectID pgtype.UUID
	if args.AssigneeID != "" {
		if assigneeID, err = parseUUID(args.AssigneeID); err != nil {
			return errorResult("invalid assigneeId %q", args.AssigneeID), dto.Issue{}, nil
		}
	}
	if args.ProjectID != "" {
		if projectID, err = parseUUID(args.ProjectID); err != nil {
			return errorResult("invalid projectId %q", args.ProjectID), dto.Issue{}, nil
		}
	}
	created, err := s.q.CreateIssue(ctx, db.CreateIssueParams{
		TeamID:      team.ID,
		Title:       args.Title,
		Description: args.Description,
		Priority:    int16(args.Priority),
		AssigneeID:  assigneeID,
		ProjectID:   projectID,
		CreatedBy:   createdBy,
	})
	if err != nil {
		return nil, dto.Issue{}, err
	}
	if err := s.applyIssueLabel(ctx, created.ID, args.LabelID); err != nil {
		return errorResult("%s", err), dto.Issue{}, nil
	}
	return s.broadcastAndReturn(ctx, created.ID, team.ID, "issue.created")
}

type updateIssueStatusArgs struct {
	ID     string `json:"id" jsonschema:"issue id or identifier"`
	Status string `json:"status" jsonschema:"backlog, todo, in_progress, done, or canceled"`
}

var validStatuses = map[string]bool{
	"backlog": true, "todo": true, "in_progress": true, "done": true, "canceled": true,
}

func (s *toolServer) updateIssueStatus(ctx context.Context, _ *mcp.CallToolRequest, args updateIssueStatusArgs) (*mcp.CallToolResult, dto.Issue, error) {
	if !auth.CanWrite(ctx) {
		return errNoWrite, dto.Issue{}, nil
	}
	if !validStatuses[args.Status] {
		return errorResult("invalid status %q", args.Status), dto.Issue{}, nil
	}
	team, err := s.currentTeam(ctx)
	if err != nil {
		return errorResult("%s", err), dto.Issue{}, nil
	}
	issue, err := s.resolveIssue(ctx, team.ID, args.ID)
	if err != nil {
		return errorResult("issue %q not found", args.ID), dto.Issue{}, nil
	}
	updated, err := s.q.UpdateIssueStatus(ctx, db.UpdateIssueStatusParams{ID: issue.ID, Status: args.Status})
	if err != nil {
		return nil, dto.Issue{}, err
	}
	return s.broadcastAndReturn(ctx, updated.ID, team.ID, "issue.updated")
}

type updateIssueArgs struct {
	ID          string  `json:"id" jsonschema:"issue id or identifier"`
	Title       *string `json:"title,omitempty" jsonschema:"new title"`
	Description *string `json:"description,omitempty" jsonschema:"new description"`
	Priority    *int    `json:"priority,omitempty" jsonschema:"new priority level 0 to 4: 0 none 1 urgent 2 high 3 medium 4 low"`
	AssigneeID  *string `json:"assigneeId,omitempty" jsonschema:"new assignee user id; empty string clears it"`
	ProjectID   *string `json:"projectId,omitempty" jsonschema:"new project id; empty string clears it"`
	CycleID     *string `json:"cycleId,omitempty" jsonschema:"new cycle id; empty string clears it"`
	LabelID     *string `json:"labelId,omitempty" jsonschema:"new label id from list_labels; empty string clears it; omit to leave unchanged"`
}

func (s *toolServer) updateIssue(ctx context.Context, _ *mcp.CallToolRequest, args updateIssueArgs) (*mcp.CallToolResult, dto.Issue, error) {
	if !auth.CanWrite(ctx) {
		return errNoWrite, dto.Issue{}, nil
	}
	team, err := s.currentTeam(ctx)
	if err != nil {
		return errorResult("%s", err), dto.Issue{}, nil
	}
	current, err := s.resolveIssue(ctx, team.ID, args.ID)
	if err != nil {
		return errorResult("issue %q not found", args.ID), dto.Issue{}, nil
	}

	title, description, priority := current.Title, current.Description, current.Priority
	if args.Title != nil {
		title = *args.Title
	}
	if args.Description != nil {
		description = *args.Description
	}
	if args.Priority != nil {
		priority = int16(*args.Priority)
	}
	assigneeID, err := mergeOptionalRef(current.AssigneeID, args.AssigneeID)
	if err != nil {
		return errorResult("invalid assigneeId %q", *args.AssigneeID), dto.Issue{}, nil
	}
	projectID, err := mergeOptionalRef(current.ProjectID, args.ProjectID)
	if err != nil {
		return errorResult("invalid projectId %q", *args.ProjectID), dto.Issue{}, nil
	}
	cycleID, err := mergeOptionalRef(current.CycleID, args.CycleID)
	if err != nil {
		return errorResult("invalid cycleId %q", *args.CycleID), dto.Issue{}, nil
	}

	updated, err := s.q.UpdateIssueDetails(ctx, db.UpdateIssueDetailsParams{
		ID: current.ID, Title: title, Description: description, Priority: priority,
		AssigneeID: assigneeID, ProjectID: projectID, CycleID: cycleID,
	})
	if err != nil {
		return nil, dto.Issue{}, err
	}
	if args.LabelID != nil {
		if err := s.applyIssueLabel(ctx, updated.ID, *args.LabelID); err != nil {
			return errorResult("%s", err), dto.Issue{}, nil
		}
	}
	return s.broadcastAndReturn(ctx, updated.ID, team.ID, "issue.updated")
}

// mergeOptionalRef applies an optional "new value" (nil = unchanged, "" =
// clear, otherwise a uuid string) over an existing nullable uuid column.
func mergeOptionalRef(current pgtype.UUID, next *string) (pgtype.UUID, error) {
	if next == nil {
		return current, nil
	}
	if *next == "" {
		return pgtype.UUID{}, nil
	}
	return parseUUID(*next)
}

func (s *toolServer) broadcastAndReturn(ctx context.Context, issueID, teamID pgtype.UUID, eventType string) (*mcp.CallToolResult, dto.Issue, error) {
	row, err := s.q.GetIssue(ctx, issueID)
	if err != nil {
		return nil, dto.Issue{}, err
	}
	body := dto.IssueFromGetRow(row)
	s.hub.Broadcast(ws.Event{Type: eventType, TeamID: teamID.String(), Payload: body})
	return nil, body, nil
}

func (s *toolServer) listProjects(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, []dto.Project, error) {
	team, err := s.currentTeam(ctx)
	if err != nil {
		return errorResult("%s", err), nil, nil
	}
	rows, err := s.q.ListProjects(ctx, team.ID)
	if err != nil {
		return nil, nil, err
	}
	out := make([]dto.Project, len(rows))
	for i, r := range rows {
		out[i] = dto.ProjectFromRow(r)
	}
	return nil, out, nil
}

func (s *toolServer) listCycles(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, []dto.Cycle, error) {
	team, err := s.currentTeam(ctx)
	if err != nil {
		return errorResult("%s", err), nil, nil
	}
	rows, err := s.q.ListCycles(ctx, team.ID)
	if err != nil {
		return nil, nil, err
	}
	out := make([]dto.Cycle, len(rows))
	for i, r := range rows {
		out[i] = dto.CycleFromRow(r)
	}
	return nil, out, nil
}

func (s *toolServer) listLabels(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, []dto.Label, error) {
	team, err := s.currentTeam(ctx)
	if err != nil {
		return errorResult("%s", err), nil, nil
	}
	rows, err := s.q.ListLabels(ctx, team.ID)
	if err != nil {
		return nil, nil, err
	}
	out := make([]dto.Label, len(rows))
	for i, r := range rows {
		out[i] = dto.LabelFromRow(r)
	}
	return nil, out, nil
}

func (s *toolServer) listUsers(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, []dto.User, error) {
	rows, err := s.q.ListUsers(ctx)
	if err != nil {
		return nil, nil, err
	}
	out := make([]dto.User, len(rows))
	for i, r := range rows {
		out[i] = dto.UserFromRow(r)
	}
	return nil, out, nil
}
