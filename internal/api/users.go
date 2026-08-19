package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/deranjer/loopira/internal/dto"
)

type listUsersOutput struct {
	Body []dto.User
}

func (s *Server) registerUserRoutes() {
	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "list-users",
		Method:      http.MethodGet,
		Path:        "/api/v1/users",
		Summary:     "List workspace members",
		Tags:        []string{"Users"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *struct{}) (*listUsersOutput, error) {
		users, err := s.q.ListUsers(ctx)
		if err != nil {
			return nil, err
		}
		out := &listUsersOutput{Body: make([]dto.User, len(users))}
		for i, u := range users {
			out.Body[i] = dto.UserFromRow(u)
		}
		return out, nil
	})
}
