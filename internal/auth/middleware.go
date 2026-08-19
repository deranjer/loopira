package auth

import (
	"net/http"
)

// RequireAuth validates the session cookie or Bearer API key and injects
// the resolved Identity into the request context, or responds 401 if
// neither is present/valid.
func RequireAuth(m *Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, err := m.IdentityFromRequest(r.Context(), r)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"status":401,"title":"Unauthorized","detail":"login required"}`))
				return
			}
			next.ServeHTTP(w, r.WithContext(WithIdentity(r.Context(), id)))
		})
	}
}
