package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/deranjer/loopira/internal/auth"
	"github.com/deranjer/loopira/internal/db"
)

type teamBody struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Key  string `json:"key"`
}

type listTeamsOutput struct {
	Body []teamBody
}

type updateTeamInput struct {
	ID   string `path:"id"`
	Body struct {
		Name string `json:"name" minLength:"1"`
	}
}

type teamOutput struct {
	Body teamBody
}

func teamToBody(t db.Team) teamBody {
	return teamBody{ID: uid(t.ID), Name: t.Name, Key: t.Key}
}

func (s *Server) registerTeamRoutes() {
	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "list-teams",
		Method:      http.MethodGet,
		Path:        "/api/v1/teams",
		Summary:     "List teams",
		Tags:        []string{"Teams"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *struct{}) (*listTeamsOutput, error) {
		teams, err := s.q.ListTeams(ctx)
		if err != nil {
			return nil, err
		}
		out := &listTeamsOutput{Body: make([]teamBody, len(teams))}
		for i, t := range teams {
			out.Body[i] = teamToBody(t)
		}
		return out, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "update-team",
		Method:      http.MethodPatch,
		Path:        "/api/v1/teams/{id}",
		Summary:     "Rename a team (workspace settings)",
		Tags:        []string{"Teams"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *updateTeamInput) (*teamOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid team id")
		}
		team, err := s.q.UpdateTeamName(ctx, db.UpdateTeamNameParams{ID: id, Name: input.Body.Name})
		if err != nil {
			return nil, huma.Error404NotFound("team not found")
		}
		out := &teamOutput{Body: teamToBody(team)}
		return out, nil
	})
}
