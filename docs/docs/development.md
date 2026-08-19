---
sidebar_position: 4
---

# Development

## Running with hot reload

Two terminals:

```sh
# Terminal 1 — backend (requires a running Postgres, e.g. `docker compose up db`)
go run ./cmd/server

# Terminal 2 — frontend; the dev server proxies /api and /ws to :8080
cd web
pnpm install
pnpm dev
```

## Making database changes

1. Add a new [goose](https://github.com/pressly/goose) migration file to
   `internal/db/migrations/`.
2. Write or update queries in `internal/db/queries/*.sql`.
3. Regenerate the typed Go query code:

   ```sh
   go run github.com/sqlc-dev/sqlc/cmd/sqlc generate
   ```

   This regenerates `internal/db` from your SQL using
   [sqlc](https://sqlc.dev/).
