// Package mcpserver exposes Loopira's issue tracker to AI agents over the
// Model Context Protocol. It's mounted as a plain http.Handler at /mcp
// (see internal/api/router.go), authenticated the same way as every other
// route — a session cookie or, more relevantly here, a Bearer API key
// (see internal/auth). Tool handlers reuse the same sqlc queries and
// response-shaping (internal/api's exported *Body types and *FromRow
// functions) as the REST API, so the two surfaces never drift apart.
package mcpserver

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/deranjer/loopira/internal/db"
	"github.com/deranjer/loopira/internal/ws"
)

type toolServer struct {
	q   *db.Queries
	hub *ws.Hub
}

// New builds the MCP server and registers every tool. v1 is
// single-workspace, so tools resolve "the" team internally rather than
// taking a teamId argument — see currentTeam.
func New(q *db.Queries, hub *ws.Hub) *mcp.Server {
	s := &toolServer{q: q, hub: hub}
	server := mcp.NewServer(&mcp.Implementation{Name: "loopira", Version: "0.1.0"}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_issues",
		Description: "List issues in the workspace, optionally filtered by status, project, or cycle.",
	}, s.listIssues)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_issue",
		Description: "Get a single issue by id (uuid) or human identifier (e.g. ENG-3).",
	}, s.getIssue)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_issue",
		Description: "Create a new issue. Requires a read-write API key.",
	}, s.createIssue)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_issue_status",
		Description: "Change an issue's status. Requires a read-write API key.",
	}, s.updateIssueStatus)

	mcp.AddTool(server, &mcp.Tool{
		Name: "update_issue",
		Description: "Update an issue's title, description, priority, assignee, project, or " +
			"cycle. Only the fields you supply change — omit the rest. Requires a read-write API key.",
	}, s.updateIssue)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_projects",
		Description: "List projects in the workspace, with completion progress.",
	}, s.listProjects)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_cycles",
		Description: "List cycles (sprints) in the workspace, with completion progress.",
	}, s.listCycles)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_users",
		Description: "List workspace members — use this to resolve a name to an assignee id before calling create_issue/update_issue.",
	}, s.listUsers)

	return server
}
