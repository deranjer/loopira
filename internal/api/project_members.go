package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/deranjer/loopira/internal/auth"
	"github.com/deranjer/loopira/internal/db"
	"github.com/deranjer/loopira/internal/dto"
)

type listProjectMembersInput struct {
	ProjectID string `path:"id"`
}

type listProjectMembersOutput struct {
	Body []dto.User
}

type addProjectMemberInput struct {
	ProjectID string `path:"id"`
	Body      struct {
		UserID string `json:"userId"`
	}
}

type removeProjectMemberInput struct {
	ProjectID string `path:"id"`
	UserID    string `path:"userId"`
}

type statusOutput struct {
	Body struct {
		Status string `json:"status"`
	}
}

func (s *Server) registerProjectMemberRoutes() {
	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "list-project-members",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{id}/members",
		Summary:     "List a project's members",
		Tags:        []string{"Projects"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *listProjectMembersInput) (*listProjectMembersOutput, error) {
		projectID, err := mustUUID(input.ProjectID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		rows, err := s.q.ListProjectMembers(ctx, projectID)
		if err != nil {
			return nil, err
		}
		out := &listProjectMembersOutput{Body: make([]dto.User, len(rows))}
		for i, u := range rows {
			out.Body[i] = dto.UserFromRow(u)
		}
		return out, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "add-project-member",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{id}/members",
		Summary:     "Add a member to a project",
		Tags:        []string{"Projects"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *addProjectMemberInput) (*listProjectMembersOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		projectID, err := mustUUID(input.ProjectID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		userID, err := mustUUID(input.Body.UserID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid userId")
		}
		if err := s.q.AddProjectMember(ctx, db.AddProjectMemberParams{ProjectID: projectID, UserID: userID}); err != nil {
			return nil, err
		}
		rows, err := s.q.ListProjectMembers(ctx, projectID)
		if err != nil {
			return nil, err
		}
		out := &listProjectMembersOutput{Body: make([]dto.User, len(rows))}
		for i, u := range rows {
			out.Body[i] = dto.UserFromRow(u)
		}
		return out, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "remove-project-member",
		Method:      http.MethodDelete,
		Path:        "/api/v1/projects/{id}/members/{userId}",
		Summary:     "Remove a member from a project",
		Tags:        []string{"Projects"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *removeProjectMemberInput) (*statusOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		projectID, err := mustUUID(input.ProjectID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		userID, err := mustUUID(input.UserID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid userId")
		}
		if err := s.q.RemoveProjectMember(ctx, db.RemoveProjectMemberParams{ProjectID: projectID, UserID: userID}); err != nil {
			return nil, err
		}
		out := &statusOutput{}
		out.Body.Status = "ok"
		return out, nil
	})
}
