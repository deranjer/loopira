package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type labelBody struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

type listLabelsInput struct {
	TeamID string `query:"teamId" required:"true"`
}

type listLabelsOutput struct {
	Body []labelBody
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
		out := &listLabelsOutput{Body: make([]labelBody, len(labels))}
		for i, l := range labels {
			out.Body[i] = labelBody{ID: uid(l.ID), Name: l.Name, Color: l.Color}
		}
		return out, nil
	})
}
