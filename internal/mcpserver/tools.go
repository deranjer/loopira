package mcpserver

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

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
	Status     string `json:"status,omitempty" jsonschema:"filter by status: backlog, todo, in_progress, done, or canceled"`
	ProjectID  string `json:"projectId,omitempty" jsonschema:"filter by project id, from list_projects"`
	CycleID    string `json:"cycleId,omitempty" jsonschema:"filter by cycle id, from list_cycles"`
	AssigneeID string `json:"assigneeId,omitempty" jsonschema:"filter by assignee id, from list_users, or 'me' for the current user"`
	Priority   *int16 `json:"priority,omitempty" jsonschema:"filter by priority: 0=none,1=urgent,2=high,3=medium,4=low"`
	LabelID    string `json:"labelId,omitempty" jsonschema:"filter by label id, from list_labels"`
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
	if args.AssigneeID != "" {
		ref := args.AssigneeID
		if ref == "me" {
			ref, _ = auth.UserID(ctx)
		}
		id, err := parseUUID(ref)
		if err != nil {
			return errorResult("invalid assigneeId %q", args.AssigneeID), nil, nil
		}
		params.AssigneeID = id
	}
	if args.LabelID != "" {
		id, err := parseUUID(args.LabelID)
		if err != nil {
			return errorResult("invalid labelId %q", args.LabelID), nil, nil
		}
		params.LabelID = id
	}
	if args.Priority != nil {
		params.Priority = pgtype.Int2{Int16: *args.Priority, Valid: true}
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

type createProjectArgs struct {
	Name        string `json:"name" jsonschema:"project name"`
	Description string `json:"description,omitempty" jsonschema:"project description"`
	Status      string `json:"status,omitempty" jsonschema:"backlog (default), planned, in_progress, paused, completed, or canceled"`
	LeadID      string `json:"leadId,omitempty" jsonschema:"project lead user id, from list_users"`
	Priority    int    `json:"priority,omitempty" jsonschema:"priority level 0 to 4: 0 none (default) 1 urgent 2 high 3 medium 4 low"`
	TargetDate  string `json:"targetDate,omitempty" jsonschema:"target date in YYYY-MM-DD format"`
	TemplateID  string `json:"templateId,omitempty" jsonschema:"template id, from list_templates; its fragments are stamped onto the new project"`
}

var validProjectStatuses = map[string]bool{
	"backlog": true, "planned": true, "in_progress": true, "paused": true, "completed": true, "canceled": true,
}

func projectTargetDate(value string) (pgtype.Date, error) {
	if value == "" {
		return pgtype.Date{}, nil
	}
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return pgtype.Date{}, err
	}
	return pgtype.Date{Time: parsed, Valid: true}, nil
}

