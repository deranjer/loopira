package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/deranjer/loopira/internal/db"
	"github.com/deranjer/loopira/internal/dto"
)

type setupStatusOutput struct {
	Body struct {
		Required bool `json:"required"`
	}
}

type setupInput struct {
	Body struct {
		Name     string `json:"name"`
		Email    string `json:"email" example:"admin@loopira.local"`
		Password string `json:"password" minLength:"8"`
	}
}

type setupOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
	Body      dto.User
}

// registerSetupRoutes wires the first-run setup wizard: an unauthenticated
// status check the frontend polls to decide whether to redirect to /setup,
// and the one-time completion endpoint that creates the admin user. Once a
// user exists, db.SetupRequired permanently reports false, so /setup can
// never run again — there's no separate "setup complete" flag to track.
func (s *Server) registerSetupRoutes() {
	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "setup-status",
		Method:      http.MethodGet,
		Path:        "/api/v1/setup/status",
		Summary:     "Check whether first-run setup is still required",
		Tags:        []string{"Setup"},
	}, func(ctx context.Context, input *struct{}) (*setupStatusOutput, error) {
		required, err := db.SetupRequired(ctx, s.q)
		if err != nil {
			return nil, err
		}
		out := &setupStatusOutput{}
		out.Body.Required = required
		return out, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "setup-complete",
		Method:      http.MethodPost,
		Path:        "/api/v1/setup",
		Summary:     "Create the admin user and finish first-run setup",
		Tags:        []string{"Setup"},
	}, func(ctx context.Context, input *setupInput) (*setupOutput, error) {
		required, err := db.SetupRequired(ctx, s.q)
		if err != nil {
			return nil, err
		}
		if !required {
			return nil, huma.Error409Conflict("setup has already been completed")
		}

		user, err := db.CompleteSetup(ctx, s.q, input.Body.Name, input.Body.Email, input.Body.Password)
		if err != nil {
			return nil, err
		}

		cookie, err := s.mgr.CreateSession(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		out := &setupOutput{SetCookie: cookie}
		out.Body = dto.UserFromRow(user)
		return out, nil
	})
}
