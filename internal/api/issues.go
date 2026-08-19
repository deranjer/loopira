package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/deranjer/loopira/internal/auth"
	"github.com/deranjer/loopira/internal/db"
	"github.com/deranjer/loopira/internal/dto"
	"github.com/deranjer/loopira/internal/ws"
)

type listIssuesInput struct {
	TeamID    string `query:"teamId" required:"true"`
	Status    string `query:"status"`
	ProjectID string `query:"projectId"`
	CycleID   string `query:"cycleId"`
}

type listIssuesOutput struct {
	Body []dto.Issue
}

type getIssueInput struct {
	ID string `path:"id"`
}

type issueOutput struct {
	Body dto.Issue
}

type createIssueInput struct {
	Body struct {
		TeamID      string  `json:"teamId"`
		Title       string  `json:"title" minLength:"1"`
		Description string  `json:"description"`
		Priority    int16   `json:"priority" minimum:"0" maximum:"4"`
		AssigneeID  *string `json:"assigneeId,omitempty"`
		ProjectID   *string `json:"projectId,omitempty"`
		LabelID     *string `json:"labelId,omitempty"`
	}
}

type updateIssueStatusInput struct {
	ID   string `path:"id"`
	Body struct {
		Status string `json:"status" enum:"backlog,todo,in_progress,done,canceled"`
	}
}

type updateIssueDetailsInput struct {
	ID   string `path:"id"`
	Body struct {
		Title       string  `json:"title" minLength:"1"`
		Description string  `json:"description"`
		Priority    int16   `json:"priority" minimum:"0" maximum:"4"`
		AssigneeID  *string `json:"assigneeId,omitempty"`
		ProjectID   *string `json:"projectId,omitempty"`
		CycleID     *string `json:"cycleId,omitempty"`
		LabelID     *string `json:"labelId,omitempty"`
	}
}

func optionalUUID(s *string) (pgtype.UUID, error) {
	if s == nil || *s == "" {
		return pgtype.UUID{}, nil
	}
	return mustUUID(*s)
}

// applyIssueLabel replaces an issue's label set with zero or one label —
// the UI only ever shows a single label per issue even though the schema
// supports many-to-many via issue_labels.
func (s *Server) applyIssueLabel(ctx context.Context, issueID pgtype.UUID, labelID *string) error {
	if err := s.q.ClearIssueLabels(ctx, issueID); err != nil {
		return err
	}
	if labelID == nil || *labelID == "" {
		return nil
	}
	id, err := mustUUID(*labelID)
	if err != nil {
		return huma.Error400BadRequest("invalid labelId")
	}
	return s.q.AddIssueLabel(ctx, db.AddIssueLabelParams{IssueID: issueID, LabelID: id})
}

