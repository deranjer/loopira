package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/deranjer/loopira/internal/auth"
)

// registerAgentGuideRoute mounts the raw markdown-serving endpoint an AI
// coding agent/CLI can be pointed at directly — like /documents, this
// returns a non-JSON body so it's a raw chi route rather than a huma
// operation, guarded by the same auth /ws and /mcp use.
func (s *Server) registerAgentGuideRoute(r *chi.Mux) {
	r.With(auth.RequireAuth(s.mgr)).Get("/api/v1/projects/{id}/agents.md", s.handleServeAgentGuide)
}

func (s *Server) handleServeAgentGuide(w http.ResponseWriter, r *http.Request) {
	projectID, err := mustUUID(chi.URLParam(r, "id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "invalid project id")
		return
	}
	fragments, err := s.q.ListProjectGuideFragments(r.Context(), projectID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "failed to load guide")
		return
	}

	var b strings.Builder
	for i, f := range fragments {
		if i > 0 {
			b.WriteString("\n\n")
		}
		fmt.Fprintf(&b, "## %s\n\n%s", f.Name, f.Content)
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(b.String()))
}
