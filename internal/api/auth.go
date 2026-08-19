package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/deranjer/loopira/internal/auth"
	"github.com/deranjer/loopira/internal/dto"
)

type loginInput struct {
	Body struct {
		Email    string `json:"email" example:"admin@loopira.local"`
		Password string `json:"password"`
	}
}

type loginOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
	Body      dto.User
}

type logoutInput struct {
	Session string `cookie:"loopira_session"`
}

type logoutOutput struct {
	SetCookie http.Cookie `header:"Set-Cookie"`
	Body      struct {
		Status string `json:"status"`
	}
}

type meOutput struct {
	Body dto.User
}

func (s *Server) registerAuthRoutes() {
	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "login",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/login",
		Summary:     "Log in with email and password",
		Tags:        []string{"Auth"},
	}, func(ctx context.Context, input *loginInput) (*loginOutput, error) {
		user, err := s.q.GetUserByEmail(ctx, input.Body.Email)
		if err != nil || !user.PasswordHash.Valid || !auth.CheckPassword(user.PasswordHash.String, input.Body.Password) {
			return nil, huma.Error401Unauthorized("invalid email or password")
		}
		cookie, err := s.mgr.CreateSession(ctx, user.ID)
		if err != nil {
			return nil, err
		}
		out := &loginOutput{SetCookie: cookie}
		out.Body = dto.UserFromRow(user)
		return out, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "logout",
		Method:      http.MethodPost,
		Path:        "/api/v1/auth/logout",
		Summary:     "Log out",
		Tags:        []string{"Auth"},
	}, func(ctx context.Context, input *logoutInput) (*logoutOutput, error) {
		_ = s.mgr.DeleteSession(ctx, input.Session)
		out := &logoutOutput{SetCookie: auth.ClearedCookie()}
		out.Body.Status = "ok"
		return out, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "me",
		Method:      http.MethodGet,
		Path:        "/api/v1/auth/me",
		Summary:     "Get the current authenticated user",
		Tags:        []string{"Auth"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *struct{}) (*meOutput, error) {
		userID, _ := auth.UserID(ctx)
		id, err := mustUUID(userID)
		if err != nil {
			return nil, huma.Error401Unauthorized("login required")
		}
		user, err := s.q.GetUserByID(ctx, id)
		if err != nil {
			return nil, huma.Error401Unauthorized("login required")
		}
		out := &meOutput{}
		out.Body = dto.UserFromRow(user)
		return out, nil
	})
}
