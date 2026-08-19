package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/deranjer/loopira/internal/dto"
)

type listCyclesInput struct {
	TeamID string `query:"teamId" required:"true"`
}

type listCyclesOutput struct {
	Body []dto.Cycle
}

func (s *Server) registerCycleRoutes() {
	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "list-cycles",
		Method:      http.MethodGet,
		Path:        "/api/v1/cycles",
		Summary:     "List cycles for a team",
		Tags:        []string{"Cycles"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *listCyclesInput) (*listCyclesOutput, error) {
		teamID, err := mustUUID(input.TeamID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid teamId")
		}
		cycles, err := s.q.ListCycles(ctx, teamID)
		if err != nil {
			return nil, err
		}
		out := &listCyclesOutput{Body: make([]dto.Cycle, len(cycles))}
		for i, c := range cycles {
			out.Body[i] = dto.CycleFromRow(c)
		}
		return out, nil
	})
}
