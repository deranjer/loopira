// Package api wires up HTTP routes: the huma-generated OpenAPI layer under
// /api/v1 (used both by the frontend and, later, by external API clients)
// and the websocket endpoint.
package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/deranjer/loopira/internal/auth"
	"github.com/deranjer/loopira/internal/db"
	"github.com/deranjer/loopira/internal/mcpserver"
	"github.com/deranjer/loopira/internal/storage"
	"github.com/deranjer/loopira/internal/ws"
)

type Server struct {
	Router  *chi.Mux
	q       *db.Queries
	mgr     *auth.Manager
	hub     *ws.Hub
	humaAPI huma.API
	store   storage.Store
}

func New(pool *pgxpool.Pool, hub *ws.Hub, store storage.Store) *Server {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	humaCfg := huma.DefaultConfig("Loopira API", "0.1.0")
	humaCfg.Servers = []*huma.Server{{URL: "/api/v1"}}
	humaAPI := humachi.New(r, humaCfg)

	q := db.New(pool)
	s := &Server{
		Router:  r,
		q:       q,
		mgr:     auth.NewManager(q),
		hub:     hub,
		humaAPI: humaAPI,
		store:   store,
	}

	huma.Register(humaAPI, huma.Operation{
		OperationID: "health",
		Method:      http.MethodGet,
		Path:        "/api/v1/health",
		Summary:     "Health check",
		Tags:        []string{"System"},
	}, func(ctx context.Context, input *struct{}) (*healthOutput, error) {
		out := &healthOutput{}
		out.Body.Status = "ok"
		out.Body.DBConnected = pool.Ping(ctx) == nil
		out.Body.WSClients = hub.ClientCount()
		return out, nil
	})

	s.registerVersionRoutes()
	s.registerSetupRoutes()
	s.registerAuthRoutes()
	s.registerUserRoutes()
	s.registerTeamRoutes()
	s.registerLabelRoutes()
	s.registerProjectRoutes()
	s.registerProjectMemberRoutes()
	s.registerAttachmentRoutes()
	s.registerWorkLogRoutes()
	s.registerCycleRoutes()
	s.registerIssueRoutes()
	s.registerAPIKeyRoutes()
	s.registerViewRoutes()
	s.registerTemplateFragmentRoutes()
	s.registerTemplateRoutes()
	s.registerProjectGuideFragmentRoutes()

	r.With(auth.RequireAuth(s.mgr)).Handle("/ws", hub)
	s.registerMCPRoute(r)
	s.registerDocumentUploadRoute(r)
	s.registerAgentGuideRoute(r)

	return s
}

// protected returns the huma middleware chain that guards non-auth API
// operations behind a valid session.
func (s *Server) protected() huma.Middlewares {
	return huma.Middlewares{requireAuth(s.humaAPI, s.mgr)}
}

// registerMCPRoute mounts the MCP (Model Context Protocol) server at
// /mcp, guarded by the same auth middleware /ws uses — a session cookie
// or, for actual MCP clients (Claude, Cursor, etc.), a Bearer API key.
// Stateless because each request is independently authenticated and there
// is no reason to pin an agent to one server process.
func (s *Server) registerMCPRoute(r *chi.Mux) {
	mcpSrv := mcpserver.New(s.q, s.hub)
	handler := mcpsdk.NewStreamableHTTPHandler(
		func(*http.Request) *mcpsdk.Server { return mcpSrv },
		&mcpsdk.StreamableHTTPOptions{Stateless: true},
	)
	r.With(auth.RequireAuth(s.mgr)).Handle("/mcp", handler)
}

type healthBody struct {
	Status      string `json:"status" example:"ok"`
	DBConnected bool   `json:"dbConnected"`
	WSClients   int    `json:"wsClients"`
}

type healthOutput struct {
	Body healthBody
}
