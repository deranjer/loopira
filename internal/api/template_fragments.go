package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/deranjer/loopira/internal/auth"
	"github.com/deranjer/loopira/internal/db"
	"github.com/deranjer/loopira/internal/dto"
)

type listTemplateFragmentsOutput struct {
	Body []dto.TemplateFragment
}

type getTemplateFragmentInput struct {
	ID string `path:"id"`
}

type templateFragmentOutput struct {
	Body dto.TemplateFragment
}

type createTemplateFragmentInput struct {
	Body struct {
		Name     string `json:"name" minLength:"1"`
		Category string `json:"category,omitempty"`
		Content  string `json:"content,omitempty"`
	}
}

type updateTemplateFragmentInput struct {
	ID   string `path:"id"`
	Body struct {
		Name     string `json:"name" minLength:"1"`
		Category string `json:"category,omitempty"`
		Content  string `json:"content,omitempty"`
	}
}

type deleteTemplateFragmentInput struct {
	ID string `path:"id"`
}

type fragmentUsageInput struct {
	ID string `path:"id"`
}

type fragmentUsageOutput struct {
	Body []dto.FragmentUsage
}

type pushFragmentUpdateInput struct {
	ID   string `path:"id"`
	Body struct {
		ProjectGuideFragmentIDs []string `json:"projectGuideFragmentIds"`
	}
}

type pushFragmentUpdateOutput struct {
	Body []dto.ProjectGuideFragment
}

func (s *Server) registerTemplateFragmentRoutes() {
	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "list-template-fragments",
		Method:      http.MethodGet,
		Path:        "/api/v1/template-fragments",
		Summary:     "List reusable template fragments (building blocks)",
		Tags:        []string{"Templates"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *struct{}) (*listTemplateFragmentsOutput, error) {
		rows, err := s.q.ListTemplateFragments(ctx)
		if err != nil {
			return nil, err
		}
		out := &listTemplateFragmentsOutput{Body: make([]dto.TemplateFragment, len(rows))}
		for i, f := range rows {
			out.Body[i] = dto.TemplateFragmentFromRow(f)
		}
		return out, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "get-template-fragment",
		Method:      http.MethodGet,
		Path:        "/api/v1/template-fragments/{id}",
		Summary:     "Get a single template fragment",
		Tags:        []string{"Templates"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *getTemplateFragmentInput) (*templateFragmentOutput, error) {
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		row, err := s.q.GetTemplateFragment(ctx, id)
		if err != nil {
			return nil, huma.Error404NotFound("fragment not found")
		}
		return &templateFragmentOutput{Body: dto.TemplateFragmentFromGetRow(row)}, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "create-template-fragment",
		Method:      http.MethodPost,
		Path:        "/api/v1/template-fragments",
		Summary:     "Create a template fragment",
		Tags:        []string{"Templates"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *createTemplateFragmentInput) (*templateFragmentOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		userIDStr, _ := auth.UserID(ctx)
		userID, err := mustUUID(userIDStr)
		if err != nil {
			return nil, huma.Error401Unauthorized("login required")
		}
		created, err := s.q.CreateTemplateFragment(ctx, db.CreateTemplateFragmentParams{
			Name:      input.Body.Name,
			Category:  pgtype.Text{String: input.Body.Category, Valid: input.Body.Category != ""},
			Content:   input.Body.Content,
			CreatedBy: userID,
		})
		if err != nil {
			return nil, err
		}
		return &templateFragmentOutput{Body: dto.TemplateFragmentFromRecord(created)}, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "update-template-fragment",
		Method:      http.MethodPatch,
		Path:        "/api/v1/template-fragments/{id}",
		Summary:     "Update a template fragment's content (bumps its version)",
		Tags:        []string{"Templates"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *updateTemplateFragmentInput) (*templateFragmentOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		updated, err := s.q.UpdateTemplateFragment(ctx, db.UpdateTemplateFragmentParams{
			ID:       id,
			Name:     input.Body.Name,
			Category: pgtype.Text{String: input.Body.Category, Valid: input.Body.Category != ""},
			Content:  input.Body.Content,
		})
		if err != nil {
			return nil, huma.Error404NotFound("fragment not found")
		}
		return &templateFragmentOutput{Body: dto.TemplateFragmentFromRecord(updated)}, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "delete-template-fragment",
		Method:      http.MethodDelete,
		Path:        "/api/v1/template-fragments/{id}",
		Summary:     "Delete a template fragment",
		Tags:        []string{"Templates"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *deleteTemplateFragmentInput) (*statusOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		if err := s.q.DeleteTemplateFragment(ctx, id); err != nil {
			return nil, err
		}
		out := &statusOutput{}
		out.Body.Status = "ok"
		return out, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "get-fragment-usage",
		Method:      http.MethodGet,
		Path:        "/api/v1/template-fragments/{id}/usage",
		Summary:     "List projects using a fragment, and whether each copy has diverged",
		Tags:        []string{"Templates"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *fragmentUsageInput) (*fragmentUsageOutput, error) {
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		rows, err := s.q.ListFragmentUsage(ctx, id)
		if err != nil {
			return nil, err
		}
		out := &fragmentUsageOutput{Body: make([]dto.FragmentUsage, len(rows))}
		for i, r := range rows {
			out.Body[i] = dto.FragmentUsageFromRow(r)
		}
		return out, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "push-fragment-update",
		Method:      http.MethodPost,
		Path:        "/api/v1/template-fragments/{id}/push",
		Summary:     "Push a fragment's current content to selected, unchanged project copies",
		Tags:        []string{"Templates"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *pushFragmentUpdateInput) (*pushFragmentUpdateOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		fragment, err := s.q.GetTemplateFragment(ctx, id)
		if err != nil {
			return nil, huma.Error404NotFound("fragment not found")
		}
		ids := make([]pgtype.UUID, 0, len(input.Body.ProjectGuideFragmentIDs))
		for _, s := range input.Body.ProjectGuideFragmentIDs {
			pid, err := mustUUID(s)
			if err != nil {
				return nil, huma.Error400BadRequest("invalid projectGuideFragmentIds entry")
			}
			ids = append(ids, pid)
		}
		updated, err := s.q.PushFragmentUpdate(ctx, db.PushFragmentUpdateParams{
			FragmentID:  id,
			Name:        fragment.Name,
			Content:     fragment.Content,
			BaseVersion: pgtype.Int4{Int32: fragment.Version, Valid: true},
			Ids:         ids,
		})
		if err != nil {
			return nil, err
		}
		out := &pushFragmentUpdateOutput{Body: make([]dto.ProjectGuideFragment, len(updated))}
		for i, u := range updated {
			out.Body[i] = dto.ProjectGuideFragmentFromRow(u)
		}
		return out, nil
	})
}
