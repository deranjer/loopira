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

type listTemplatesOutput struct {
	Body []dto.Template
}

type getTemplateInput struct {
	ID string `path:"id"`
}

type templateOutput struct {
	Body dto.Template
}

type createTemplateInput struct {
	Body struct {
		Name        string `json:"name" minLength:"1"`
		Description string `json:"description,omitempty"`
	}
}

type updateTemplateInput struct {
	ID   string `path:"id"`
	Body struct {
		Name        string `json:"name" minLength:"1"`
		Description string `json:"description,omitempty"`
	}
}

type deleteTemplateInput struct {
	ID string `path:"id"`
}

type addTemplateFragmentInput struct {
	ID   string `path:"id"`
	Body struct {
		FragmentID string `json:"fragmentId"`
	}
}

type removeTemplateFragmentInput struct {
	ID         string `path:"id"`
	FragmentID string `path:"fragmentId"`
}

type reorderTemplateFragmentsInput struct {
	ID   string `path:"id"`
	Body struct {
		FragmentIDs []string `json:"fragmentIds"`
	}
}

// loadTemplateFragments fetches and attaches a template's ordered fragment
// list — split out of GetTemplate/ListTemplates since the list view doesn't
// need the full composition, only the detail/edit view does.
func (s *Server) loadTemplateFragments(ctx context.Context, t *dto.Template, id pgtype.UUID) error {
	links, err := s.q.ListTemplateLinks(ctx, id)
	if err != nil {
		return err
	}
	t.Fragments = make([]dto.TemplateFragmentRef, len(links))
	for i, l := range links {
		t.Fragments[i] = dto.TemplateFragmentRefFromRow(l)
	}
	return nil
}

func (s *Server) registerTemplateRoutes() {
	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "list-templates",
		Method:      http.MethodGet,
		Path:        "/api/v1/templates",
		Summary:     "List templates",
		Tags:        []string{"Templates"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *struct{}) (*listTemplatesOutput, error) {
		rows, err := s.q.ListTemplates(ctx)
		if err != nil {
			return nil, err
		}
		out := &listTemplatesOutput{Body: make([]dto.Template, len(rows))}
		for i, t := range rows {
			out.Body[i] = dto.TemplateFromRow(t)
		}
		return out, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "get-template",
		Method:      http.MethodGet,
		Path:        "/api/v1/templates/{id}",
		Summary:     "Get a template and its ordered fragment composition",
		Tags:        []string{"Templates"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *getTemplateInput) (*templateOutput, error) {
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		row, err := s.q.GetTemplate(ctx, id)
		if err != nil {
			return nil, huma.Error404NotFound("template not found")
		}
		body := dto.TemplateFromGetRow(row)
		if err := s.loadTemplateFragments(ctx, &body, id); err != nil {
			return nil, err
		}
		return &templateOutput{Body: body}, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "create-template",
		Method:      http.MethodPost,
		Path:        "/api/v1/templates",
		Summary:     "Create a template",
		Tags:        []string{"Templates"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *createTemplateInput) (*templateOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		userIDStr, _ := auth.UserID(ctx)
		userID, err := mustUUID(userIDStr)
		if err != nil {
			return nil, huma.Error401Unauthorized("login required")
		}
		created, err := s.q.CreateTemplate(ctx, db.CreateTemplateParams{
			Name:        input.Body.Name,
			Description: pgtype.Text{String: input.Body.Description, Valid: input.Body.Description != ""},
			CreatedBy:   userID,
		})
		if err != nil {
			return nil, err
		}
		return &templateOutput{Body: dto.TemplateFromRecord(created)}, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "update-template",
		Method:      http.MethodPatch,
		Path:        "/api/v1/templates/{id}",
		Summary:     "Rename/update a template's description",
		Tags:        []string{"Templates"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *updateTemplateInput) (*templateOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		updated, err := s.q.UpdateTemplate(ctx, db.UpdateTemplateParams{
			ID:          id,
			Name:        input.Body.Name,
			Description: pgtype.Text{String: input.Body.Description, Valid: input.Body.Description != ""},
		})
		if err != nil {
			return nil, huma.Error404NotFound("template not found")
		}
		body := dto.TemplateFromRecord(updated)
		if err := s.loadTemplateFragments(ctx, &body, id); err != nil {
			return nil, err
		}
		return &templateOutput{Body: body}, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "delete-template",
		Method:      http.MethodDelete,
		Path:        "/api/v1/templates/{id}",
		Summary:     "Delete a template",
		Tags:        []string{"Templates"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *deleteTemplateInput) (*statusOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		if err := s.q.DeleteTemplate(ctx, id); err != nil {
			return nil, err
		}
		out := &statusOutput{}
		out.Body.Status = "ok"
		return out, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "add-template-fragment",
		Method:      http.MethodPost,
		Path:        "/api/v1/templates/{id}/fragments",
		Summary:     "Append a fragment to a template's composition",
		Tags:        []string{"Templates"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *addTemplateFragmentInput) (*templateOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		fragmentID, err := mustUUID(input.Body.FragmentID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid fragmentId")
		}
		if err := s.q.AddTemplateFragment(ctx, db.AddTemplateFragmentParams{TemplateID: id, FragmentID: fragmentID}); err != nil {
			return nil, err
		}
		row, err := s.q.GetTemplate(ctx, id)
		if err != nil {
			return nil, huma.Error404NotFound("template not found")
		}
		body := dto.TemplateFromGetRow(row)
		if err := s.loadTemplateFragments(ctx, &body, id); err != nil {
			return nil, err
		}
		return &templateOutput{Body: body}, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "remove-template-fragment",
		Method:      http.MethodDelete,
		Path:        "/api/v1/templates/{id}/fragments/{fragmentId}",
		Summary:     "Remove a fragment from a template's composition",
		Tags:        []string{"Templates"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *removeTemplateFragmentInput) (*templateOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		fragmentID, err := mustUUID(input.FragmentID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid fragmentId")
		}
		if err := s.q.RemoveTemplateFragment(ctx, db.RemoveTemplateFragmentParams{TemplateID: id, FragmentID: fragmentID}); err != nil {
			return nil, err
		}
		row, err := s.q.GetTemplate(ctx, id)
		if err != nil {
			return nil, huma.Error404NotFound("template not found")
		}
		body := dto.TemplateFromGetRow(row)
		if err := s.loadTemplateFragments(ctx, &body, id); err != nil {
			return nil, err
		}
		return &templateOutput{Body: body}, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "reorder-template-fragments",
		Method:      http.MethodPatch,
		Path:        "/api/v1/templates/{id}/fragments/reorder",
		Summary:     "Set the composition order of a template's fragments",
		Tags:        []string{"Templates"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *reorderTemplateFragmentsInput) (*templateOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		for i, fragmentIDStr := range input.Body.FragmentIDs {
			fragmentID, err := mustUUID(fragmentIDStr)
			if err != nil {
				return nil, huma.Error400BadRequest("invalid fragmentIds entry")
			}
			if err := s.q.SetTemplateFragmentPosition(ctx, db.SetTemplateFragmentPositionParams{
				TemplateID: id,
				FragmentID: fragmentID,
				Position:   int32(i),
			}); err != nil {
				return nil, err
			}
		}
		row, err := s.q.GetTemplate(ctx, id)
		if err != nil {
			return nil, huma.Error404NotFound("template not found")
		}
		body := dto.TemplateFromGetRow(row)
		if err := s.loadTemplateFragments(ctx, &body, id); err != nil {
			return nil, err
		}
		return &templateOutput{Body: body}, nil
	})
}
