package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/deranjer/loopira/internal/auth"
	"github.com/deranjer/loopira/internal/db"
	"github.com/deranjer/loopira/internal/dto"
)

type listViewsOutput struct {
	Body []dto.View
}

type createViewInput struct {
	Body struct {
		Name       string          `json:"name" minLength:"1"`
		Definition json.RawMessage `json:"definition"`
		Shared     bool            `json:"shared,omitempty"`
	}
}

type updateViewInput struct {
	ID   string `path:"id"`
	Body struct {
		Name       string          `json:"name" minLength:"1"`
		Definition json.RawMessage `json:"definition"`
		Shared     bool            `json:"shared,omitempty"`
	}
}

type deleteViewInput struct {
	ID string `path:"id"`
}

type viewOutput struct {
	Body dto.View
}

func (s *Server) registerViewRoutes() {
	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "list-views",
		Method:      http.MethodGet,
		Path:        "/api/v1/views",
		Summary:     "List saved views owned by the current user, plus shared views",
		Tags:        []string{"Views"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *struct{}) (*listViewsOutput, error) {
		userIDStr, _ := auth.UserID(ctx)
		userID, err := mustUUID(userIDStr)
		if err != nil {
			return nil, huma.Error401Unauthorized("login required")
		}
		rows, err := s.q.ListViews(ctx, userID)
		if err != nil {
			return nil, err
		}
		out := &listViewsOutput{Body: make([]dto.View, len(rows))}
		for i, v := range rows {
			out.Body[i] = dto.ViewFromRow(v)
		}
		return out, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "create-view",
		Method:      http.MethodPost,
		Path:        "/api/v1/views",
		Summary:     "Save a new view",
		Tags:        []string{"Views"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *createViewInput) (*viewOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		userIDStr, _ := auth.UserID(ctx)
		userID, err := mustUUID(userIDStr)
		if err != nil {
			return nil, huma.Error401Unauthorized("login required")
		}
		created, err := s.q.CreateView(ctx, db.CreateViewParams{
			OwnerID:    userID,
			Name:       input.Body.Name,
			Definition: input.Body.Definition,
			Shared:     input.Body.Shared,
		})
		if err != nil {
			return nil, err
		}
		return &viewOutput{Body: dto.ViewFromRow(created)}, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "update-view",
		Method:      http.MethodPatch,
		Path:        "/api/v1/views/{id}",
		Summary:     "Update a saved view (owner only)",
		Tags:        []string{"Views"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *updateViewInput) (*viewOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		userIDStr, _ := auth.UserID(ctx)
		userID, err := mustUUID(userIDStr)
		if err != nil {
			return nil, huma.Error401Unauthorized("login required")
		}
		updated, err := s.q.UpdateView(ctx, db.UpdateViewParams{
			ID:         id,
			OwnerID:    userID,
			Name:       input.Body.Name,
			Definition: input.Body.Definition,
			Shared:     input.Body.Shared,
		})
		if err != nil {
			return nil, huma.Error404NotFound("view not found")
		}
		return &viewOutput{Body: dto.ViewFromRow(updated)}, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "delete-view",
		Method:      http.MethodDelete,
		Path:        "/api/v1/views/{id}",
		Summary:     "Delete a saved view (owner only)",
		Tags:        []string{"Views"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *deleteViewInput) (*statusOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		userIDStr, _ := auth.UserID(ctx)
		userID, err := mustUUID(userIDStr)
		if err != nil {
			return nil, huma.Error401Unauthorized("login required")
		}
		if err := s.q.DeleteView(ctx, db.DeleteViewParams{ID: id, OwnerID: userID}); err != nil {
			return nil, err
		}
		out := &statusOutput{}
		out.Body.Status = "ok"
		return out, nil
	})
}
