package api

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-chi/chi/v5"

	"github.com/deranjer/loopira/internal/auth"
	"github.com/deranjer/loopira/internal/db"
	"github.com/deranjer/loopira/internal/dto"
)

// randomKey generates a filesystem-safe random identifier used to
// namespace an uploaded file's storage key, avoiding collisions without
// needing the attachment's DB-generated id up front.
func randomKey() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return hex.EncodeToString(b)
}

// Uploading and downloading files aren't a good fit for huma's typed-JSON
// request/response model, so — like /ws and /mcp — they're registered as
// raw chi handlers guarded by the same auth.RequireAuth middleware.

const maxUploadSize = 25 << 20 // 25 MiB

type listDocumentsInput struct {
	ProjectID string `path:"id"`
}

type listDocumentsOutput struct {
	Body []dto.Attachment
}

type deleteAttachmentInput struct {
	ID string `path:"id"`
}

func (s *Server) registerAttachmentRoutes() {
	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "list-project-documents",
		Method:      http.MethodGet,
		Path:        "/api/v1/projects/{id}/documents",
		Summary:     "List a project's uploaded documents",
		Tags:        []string{"Projects"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *listDocumentsInput) (*listDocumentsOutput, error) {
		projectID, err := mustUUID(input.ProjectID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		rows, err := s.q.ListProjectAttachments(ctx, projectID)
		if err != nil {
			return nil, err
		}
		out := &listDocumentsOutput{Body: make([]dto.Attachment, len(rows))}
		for i, a := range rows {
			out.Body[i] = dto.AttachmentFromRow(a)
		}
		return out, nil
	})

	huma.Register(s.humaAPI, huma.Operation{
		OperationID: "delete-attachment",
		Method:      http.MethodDelete,
		Path:        "/api/v1/attachments/{id}",
		Summary:     "Delete an uploaded document",
		Tags:        []string{"Projects"},
		Middlewares: s.protected(),
	}, func(ctx context.Context, input *deleteAttachmentInput) (*statusOutput, error) {
		if !auth.CanWrite(ctx) {
			return nil, huma.Error403Forbidden("read-only API key")
		}
		id, err := mustUUID(input.ID)
		if err != nil {
			return nil, huma.Error400BadRequest("invalid id")
		}
		attachment, err := s.q.GetAttachment(ctx, id)
		if err != nil {
			return nil, huma.Error404NotFound("attachment not found")
		}
		if err := s.q.DeleteAttachment(ctx, id); err != nil {
			return nil, err
		}
		if err := s.store.Delete(ctx, attachment.StorageKey); err != nil {
			return nil, err
		}
		out := &statusOutput{}
		out.Body.Status = "ok"
		return out, nil
	})
}

// registerDocumentUploadRoute mounts the raw multipart-upload and
// binary-download handlers, guarded by the same session/API-key auth /ws
// and /mcp use.
func (s *Server) registerDocumentUploadRoute(r *chi.Mux) {
	r.With(auth.RequireAuth(s.mgr)).Post("/api/v1/projects/{id}/documents", s.handleUploadDocument)
	r.With(auth.RequireAuth(s.mgr)).Get("/api/v1/attachments/{id}/download", s.handleDownloadAttachment)
}

func writeJSONError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"detail": msg})
}

func (s *Server) handleUploadDocument(w http.ResponseWriter, r *http.Request) {
	if !auth.CanWrite(r.Context()) {
		writeJSONError(w, http.StatusForbidden, "read-only API key")
		return
	}
	projectID, err := mustUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	userIDStr, _ := auth.UserID(r.Context())
	uploadedBy, err := mustUUID(userIDStr)
	if err != nil {
		writeJSONError(w, http.StatusUnauthorized, "login required")
		return
	}

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		writeJSONError(w, http.StatusBadRequest, "file too large or malformed upload")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "missing file field")
		return
	}
	defer file.Close()

	storageKey := fmt.Sprintf("%s/%s", randomKey(), header.Filename)
	size, err := s.store.Save(r.Context(), storageKey, file)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to store file")
		return
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	created, err := s.q.CreateAttachment(r.Context(), db.CreateAttachmentParams{
		ProjectID:   projectID,
		Filename:    header.Filename,
		ContentType: contentType,
		SizeBytes:   size,
		StorageKey:  storageKey,
		UploadedBy:  uploadedBy,
	})
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to save attachment record")
		return
	}

	rows, err := s.q.ListProjectAttachments(r.Context(), projectID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to reload documents")
		return
	}
	for _, a := range rows {
		if a.ID == created.ID {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(dto.AttachmentFromRow(a))
			return
		}
	}
	writeJSONError(w, http.StatusInternalServerError, "attachment not found after create")
}

func (s *Server) handleDownloadAttachment(w http.ResponseWriter, r *http.Request) {
	id, err := mustUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid attachment id")
		return
	}
	attachment, err := s.q.GetAttachment(r.Context(), id)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "attachment not found")
		return
	}
	f, err := s.store.Open(r.Context(), attachment.StorageKey)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "file not found")
		return
	}
	defer f.Close()

	w.Header().Set("Content-Type", attachment.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", attachment.Filename))
	_, _ = io.Copy(w, f)
}
