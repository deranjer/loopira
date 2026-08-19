package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/deranjer/loopira/internal/auth"
	"github.com/deranjer/loopira/internal/db"
	"github.com/deranjer/loopira/internal/dto"
)

type listCyclesInput struct {
	TeamID string `query:"teamId" required:"true"`
}

type listCyclesOutput struct {
	Body []dto.Cycle
}

type createCycleInput struct {
	Body struct {
		TeamID    string `json:"teamId"`
		StartDate string `json:"startDate" example:"2026-08-25"`
		EndDate   string `json:"endDate" example:"2026-09-08"`
	}
}

type cycleOutput struct {
	Body dto.Cycle
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

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "create-cycle",
		Method:      http.MethodPost,
		Path:        "/api/v1/cycles",
		Summary:     "Create a cycle (number is assigned automatically)",
		Tags:        []string{"Cycles"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *createCycleInput) (*cycleOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		teamID, err := mustUUID(input.Body.TeamID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid teamId")
		}
		startDate, err := mustDate(input.Body.StartDate)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid startDate, expected YYYY-MM-DD")
		}
		endDate, err := mustDate(input.Body.EndDate)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid endDate, expected YYYY-MM-DD")
		}
		cycle, err := s.q.CreateCycle(ctx, db.CreateCycleParams{
			TeamID:    teamID,
			StartDate: startDate,
			EndDate:   endDate,
		})
		if err != nil {
			return nil, err
		}
		return &cycleOutput{Body: dto.CycleFromNew(cycle)}, nil
	})
}
