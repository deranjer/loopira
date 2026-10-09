// Package mcpserver exposes Loopira's issue tracker to AI agents over the
// Model Context Protocol. It's mounted as a plain http.Handler at /mcp
// (see internal/api/router.go), authenticated the same way as every other
// route — a session cookie or, more relevantly here, a Bearer API key
// (see internal/auth). Tool handlers reuse the same sqlc queries and
// response-shaping (internal/dto's types and *FromRow functions) as the
// REST API, so the two surfaces never drift apart.
package mcpserver

import (
	"context"
	"encoding/json"

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

	addTool(server, &mcp.Tool{
		Name: "list_issues",
		Description: "List issues as compact summaries (no descriptions; use get_issue for the full text). " +
			"Closed (done/canceled) issues are hidden unless includeClosed is set or status is given. " +
			"Narrow with filters or search, and page with limit/offset.",
	}, s.listIssues)

	addTool(server, &mcp.Tool{
		Name:        "get_issue",
		Description: "Get a single issue by id (uuid) or human identifier (e.g. ENG-3).",
	}, s.getIssue)

	addTool(server, &mcp.Tool{
		Name:        "get_issue_history",
		Description: "List every recorded change to an issue, including who made it and when.",
	}, s.getIssueHistory)

	addTool(server, &mcp.Tool{
		Name:        "list_comments",
		Description: "List comments on an issue, oldest first.",
	}, s.listComments)

	addTool(server, &mcp.Tool{
		Name:        "add_comment",
		Description: "Add a comment to an issue. Requires a read-write API key.",
	}, s.addComment)

	addTool(server, &mcp.Tool{
		Name:        "get_issue_links",
		Description: "Get an issue's parent, sub-issues, and blocking dependencies (blockedBy / blocks).",
	}, s.getIssueLinks)

	addTool(server, &mcp.Tool{
		Name:        "set_issue_parent",
		Description: "Make an issue a sub-issue of another, or detach it by omitting parent. Requires a read-write API key.",
	}, s.setIssueParent)

	addTool(server, &mcp.Tool{
		Name:        "set_issue_blocker",
		Description: "Mark an issue as blocked by another issue, or remove that dependency with remove=true. Requires a read-write API key.",
	}, s.setIssueBlocker)

	addTool(server, &mcp.Tool{
		Name:        "create_issue",
		Description: "Create a new issue. Requires a read-write API key.",
	}, s.createIssue)

	addTool(server, &mcp.Tool{
		Name:        "update_issue_status",
		Description: "Change an issue's status. Requires a read-write API key.",
	}, s.updateIssueStatus)

	addTool(server, &mcp.Tool{
		Name: "update_issue",
		Description: "Update an issue's title, description, priority, assignee, project, or " +
			"cycle. Only the fields you supply change — omit the rest. Requires a read-write API key.",
	}, s.updateIssue)

	addTool(server, &mcp.Tool{
		Name:        "list_projects",
		Description: "List projects in the workspace, with completion progress.",
	}, s.listProjects)

	addTool(server, &mcp.Tool{
		Name: "create_project",
		Description: "Create a project in the workspace. Optionally stamp it from a tech-stack template; " +
			"use add_project_guide_fragment afterward to attach a standalone catalog fragment. Requires a read-write API key.",
	}, s.createProject)

	addTool(server, &mcp.Tool{
		Name:        "list_cycles",
		Description: "List cycles (sprints) in the workspace, with completion progress.",
	}, s.listCycles)

	addTool(server, &mcp.Tool{
		Name:        "list_users",
		Description: "List workspace members — use this to resolve a name to an assignee id before calling create_issue/update_issue.",
	}, s.listUsers)

	addTool(server, &mcp.Tool{
		Name:        "list_labels",
		Description: "List labels in the workspace — use this to resolve a label name to an id before calling create_issue/update_issue.",
	}, s.listLabels)

	addTool(server, &mcp.Tool{
		Name: "add_work_log",
		Description: "Add a work log entry to a project's changelog — use this after finishing a session or " +
			"feature to record what was done and why. Entries are permanent: there is no edit or delete. Requires a read-write API key.",
	}, s.addWorkLog)

	addTool(server, &mcp.Tool{
		Name: "list_work_log",
		Description: "List recent work log entries as short previews, newest first, optionally filtered by project or search text. " +
			"Use get_work_log for an entry's full body.",
	}, s.listWorkLog)

	addTool(server, &mcp.Tool{
		Name:        "get_work_log",
		Description: "Get one work log entry, including its full body.",
	}, s.getWorkLog)

	addTool(server, &mcp.Tool{
		Name: "get_project_guide",
		Description: "Get the tech-stack template a project was stamped from (if any) and its full agent guide — the " +
			"per-fragment stack/conventions content, each showing which base fragment and version it came from and " +
			"whether it's been locally modified. Call this first when starting work on a project to pick up its " +
			"stack, conventions, and guardrails.",
	}, s.getProjectGuide)

	addTool(server, &mcp.Tool{
		Name: "add_project_guide_fragment",
		Description: "Attach a catalog guide fragment to a project, or add a custom guide fragment. " +
			"Requires a read-write API key.",
	}, s.addProjectGuideFragment)

	addTool(server, &mcp.Tool{
		Name:        "list_templates",
		Description: "List available tech-stack templates that a project can be stamped from.",
	}, s.listTemplates)

	addTool(server, &mcp.Tool{
		Name:        "get_template",
		Description: "Get a template's description and its ordered composition of fragments.",
	}, s.getTemplate)

	addTool(server, &mcp.Tool{
		Name:        "list_template_fragments",
		Description: "List the catalog of reusable guide fragments (building blocks) that templates are composed from. Names and categories only; content is omitted.",
	}, s.listTemplateFragments)

	return server
}

// addTool registers a typed tool whose result is a single compact-JSON text
// block. The SDK's default for a typed output is to send the same payload
// twice (structuredContent plus a text copy) and to advertise an output
// schema in tools/list; for an LLM client that's pure token overhead, so
// handlers keep their typed signatures but the wire form is text only.
func addTool[In, Out any](server *mcp.Server, tool *mcp.Tool, h func(context.Context, *mcp.CallToolRequest, In) (*mcp.CallToolResult, Out, error)) {
	mcp.AddTool(server, tool, func(ctx context.Context, req *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
		res, out, err := h(ctx, req, in)
		if err != nil || res != nil {
			return res, nil, err
		}
		encoded, err := json.Marshal(out)
		if err != nil {
			return nil, nil, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(encoded)}}}, nil, nil
	})
}