// createProject deliberately mirrors the REST API's project-creation flow:
// validate the optional template before inserting, create the project, then
// copy each template fragment into that project's guide. This gives an agent
// the same durable snapshot of the stack guide a UI-created project receives.
func (s *toolServer) createProject(ctx context.Context, _ *mcp.CallToolRequest, args createProjectArgs) (*mcp.CallToolResult, dto.Project, error) {
	if !auth.CanWrite(ctx) {
		return errNoWrite, dto.Project{}, nil
	}
	if strings.TrimSpace(args.Name) == "" {
		return errorResult("name is required"), dto.Project{}, nil
	}
	status := args.Status
	if status == "" {
		status = "backlog"
	}
	if !validProjectStatuses[status] {
		return errorResult("invalid status %q", args.Status), dto.Project{}, nil
	}
	if args.Priority < 0 || args.Priority > 4 {
		return errorResult("priority must be between 0 and 4"), dto.Project{}, nil
	}

	team, err := s.currentTeam(ctx)
	if err != nil {
		return errorResult("%s", err), dto.Project{}, nil
	}
	var leadID, templateID pgtype.UUID
	if args.LeadID != "" {
		if leadID, err = parseUUID(args.LeadID); err != nil {
			return errorResult("invalid leadId %q", args.LeadID), dto.Project{}, nil
		}
	}
	if args.TemplateID != "" {
		if templateID, err = parseUUID(args.TemplateID); err != nil {
			return errorResult("invalid templateId %q", args.TemplateID), dto.Project{}, nil
		}
		if _, err := s.q.GetTemplate(ctx, templateID); err != nil {
			return errorResult("template %q not found", args.TemplateID), dto.Project{}, nil
		}
	}
	targetDate, err := projectTargetDate(args.TargetDate)
	if err != nil {
		return errorResult("invalid targetDate %q; expected YYYY-MM-DD", args.TargetDate), dto.Project{}, nil
	}

	var stampFragments []db.ListTemplateFragmentsForStampRow
	if templateID.Valid {
		stampFragments, err = s.q.ListTemplateFragmentsForStamp(ctx, templateID)
		if err != nil {
			return nil, dto.Project{}, err
		}
	}
	created, err := s.q.CreateProject(ctx, db.CreateProjectParams{
		TeamID:      team.ID,
		Name:        args.Name,
		Description: pgtype.Text{String: args.Description, Valid: args.Description != ""},
		Status:      status,
		LeadID:      leadID,
		Priority:    int16(args.Priority),
		TargetDate:  targetDate,
		TemplateID:  templateID,
	})
	if err != nil {
		return nil, dto.Project{}, err
	}
	if leadID.Valid {
		if err := s.q.AddProjectMember(ctx, db.AddProjectMemberParams{ProjectID: created.ID, UserID: leadID}); err != nil {
			return nil, dto.Project{}, err
		}
	}
	for _, fragment := range stampFragments {
		if _, err := s.q.AddProjectGuideFragment(ctx, db.AddProjectGuideFragmentParams{
			ProjectID:   created.ID,
			FragmentID:  fragment.ID,
			Name:        fragment.Name,
			Content:     fragment.Content,
			BaseVersion: pgtype.Int4{Int32: fragment.Version, Valid: true},
		}); err != nil {
			return nil, dto.Project{}, err
		}
	}
	row, err := s.q.GetProject(ctx, created.ID)
	if err != nil {
		return nil, dto.Project{}, err
	}
	out := dto.ProjectFromGetRow(row)
	s.hub.Broadcast(ws.Event{Type: "project.created", TeamID: team.ID.String(), Payload: out})
	return nil, out, nil
}

type addProjectGuideFragmentArgs struct {
	ProjectID  string `json:"projectId" jsonschema:"project id, from list_projects"`
	FragmentID string `json:"fragmentId,omitempty" jsonschema:"catalog fragment id, from list_template_fragments"`
	Name       string `json:"name,omitempty" jsonschema:"name for a custom guide fragment; required when fragmentId is omitted"`
	Content    string `json:"content,omitempty" jsonschema:"content for a custom guide fragment"`
}

