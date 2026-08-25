package api

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/deranjer/loopira/internal/auth"
	"github.com/deranjer/loopira/internal/db"
	"github.com/deranjer/loopira/internal/dto"
	"github.com/deranjer/loopira/internal/ws"
)

type listProjectWorkLogsInput struct {
	ProjectID string `path:"id"`
}

type listProjectWorkLogsOutput struct {
	Body []dto.WorkLog
}

type createWorkLogInput struct {
	ProjectID string `path:"id"`
	Body      struct {
		Title string `json:"title" minLength:"1"`
		Body  string `json:"body" minLength:"1"`
	}
}

type workLogOutput struct {
	Body dto.WorkLog
}

type listWorkLogsInput struct {
	ProjectID string `query:"projectId"`
	AuthorID  string `query:"authorId"`
	Source    string `query:"source" enum:"human,agent"`
	Search    string `query:"search"`
	From      string `query:"from"`
	To        string `query:"to"`
	Limit     int32  `query:"limit" default:"25" minimum:"1" maximum:"100"`
	Offset    int32  `query:"offset" default:"0" minimum:"0"`
}

type listWorkLogsOutput struct {
	Body struct {
		Items []dto.WorkLog `json:"items"`
		Total int32          `json:"total"`
	}
}

// startOfDay parses an optional "2006-01-02" date into a timestamptz at
// midnight; endExclusive shifts it a day forward so a "to" filter can use
// a half-open [from, to) range without ever excluding same-day entries.
func startOfDay(s string, endExclusive bool) (pgtype.Timestamptz, error) {
	if s == "" {
		return pgtype.Timestamptz{}, nil
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return pgtype.Timestamptz{}, err
	}
	if endExclusive {
		t = t.AddDate(0, 0, 1)
	}
	return pgtype.Timestamptz{Time: t, Valid: true}, nil
}

func optionalTextArg(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func (s *Server) registerWorkLogRoutes() {
	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "list-project-work-logs",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{id}/worklogs",
		Summary:     "List a project's work log entries",
		Tags:        []string{"Projects"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *listProjectWorkLogsInput) (*listProjectWorkLogsOutput, error) {
		projectID, err := mustUUID(input.ProjectID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		rows, err := s.q.ListProjectWorkLogs(ctx, projectID)
		if err != nil {
			return nil, err
		}
		out := &listProjectWorkLogsOutput{Body: make([]dto.WorkLog, len(rows))}
		for i, r := range rows {
			out.Body[i] = dto.WorkLogFromProjectRow(r)
		}
		return out, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "create-work-log",
		Method:      http.MethodPost,
		Path:        "/api/v1/projects/{id}/worklogs",
		Summary:     "Add a work log entry to a project. Entries are permanent — there is no edit or delete.",
		Tags:        []string{"Projects"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *createWorkLogInput) (*workLogOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		projectID, err := mustUUID(input.ProjectID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		if _, err := s.q.GetProject(ctx, projectID); err != nil {
			return nil, huma.Error404NotFound("project not found")
		}
		userIDStr, _ := auth.UserID(ctx)
		authorID, err := mustUUID(userIDStr)
		if err != nil {
			return nil, huma.Error401Unauthorized("login required")
		}
		source := "human"
		if auth.ViaAPIKey(ctx) {
			source = "agent"
		}
		created, err := s.q.CreateWorkLog(ctx, db.CreateWorkLogParams{
			ProjectID: projectID,
			AuthorID:  authorID,
			Source:    source,
			Title:     input.Body.Title,
			Body:      input.Body.Body,
		})
		if err != nil {
			return nil, err
		}
		row, err := s.q.GetWorkLog(ctx, created.ID)
		if err != nil {
			return nil, err
		}
		body := dto.WorkLogFromGetRow(row)
		project, err := s.q.GetProject(ctx, projectID)
		if err != nil {
			return nil, err
		}
		s.hub.Broadcast(ws.Event{Type: "worklog.created", TeamID: uid(project.TeamID), Payload: body})
		return &workLogOutput{Body: body}, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "list-work-logs",
		Method:      http.MethodGet,
		Path:        "/api/v1/worklogs",
		Summary:     "List work log entries across all projects, with optional filters and pagination",
		Tags:        []string{"Projects"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *listWorkLogsInput) (*listWorkLogsOutput, error) {
		projectID, err := optionalUUID(strPtr(input.ProjectID))
		if err != nil {
			return nil, huma.Error400BadRequest("invalid projectId")
		}
		authorID, err := optionalUUID(strPtr(input.AuthorID))
		if err != nil {
			return nil, huma.Error400BadRequest("invalid authorId")
		}
		createdFrom, err := startOfDay(input.From, false)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid from date")
		}
		createdTo, err := startOfDay(input.To, true)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid to date")
		}

		limit := input.Limit
		if limit <= 0 {
			limit = 25
		}

		listParams := db.ListWorkLogsParams{
			ProjectID:   projectID,
			AuthorID:    authorID,
			Source:      optionalTextArg(input.Source),
			CreatedFrom: createdFrom,
			CreatedTo:   createdTo,
			Search:      optionalTextArg(input.Search),
			LimitCount:  limit,
			OffsetCount: input.Offset,
		}
		rows, err := s.q.ListWorkLogs(ctx, listParams)
		if err != nil {
			return nil, err
		}
		total, err := s.q.CountWorkLogs(ctx, db.CountWorkLogsParams{
			ProjectID:   listParams.ProjectID,
			AuthorID:    listParams.AuthorID,
			Source:      listParams.Source,
			CreatedFrom: listParams.CreatedFrom,
			CreatedTo:   listParams.CreatedTo,
			Search:      listParams.Search,
		})
		if err != nil {
			return nil, err
		}

		out := &listWorkLogsOutput{}
		out.Body.Items = make([]dto.WorkLog, len(rows))
		for i, r := range rows {
			out.Body.Items[i] = dto.WorkLogFromGlobalRow(r)
		}
		out.Body.Total = total
		return out, nil
	})
}
