# Loopira

A self-hosted, single-workspace issue tracker (Linear clone). Architecture and
feature scope are documented in the plan this scaffold was built from.

## Stack

- **Backend**: Go (`chi` + `huma` for the REST/OpenAPI API, `pgx` for Postgres,
  in-process websocket hub for real-time updates, in-process scheduler for
  background jobs). Compiles to a single binary that also serves the built
  frontend and auto-runs database migrations on startup.
- **Frontend**: React + Vite + TypeScript, Tailwind, TanStack Query, React
  Router.
- **Database**: PostgreSQL.
- **Deployment**: one app container + one Postgres container. No Redis, no
  S3/MinIO — attachments live on a mounted local disk, real-time fan-out and
  job scheduling run in-process.

## Running locally with Docker (closest to production)

```sh
docker compose up --build
```

This builds the frontend, embeds it into the Go binary, builds the app
image, and starts both containers. Migrations run automatically on boot.

- App: http://localhost:8080
- Health check: http://localhost:8080/api/v1/health

## Running for development (hot reload)

Two terminals:

```sh
# Terminal 1 — backend (requires a running Postgres; e.g. `docker compose up db`)
go run ./cmd/server

# Terminal 2 — frontend, dev server proxies /api and /ws to :8080
cd web
pnpm install
pnpm dev
```

## Project layout

```
cmd/server/         entrypoint — wires DB, migrations, router, websocket hub, scheduler
internal/api/        HTTP routes (huma/chi)
internal/auth/       session auth / GitHub OAuth (not yet implemented)
internal/db/          Postgres connection, embedded goose migrations, sqlc-generated queries
internal/db/migrations/  SQL migration files (goose)
internal/db/queries/     hand-written SQL, source for sqlc codegen
internal/jobs/        in-process background job scheduler
internal/storage/     local-disk attachment storage (Store interface, swappable for S3 later)
internal/webapp/      embeds the built frontend (web/ build output lands here)
internal/ws/           in-process websocket broadcast hub
web/                  React/Vite/TypeScript frontend
```

## Connecting an AI agent (MCP)

Loopira exposes an MCP (Model Context Protocol) server at `/mcp` so agents
like Claude, Cursor, or ChatGPT can read and manage issues directly.

1. In the app, go to **Settings → API Keys → New API key**. Choose
   read-write (can create/edit issues) or read-only. The plaintext key is
   shown once — copy it.
2. Connect your MCP client. For Claude Code:
   ```sh
   claude mcp add --transport http loopira http://localhost:8080/mcp \
     --header "Authorization: Bearer <your key>"
   ```
   Other clients (Cursor, ChatGPT, etc.) take the same URL and header —
   check their MCP configuration docs for the exact syntax.

The same key also works as a Bearer token against the regular REST API
(`/api/v1/...`), for scripts that don't need the full MCP tool surface.

## Database changes

Add a new goose migration file to `internal/db/migrations/`, then write
queries in `internal/db/queries/*.sql` and run:

```sh
go run github.com/sqlc-dev/sqlc/cmd/sqlc generate
```

to regenerate typed Go query code into `internal/db`.

## Status

This is the initial repo scaffold: it proves out the architecture end-to-end
(DB connection + auto-migration, API + OpenAPI, websocket endpoint, embedded
frontend build) but implements no product features yet. See the project plan
for the full v1 feature scope and build order.