func (s *Server) registerIssueRoutes() {
	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "list-issues",
		Method:      http.MethodGet,
		Path:        "/api/v1/issues",
		Summary:     "List issues for a team, optionally filtered",
		Tags:        []string{"Issues"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *listIssuesInput) (*listIssuesOutput, error) {
		teamID, err := mustUUID(input.TeamID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid teamId")
		}
		projectID, err := optionalUUID(strPtr(input.ProjectID))
		if err != nil {
			return nil, huma.Error400BadRequest("invalid projectId")
		}
		cycleID, err := optionalUUID(strPtr(input.CycleID))
		if err != nil {
			return nil, huma.Error400BadRequest("invalid cycleId")
		}
		params := db.ListIssuesParams{TeamID: teamID}
		if input.Status != "" {
			params.Status = pgtype.Text{String: input.Status, Valid: true}
		}
		if projectID.Valid {
			params.ProjectID = projectID
		}
		if cycleID.Valid {
			params.CycleID = cycleID
		}
		rows, err := s.q.ListIssues(ctx, params)
		if err != nil {
			return nil, err
		}
		out := &listIssuesOutput{Body: make([]dto.Issue, len(rows))}
		for i, r := range rows {
			out.Body[i] = dto.IssueFromListRow(r)
		}
		return out, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "get-issue",
		Method:      http.MethodGet,
		Path:        "/api/v1/issues/{id}",
		Summary:     "Get a single issue",
		Tags:        []string{"Issues"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *getIssueInput) (*issueOutput, error) {
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		row, err := s.q.GetIssue(ctx, id)
		if err != nil {
			return nil, huma.Error404NotFound("issue not found")
		}
		return &issueOutput{Body: dto.IssueFromGetRow(row)}, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "create-issue",
		Method:      http.MethodPost,
		Path:        "/api/v1/issues",
		Summary:     "Create an issue",
		Tags:        []string{"Issues"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *createIssueInput) (*issueOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		teamID, err := mustUUID(input.Body.TeamID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid teamId")
		}
		assigneeID, err := optionalUUID(input.Body.AssigneeID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid assigneeId")
		}
		projectID, err := optionalUUID(input.Body.ProjectID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid projectId")
		}
		userIDStr, _ := auth.UserID(ctx)
		createdBy, err := mustUUID(userIDStr)
		if err != nil {
			return nil, huma.Error401Unauthorized("login required")
		}
		created, err := s.q.CreateIssue(ctx, db.CreateIssueParams{
			TeamID:      teamID,
			Title:       input.Body.Title,
			Description: input.Body.Description,
			Priority:    input.Body.Priority,
			AssigneeID:  assigneeID,
			ProjectID:   projectID,
			CreatedBy:   createdBy,
		})
		if err != nil {
			return nil, err
		}
		if err := s.applyIssueLabel(ctx, created.ID, input.Body.LabelID); err != nil {
			return nil, err
		}
		row, err := s.q.GetIssue(ctx, created.ID)
		if err != nil {
			return nil, err
		}
		body := dto.IssueFromGetRow(row)
		s.hub.Broadcast(ws.Event{Type: "issue.created", TeamID: input.Body.TeamID, Payload: body})
		return &issueOutput{Body: body}, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "update-issue-status",
		Method:      http.MethodPatch,
		Path:        "/api/v1/issues/{id}/status",
		Summary:     "Update an issue's status (board drag-and-drop)",
		Tags:        []string{"Issues"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *updateIssueStatusInput) (*issueOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		updated, err := s.q.UpdateIssueStatus(ctx, db.UpdateIssueStatusParams{ID: id, Status: input.Body.Status})
		if err != nil {
			return nil, huma.Error404NotFound("issue not found")
		}
		row, err := s.q.GetIssue(ctx, updated.ID)
		if err != nil {
			return nil, err
		}
		body := dto.IssueFromGetRow(row)
		s.hub.Broadcast(ws.Event{Type: "issue.updated", TeamID: uid(updated.TeamID), Payload: body})
		return &issueOutput{Body: body}, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "update-issue-details",
		Method:      http.MethodPatch,
		Path:        "/api/v1/issues/{id}",
		Summary:     "Update an issue's details (title/description/priority/assignee/project/cycle)",
		Tags:        []string{"Issues"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *updateIssueDetailsInput) (*issueOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		assigneeID, err := optionalUUID(input.Body.AssigneeID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid assigneeId")
		}
		projectID, err := optionalUUID(input.Body.ProjectID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid projectId")
		}
		cycleID, err := optionalUUID(input.Body.CycleID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid cycleId")
		}
		updated, err := s.q.UpdateIssueDetails(ctx, db.UpdateIssueDetailsParams{
			ID:          id,
			Title:       input.Body.Title,
			Description: input.Body.Description,
			Priority:    input.Body.Priority,
			AssigneeID:  assigneeID,
			ProjectID:   projectID,
			CycleID:     cycleID,
		})
		if err != nil {
			return nil, huma.Error404NotFound("issue not found")
		}
		if err := s.applyIssueLabel(ctx, updated.ID, input.Body.LabelID); err != nil {
			return nil, err
		}
		row, err := s.q.GetIssue(ctx, updated.ID)
		if err != nil {
			return nil, err
		}
		body := dto.IssueFromGetRow(row)
		s.hub.Broadcast(ws.Event{Type: "issue.updated", TeamID: uid(updated.TeamID), Payload: body})
		return &issueOutput{Body: body}, nil
	})
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
