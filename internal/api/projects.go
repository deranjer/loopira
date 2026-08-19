package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/deranjer/loopira/internal/dto"
)

type listProjectsInput struct {
	TeamID string `query:"teamId" required:"true"`
}

type listProjectsOutput struct {
	Body []dto.Project
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
}
