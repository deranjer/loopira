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

type listProjectGuideFragmentsInput struct {
	ProjectID string `path:"id"`
}

type listProjectGuideFragmentsOutput struct {
	Body []dto.ProjectGuideFragment
}

type projectGuideFragmentOutput struct {
	Body dto.ProjectGuideFragment
}

type addProjectGuideFragmentInput struct {
	ProjectID string `path:"id"`
	Body      struct {
		FragmentID *string `json:"fragmentId,omitempty"`
		Name       string  `json:"name,omitempty"`
		Content    string  `json:"content,omitempty"`
	}
}

type updateProjectGuideFragmentInput struct {
	ProjectID string `path:"id"`
	ID        string `path:"fragmentInstanceId"`
	Body      struct {
		Name    string `json:"name" minLength:"1"`
		Content string `json:"content,omitempty"`
	}
}

type deleteProjectGuideFragmentInput struct {
	ProjectID string `path:"id"`
	ID        string `path:"fragmentInstanceId"`
}

type resetProjectGuideFragmentInput struct {
	ProjectID string `path:"id"`
	ID        string `path:"fragmentInstanceId"`
}

type reorderProjectGuideFragmentsInput struct {
	ProjectID string `path:"id"`
	Body      struct {
		IDs []string `json:"ids"`
	}
}

func (s *Server) registerProjectGuideFragmentRoutes() {
	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "list-project-guide-fragments",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{id}/guide-fragments",
		Summary:     "List a project's agent guide fragments, in order",
		Tags:        []string{"Projects"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *listProjectGuideFragmentsInput) (*listProjectGuideFragmentsOutput, error) {
		projectID, err := mustUUID(input.ProjectID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		rows, err := s.q.ListProjectGuideFragments(ctx, projectID)
		if err != nil {
			return nil, err
		}
		out := &listProjectGuideFragmentsOutput{Body: make([]dto.ProjectGuideFragment, len(rows))}
		for i, f := range rows {
			out.Body[i] = dto.ProjectGuideFragmentFromRow(f)
		}
		return out, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "add-project-guide-fragment",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{id}/guide-fragments",
		Summary:     "Add a guide fragment to a project, either from the catalog or fully custom",
		Tags:        []string{"Projects"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *addProjectGuideFragmentInput) (*projectGuideFragmentOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		projectID, err := mustUUID(input.ProjectID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}

		var (
			fragmentID  pgtype.UUID
			name        string
			content     string
			baseVersion pgtype.Int4
		)
		if input.Body.FragmentID != nil && *input.Body.FragmentID != "" {
			fid, err := mustUUID(*input.Body.FragmentID)
			if err != nil {
				return nil, huma.Error400BadRequest("invalid fragmentId")
			}
			fragment, err := s.q.GetTemplateFragment(ctx, fid)
			if err != nil {
				return nil, huma.Error404NotFound("fragment not found")
			}
			fragmentID = fid
			name = fragment.Name
			content = fragment.Content
			baseVersion = pgtype.Int4{Int32: fragment.Version, Valid: true}
		} else {
			if input.Body.Name == "" {
				return nil, huma.Error400BadRequest("name is required for a custom guide fragment")
			}
			name = input.Body.Name
			content = input.Body.Content
		}

		created, err := s.q.AddProjectGuideFragment(ctx, db.AddProjectGuideFragmentParams{
			ProjectID:   projectID,
			FragmentID:  fragmentID,
			Name:        name,
			Content:     content,
			BaseVersion: baseVersion,
		})
		if err != nil {
			return nil, err
		}
		return &projectGuideFragmentOutput{Body: dto.ProjectGuideFragmentFromRow(created)}, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "update-project-guide-fragment",
		Method:      http.MethodPatch,
		Path:        "/api/v1/projects/{id}/guide-fragments/{fragmentInstanceId}",
		Summary:     "Edit a project's copy of a guide fragment (marks it locally modified if it has a base)",
		Tags:        []string{"Projects"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *updateProjectGuideFragmentInput) (*projectGuideFragmentOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		updated, err := s.q.UpdateProjectGuideFragment(ctx, db.UpdateProjectGuideFragmentParams{
			ID:      id,
			Name:    input.Body.Name,
			Content: input.Body.Content,
		})
		if err != nil {
			return nil, huma.Error404NotFound("guide fragment not found")
		}
		return &projectGuideFragmentOutput{Body: dto.ProjectGuideFragmentFromRow(updated)}, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "delete-project-guide-fragment",
		Method:      http.MethodDelete,
		Path:        "/api/v1/projects/{id}/guide-fragments/{fragmentInstanceId}",
		Summary:     "Remove a guide fragment from a project",
		Tags:        []string{"Projects"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *deleteProjectGuideFragmentInput) (*statusOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		if err := s.q.DeleteProjectGuideFragment(ctx, id); err != nil {
			return nil, err
		}
		out := &statusOutput{}
		out.Body.Status = "ok"
		return out, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "reset-project-guide-fragment",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{id}/guide-fragments/{fragmentInstanceId}/reset",
		Summary:     "Reset a project's copy of a guide fragment back to its base fragment's current content",
		Tags:        []string{"Projects"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *resetProjectGuideFragmentInput) (*projectGuideFragmentOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		reset, err := s.q.ResetProjectGuideFragmentToBase(ctx, id)
		if err != nil {
			return nil, huma.Error404NotFound("guide fragment has no base to reset to")
		}
		return &projectGuideFragmentOutput{Body: dto.ProjectGuideFragmentFromRow(reset)}, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "reorder-project-guide-fragments",
		Method:      http.MethodPatch,
		Path:        "/api/v1/projects/{id}/guide-fragments/reorder",
		Summary:     "Set the display order of a project's guide fragments",
		Tags:        []string{"Projects"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *reorderProjectGuideFragmentsInput) (*listProjectGuideFragmentsOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		projectID, err := mustUUID(input.ProjectID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		for i, idStr := range input.Body.IDs {
			id, err := mustUUID(idStr)
			if err != nil {
				return nil, huma.Error400BadRequest("invalid ids entry")
			}
			if err := s.q.SetProjectGuideFragmentPosition(ctx, db.SetProjectGuideFragmentPositionParams{
				ID:       id,
				Position: int32(i),
			}); err != nil {
				return nil, err
			}
		}
		rows, err := s.q.ListProjectGuideFragments(ctx, projectID)
		if err != nil {
			return nil, err
		}
		out := &listProjectGuideFragmentsOutput{Body: make([]dto.ProjectGuideFragment, len(rows))}
		for i, f := range rows {
			out.Body[i] = dto.ProjectGuideFragmentFromRow(f)
		}
		return out, nil
	})
}
