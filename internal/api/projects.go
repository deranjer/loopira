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

type listProjectsInput struct {
	TeamID string `query:"teamId" required:"true"`
}

type listProjectsOutput struct {
	Body []dto.Project
}

type getProjectInput struct {
	ID string `path:"id"`
}

type createProjectInput struct {
	Body struct {
		TeamID      string  `json:"teamId"`
		Name        string  `json:"name" minLength:"1"`
		Description string  `json:"description,omitempty"`
		Status      string  `json:"status,omitempty" enum:"backlog,planned,in_progress,paused,completed,canceled"`
		LeadID      *string `json:"leadId,omitempty"`
		Priority    int16   `json:"priority,omitempty" minimum:"0" maximum:"4"`
		TargetDate  *string `json:"targetDate,omitempty"`
	}
}

type updateProjectInput struct {
	ID   string `path:"id"`
	Body struct {
		Name        string  `json:"name" minLength:"1"`
		Description string  `json:"description,omitempty"`
		Status      string  `json:"status" enum:"backlog,planned,in_progress,paused,completed,canceled"`
		LeadID      *string `json:"leadId,omitempty"`
		Priority    int16   `json:"priority" minimum:"0" maximum:"4"`
		TargetDate  *string `json:"targetDate,omitempty"`
	}
}

type projectOutput struct {
	Body dto.Project
}

// optionalDate parses an optional "2006-01-02" date string, returning a
// zero-value invalid pgtype.Date when s is nil or empty — mirrors
// optionalUUID's nil-safe handling in issues.go.
func optionalDate(s *string) (pgtype.Date, error) {
	if s == nil || *s == "" {
		return pgtype.Date{}, nil
	}
	return mustDate(*s)
}

func (s *Server) registerProjectRoutes() {
	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "list-projects",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects",
		Summary:     "List projects for a team",
		Tags:        []string{"Projects"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *listProjectsInput) (*listProjectsOutput, error) {
		teamID, err := mustUUID(input.TeamID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid teamId")
		}
		projects, err := s.q.ListProjects(ctx, teamID)
		if err != nil {
			return nil, err
		}
		out := &listProjectsOutput{Body: make([]dto.Project, len(projects))}
		for i, p := range projects {
			out.Body[i] = dto.ProjectFromRow(p)
		}
		return out, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "create-project",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects",
		Summary:     "Create a project",
		Tags:        []string{"Projects"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *createProjectInput) (*projectOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		teamID, err := mustUUID(input.Body.TeamID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid teamId")
		}
		leadID, err := optionalUUID(input.Body.LeadID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid leadId")
		}
		targetDate, err := optionalDate(input.Body.TargetDate)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid targetDate")
		}
		status := input.Body.Status
		if status == "" {
			status = "backlog"
		}
		created, err := s.q.CreateProject(ctx, db.CreateProjectParams{
			TeamID:      teamID,
			Name:        input.Body.Name,
			Description: pgtype.Text{String: input.Body.Description, Valid: input.Body.Description != ""},
			Status:      status,
			LeadID:      leadID,
			Priority:    input.Body.Priority,
			TargetDate:  targetDate,
		})
		if err != nil {
			return nil, err
		}
		row, err := s.q.GetProject(ctx, created.ID)
		if err != nil {
			return nil, err
		}
		return &projectOutput{Body: dto.ProjectFromGetRow(row)}, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "get-project",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{id}",
		Summary:     "Get a single project",
		Tags:        []string{"Projects"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *getProjectInput) (*projectOutput, error) {
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		row, err := s.q.GetProject(ctx, id)
		if err != nil {
			return nil, huma.Error404NotFound("project not found")
		}
		return &projectOutput{Body: dto.ProjectFromGetRow(row)}, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "update-project",
		Method:      http.MethodPatch,
		Path:        "/api/v1/projects/{id}",
		Summary:     "Update a project's details (name/description/status/priority/lead/target date)",
		Tags:        []string{"Projects"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *updateProjectInput) (*projectOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		leadID, err := optionalUUID(input.Body.LeadID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid leadId")
		}
		targetDate, err := optionalDate(input.Body.TargetDate)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid targetDate")
		}
		updated, err := s.q.UpdateProject(ctx, db.UpdateProjectParams{
			ID:          id,
			Name:        input.Body.Name,
			Description: pgtype.Text{String: input.Body.Description, Valid: input.Body.Description != ""},
			Status:      input.Body.Status,
			LeadID:      leadID,
			Priority:    input.Body.Priority,
			TargetDate:  targetDate,
		})
		if err != nil {
			return nil, huma.Error404NotFound("project not found")
		}
		if leadID.Valid {
			// A project lead is expected to also be a member; upsert rather
			// than requiring the caller to add them separately first.
			if err := s.q.AddProjectMember(ctx, db.AddProjectMemberParams{ProjectID: id, UserID: leadID}); err != nil {
				return nil, err
			}
		}
		row, err := s.q.GetProject(ctx, updated.ID)
		if err != nil {
			return nil, err
		}
		body := dto.ProjectFromGetRow(row)
		s.hub.Broadcast(ws.Event{Type: "project.updated", TeamID: uid(updated.TeamID), Payload: body})
		return &projectOutput{Body: body}, nil
	})
}
