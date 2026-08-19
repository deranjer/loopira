package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/deranjer/loopira/internal/auth"
	"github.com/deranjer/loopira/internal/db"
	"github.com/deranjer/loopira/internal/dto"
)

type listProjectsInput struct {
	TeamID string `query:"teamId" required:"true"`
}

type listProjectsOutput struct {
	Body []dto.Project
}

type createProjectInput struct {
	Body struct {
		TeamID      string `json:"teamId"`
		Name        string `json:"name" minLength:"1"`
		Description string `json:"description,omitempty"`
	}
}

type projectOutput struct {
	Body dto.Project
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
		project, err := s.q.CreateProject(ctx, db.CreateProjectParams{
			TeamID:      teamID,
			Name:        input.Body.Name,
			Description: pgtype.Text{String: input.Body.Description, Valid: input.Body.Description != ""},
		})
		if err != nil {
			return nil, err
		}
		return &projectOutput{Body: dto.ProjectFromNew(project)}, nil
	})
}