func (s *toolServer) addProjectGuideFragment(ctx context.Context, _ *mcp.CallToolRequest, args addProjectGuideFragmentArgs) (*mcp.CallToolResult, dto.ProjectGuideFragment, error) {
	if !auth.CanWrite(ctx) {
		return errNoWrite, dto.ProjectGuideFragment{}, nil
	}
	projectID, err := parseUUID(args.ProjectID)
	if err != nil {
		return errorResult("invalid projectId %q", args.ProjectID), dto.ProjectGuideFragment{}, nil
	}
	if _, err := s.q.GetProject(ctx, projectID); err != nil {
		return errorResult("project %q not found", args.ProjectID), dto.ProjectGuideFragment{}, nil
	}

	var fragmentID pgtype.UUID
	var baseVersion pgtype.Int4
	name, content := args.Name, args.Content
	if args.FragmentID != "" {
		if fragmentID, err = parseUUID(args.FragmentID); err != nil {
			return errorResult("invalid fragmentId %q", args.FragmentID), dto.ProjectGuideFragment{}, nil
		}
		fragment, err := s.q.GetTemplateFragment(ctx, fragmentID)
		if err != nil {
			return errorResult("fragment %q not found", args.FragmentID), dto.ProjectGuideFragment{}, nil
		}
		name, content = fragment.Name, fragment.Content
		baseVersion = pgtype.Int4{Int32: fragment.Version, Valid: true}
	} else if strings.TrimSpace(name) == "" {
		return errorResult("name is required when fragmentId is omitted"), dto.ProjectGuideFragment{}, nil
	}

	created, err := s.q.AddProjectGuideFragment(ctx, db.AddProjectGuideFragmentParams{
		ProjectID: projectID, FragmentID: fragmentID, Name: name, Content: content, BaseVersion: baseVersion,
	})
	if err != nil {
		return nil, dto.ProjectGuideFragment{}, err
	}
	return nil, dto.ProjectGuideFragmentFromRow(created), nil
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

type addWorkLogArgs struct {
	ProjectID string `json:"projectId" jsonschema:"project id, from list_projects"`
	Title     string `json:"title" jsonschema:"short title for the work log entry"`
	Body      string `json:"body" jsonschema:"markdown body describing what was done and why"`
}

// addWorkLog always tags the entry source as "agent" — this tool only
// exists on the MCP surface, so every call through it is by definition
// agent-initiated, regardless of whether the calling API key also has a
// human owner. Compare the REST create-work-log handler, which derives
// source from auth.ViaAPIKey since a human can call it directly.
func (s *toolServer) addWorkLog(ctx context.Context, _ *mcp.CallToolRequest, args addWorkLogArgs) (*mcp.CallToolResult, dto.WorkLog, error) {
	if !auth.CanWrite(ctx) {
		return errNoWrite, dto.WorkLog{}, nil
	}
	if strings.TrimSpace(args.Title) == "" {
		return errorResult("title is required"), dto.WorkLog{}, nil
	}
	if strings.TrimSpace(args.Body) == "" {
		return errorResult("body is required"), dto.WorkLog{}, nil
	}
	projectID, err := parseUUID(args.ProjectID)
	if err != nil {
		return errorResult("invalid projectId %q", args.ProjectID), dto.WorkLog{}, nil
	}
	project, err := s.q.GetProject(ctx, projectID)
	if err != nil {
		return errorResult("project %q not found", args.ProjectID), dto.WorkLog{}, nil
	}
	userIDStr, _ := auth.UserID(ctx)
	authorID, err := parseUUID(userIDStr)
	if err != nil {
		return errorResult("could not resolve caller"), dto.WorkLog{}, nil
	}
	created, err := s.q.CreateWorkLog(ctx, db.CreateWorkLogParams{
		ProjectID: projectID,
		AuthorID:  authorID,
		Source:    "agent",
		Title:     args.Title,
		Body:      args.Body,
	})
	if err != nil {
		return nil, dto.WorkLog{}, err
	}
	row, err := s.q.GetWorkLog(ctx, created.ID)
	if err != nil {
		return nil, dto.WorkLog{}, err
	}
	body := dto.WorkLogFromGetRow(row)
	s.hub.Broadcast(ws.Event{Type: "worklog.created", TeamID: project.TeamID.String(), Payload: body})
	return nil, body, nil
}

// loadTemplateFragments fetches and attaches a template's ordered fragment
// list. Mirrors internal/api's own loadTemplateFragments; not shared
// directly for the same reason applyIssueLabel isn't — mcpserver can't
// import internal/api's unexported helpers without reintroducing the
// import cycle this package's whole existence avoids.
func (s *toolServer) loadTemplateFragments(ctx context.Context, t *dto.Template, id pgtype.UUID) error {
	links, err := s.q.ListTemplateLinks(ctx, id)
	if err != nil {
		return err
	}
	t.Fragments = make([]dto.TemplateFragmentRef, len(links))
	for i, l := range links {
		t.Fragments[i] = dto.TemplateFragmentRefFromRow(l)
	}
	return nil
}

func (s *toolServer) listTemplates(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, []dto.Template, error) {
	rows, err := s.q.ListTemplates(ctx)
	if err != nil {
		return nil, nil, err
	}
	out := make([]dto.Template, len(rows))
	for i, t := range rows {
		out[i] = dto.TemplateFromRow(t)
	}
	return nil, out, nil
}

type getTemplateArgs struct {
	ID string `json:"id" jsonschema:"template id, from list_templates"`
}

func (s *toolServer) getTemplate(ctx context.Context, _ *mcp.CallToolRequest, args getTemplateArgs) (*mcp.CallToolResult, dto.Template, error) {
	id, err := parseUUID(args.ID)
	if err != nil {
		return errorResult("invalid id %q", args.ID), dto.Template{}, nil
	}
	row, err := s.q.GetTemplate(ctx, id)
	if err != nil {
		return errorResult("template %q not found", args.ID), dto.Template{}, nil
	}
	out := dto.TemplateFromGetRow(row)
	if err := s.loadTemplateFragments(ctx, &out, id); err != nil {
		return nil, dto.Template{}, err
	}
	return nil, out, nil
}

func (s *toolServer) listTemplateFragments(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, []dto.TemplateFragment, error) {
	rows, err := s.q.ListTemplateFragments(ctx)
	if err != nil {
		return nil, nil, err
	}
	out := make([]dto.TemplateFragment, len(rows))
	for i, f := range rows {
		out[i] = dto.TemplateFragmentFromRow(f)
	}
	return nil, out, nil
}

type projectGuideResult struct {
	// TemplateID/TemplateName are also on Project, repeated here since
	// they're the field an agent asking "what template is this project
	// on" cares about most.
	TemplateID   *string                    `json:"templateId"`
	TemplateName *string                    `json:"templateName"`
	Fragments    []dto.ProjectGuideFragment `json:"fragments"`
}

type getProjectGuideArgs struct {
	ProjectID string `json:"projectId" jsonschema:"project id, from list_projects"`
}

// getProjectGuide answers "what tech-stack template is this project on,
// and what does its agent guide say" — the MCP equivalent of the
// GET /projects/{id}/agents.md REST endpoint, but structured per-fragment
// (with category/version/divergence metadata) rather than flattened to
// one markdown blob.
func (s *toolServer) getProjectGuide(ctx context.Context, _ *mcp.CallToolRequest, args getProjectGuideArgs) (*mcp.CallToolResult, projectGuideResult, error) {
	projectID, err := parseUUID(args.ProjectID)
	if err != nil {
		return errorResult("invalid projectId %q", args.ProjectID), projectGuideResult{}, nil
	}
	project, err := s.q.GetProject(ctx, projectID)
	if err != nil {
		return errorResult("project %q not found", args.ProjectID), projectGuideResult{}, nil
	}
	fragments, err := s.q.ListProjectGuideFragments(ctx, projectID)
	if err != nil {
		return nil, projectGuideResult{}, err
	}
	out := projectGuideResult{
		TemplateID:   nullableUID(project.TemplateID),
		TemplateName: nullableText(project.TemplateName),
		Fragments:    make([]dto.ProjectGuideFragment, len(fragments)),
	}
	for i, f := range fragments {
		out.Fragments[i] = dto.ProjectGuideFragmentFromRow(f)
	}
	return nil, out, nil
}

const listWorkLogLimit = 50

type listWorkLogArgs struct {
	ProjectID string `json:"projectId,omitempty" jsonschema:"filter by project id, from list_projects; omit to see recent entries across all projects"`
	Search    string `json:"search,omitempty" jsonschema:"free-text search over title and body"`
}

func (s *toolServer) listWorkLog(ctx context.Context, _ *mcp.CallToolRequest, args listWorkLogArgs) (*mcp.CallToolResult, []dto.WorkLog, error) {
	params := db.ListWorkLogsParams{LimitCount: listWorkLogLimit}
	if args.ProjectID != "" {
		id, err := parseUUID(args.ProjectID)
		if err != nil {
			return errorResult("invalid projectId %q", args.ProjectID), nil, nil
		}
		params.ProjectID = id
	}
	if args.Search != "" {
		params.Search = pgtype.Text{String: args.Search, Valid: true}
	}
	rows, err := s.q.ListWorkLogs(ctx, params)
	if err != nil {
		return nil, nil, err
	}
	out := make([]dto.WorkLog, len(rows))
	for i, r := range rows {
		out[i] = dto.WorkLogFromGlobalRow(r)
	}
	return nil, out, nil
}
