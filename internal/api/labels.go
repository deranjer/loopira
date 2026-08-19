package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/deranjer/loopira/internal/auth"
	"github.com/deranjer/loopira/internal/db"
	"github.com/deranjer/loopira/internal/dto"
)

type listLabelsInput struct {
	TeamID string `query:"teamId" required:"true"`
}

type listLabelsOutput struct {
	Body []dto.Label
}

type createLabelInput struct {
	Body struct {
		TeamID string `json:"teamId"`
		Name   string `json:"name" minLength:"1"`
		Color  string `json:"color" minLength:"1"`
	}
}

type labelOutput struct {
	Body dto.Label
}

func (s *Server) registerLabelRoutes() {
	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "list-labels",
		Method:      http.MethodGet,
		Path:        "/api/v1/labels",
		Summary:     "List labels for a team",
		Tags:        []string{"Labels"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *listLabelsInput) (*listLabelsOutput, error) {
		teamID, err := mustUUID(input.TeamID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid teamId")
		}
		labels, err := s.q.ListLabels(ctx, teamID)
		if err != nil {
			return nil, err
		}
		out := &listLabelsOutput{Body: make([]dto.Label, len(labels))}
		for i, l := range labels {
			out.Body[i] = dto.LabelFromRow(l)
		}
		return out, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "create-label",
		Method:      http.MethodPost,
		Path:        "/api/v1/labels",
		Summary:     "Create a label",
		Tags:        []string{"Labels"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *createLabelInput) (*labelOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		teamID, err := mustUUID(input.Body.TeamID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid teamId")
		}
		label, err := s.q.CreateLabel(ctx, db.CreateLabelParams{
			TeamID: teamID,
			Name:   input.Body.Name,
			Color:  input.Body.Color,
		})
		if err != nil {
			return nil, err
		}
		return &labelOutput{Body: dto.LabelFromRow(label)}, nil
	})
}
