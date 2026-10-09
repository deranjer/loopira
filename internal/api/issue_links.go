package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/deranjer/loopira/internal/auth"
	"github.com/deranjer/loopira/internal/dto"
	"github.com/deranjer/loopira/internal/issuehistory"
	"github.com/deranjer/loopira/internal/issuelinks"
	"github.com/deranjer/loopira/internal/ws"
)

type issueLinksOutput struct {
	Body dto.IssueLinks
}

type setIssueParentInput struct {
	ID   string `path:"id"`
	Body struct {
		ParentID *string `json:"parentId,omitempty" doc:"issue id of the new parent; omit or null to detach"`
	}
}

type issueBlockerInput struct {
	ID        string `path:"id" doc:"the blocked issue"`
	BlockerID string `path:"blockerId" doc:"the issue that blocks it"`
}

// linkError maps issuelinks validation failures to a 422 and passes
// everything else through as an internal error.
func linkError(err error) error {
	if errors.Is(err, issuelinks.ErrInvalid) {
		return huma.Error422UnprocessableEntity(err.Error())
	}
	return err
}

func (s *Server) registerIssueLinkRoutes() {
	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "get-issue-links",
		Method:      http.MethodGet,
		Path:        "/api/v1/issues/{id}/links",
		Summary:     "Get an issue's parent, sub-issues, and blocking dependencies",
		Tags:        []string{"Issues"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *getIssueInput) (*issueLinksOutput, error) {
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		if _, err := s.q.GetIssue(ctx, id); err != nil {
			return nil, huma.Error404NotFound("issue not found")
		}
		rows, err := s.q.ListIssueLinks(ctx, id)
		if err != nil {
			return nil, err
		}
		return &issueLinksOutput{Body: dto.IssueLinksFromRows(rows)}, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "set-issue-parent",
		Method:      http.MethodPut,
		Path:        "/api/v1/issues/{id}/parent",
		Summary:     "Set or clear an issue's parent",
		Tags:        []string{"Issues"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *setIssueParentInput) (*issueOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		parentID, err := optionalUUID(input.Body.ParentID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid parentId")
		}
		actorID, err := s.actorID(ctx)
		if err != nil {
			return nil, err
		}
		before, err := s.q.GetIssue(ctx, id)
		if err != nil {
			return nil, huma.Error404NotFound("issue not found")
		}
		if _, err := issuelinks.SetParent(ctx, s.q, before, parentID); err != nil {
			return nil, linkError(err)
		}
		row, err := s.q.GetIssue(ctx, id)
		if err != nil {
			return nil, err
		}
		if changes := issuehistory.Diff(before, row); len(changes) > 0 {
			if err := issuehistory.Record(ctx, s.q, id, actorID, "updated", changes); err != nil {
				return nil, err
			}
		}
		body := dto.IssueFromGetRow(row)
		s.hub.Broadcast(ws.Event{Type: "issue.updated", TeamID: uid(row.TeamID), Payload: body})
		return &issueOutput{Body: body}, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "add-issue-blocker",
		Method:      http.MethodPut,
		Path:        "/api/v1/issues/{id}/blockers/{blockerId}",
		Summary:     "Mark an issue as blocked by another issue",
		Tags:        []string{"Issues"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *issueBlockerInput) (*issueLinksOutput, error) {
		return s.changeBlocker(ctx, input, true)
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "remove-issue-blocker",
		Method:      http.MethodDelete,
		Path:        "/api/v1/issues/{id}/blockers/{blockerId}",
		Summary:     "Remove a blocking dependency",
		Tags:        []string{"Issues"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *issueBlockerInput) (*issueLinksOutput, error) {
		return s.changeBlocker(ctx, input, false)
	})
}

func (s *Server) changeBlocker(ctx context.Context, input *issueBlockerInput, add bool) (*issueLinksOutput, error) {
	if !auth.CanWrite(ctx) {
		return nil, huma.Error403Forbidden("read-only API key")
	}
	blockedID, err := mustUUID(input.ID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid id")
	}
	blockerID, err := mustUUID(input.BlockerID)
	if err != nil {
		return nil, huma.Error400BadRequest("invalid blockerId")
	}
	actorID, err := s.actorID(ctx)
	if err != nil {
		return nil, err
	}
	blocked, err := s.q.GetIssue(ctx, blockedID)
	if err != nil {
		return nil, huma.Error404NotFound("issue not found")
	}
	blocker, err := s.q.GetIssue(ctx, blockerID)
	if err != nil {
		return nil, huma.Error404NotFound("blocking issue not found")
	}

	var changed bool
	change := issuehistory.Change{}
	ref := fmt.Sprintf("%s-%d", blocker.TeamKey, blocker.Number)
	if add {
		changed, err = issuelinks.AddBlocker(ctx, s.q, blocker, blocked)
		change.To = ref
	} else {
		changed, err = issuelinks.RemoveBlocker(ctx, s.q, blockerID, blockedID)
		change.From = ref
	}
	if err != nil {
		return nil, linkError(err)
	}
	if changed {
		if err := issuehistory.Record(ctx, s.q, blockedID, actorID, "updated", map[string]issuehistory.Change{"blockedBy": change}); err != nil {
			return nil, err
		}
		s.broadcastIssue(ctx, blockedID)
		s.broadcastIssue(ctx, blockerID)
	}
	rows, err := s.q.ListIssueLinks(ctx, blockedID)
	if err != nil {
		return nil, err
	}
	return &issueLinksOutput{Body: dto.IssueLinksFromRows(rows)}, nil
}

// actorID resolves the calling user, or a 401 if there isn't one.
func (s *Server) actorID(ctx context.Context) (pgtype.UUID, error) {
	userID, _ := auth.UserID(ctx)
	id, err := mustUUID(userID)
	if err != nil {
		return pgtype.UUID{}, huma.Error401Unauthorized("login required")
	}
	return id, nil
}

// broadcastIssue pushes the issue's current state to live clients.
func (s *Server) broadcastIssue(ctx context.Context, id pgtype.UUID) {
	row, err := s.q.GetIssue(ctx, id)
	if err != nil {
		return
	}
	s.hub.Broadcast(ws.Event{Type: "issue.updated", TeamID: uid(row.TeamID), Payload: dto.IssueFromGetRow(row)})
}
