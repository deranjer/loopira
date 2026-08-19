---
slug: /
sidebar_position: 1
---

# Introduction

Loopira is a self-hosted, single-workspace issue tracker in the spirit of
Linear.

- **Backend**: Go, compiled to a single binary. It serves the REST/OpenAPI
  API (`chi` + `huma`), a websocket endpoint for real-time updates, and an
  in-process job scheduler — all in one process, with no external
  dependencies like Redis.
- **Frontend**: React, Vite, and TypeScript. The production frontend build
  is embedded directly into the Go binary via `go:embed`, so the backend
  binary is the only artifact you need to run the app.
- **Database**: PostgreSQL. Schema migrations run automatically on startup.
- **Deployment**: one app container plus one Postgres container. Attachments
  are stored on a mounted local disk — no S3/MinIO required.

## Where to go next

- [Installation](./installation.md) — run Loopira with Docker Compose.
- [Configuration](./configuration.md) — environment variables.
- [Development](./development.md) — run the backend and frontend locally
  with hot reload, and how to make database changes.
- [MCP integration](./mcp-integration.md) — connect an AI agent (Claude,
  Cursor, etc.) directly to your Loopira workspace.
- [Deployment](./deployment.md) — run the published Docker image in
  production.
- [Project layout](./project-layout.md) — a tour of the source tree.
