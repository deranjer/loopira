---
sidebar_position: 7
---

# Project layout

```
cmd/server/              entrypoint — wires DB, migrations, router, websocket hub, scheduler
internal/api/            HTTP routes (huma/chi)
internal/auth/           session auth / GitHub OAuth (not yet implemented)
internal/db/             Postgres connection, embedded goose migrations, sqlc-generated queries
internal/db/migrations/  SQL migration files (goose)
internal/db/queries/     hand-written SQL, source for sqlc codegen
internal/jobs/           in-process background job scheduler
internal/storage/        local-disk attachment storage (Store interface, swappable for S3 later)
internal/webapp/         embeds the built frontend (web/ build output lands here)
internal/ws/             in-process websocket broadcast hub
web/                     React/Vite/TypeScript frontend
```
