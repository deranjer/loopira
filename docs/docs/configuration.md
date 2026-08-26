---
sidebar_position: 3
---

# Configuration

Loopira is configured entirely through environment variables — there is no
config file.

| Variable | Default | Notes |
| --- | --- | --- |
| `DATABASE_URL` | `postgres://postgres:postgres@localhost:5432/loopira?sslmode=disable` | Postgres connection string. |
| `ADDR` | `:8080` | Address the HTTP server listens on. |

Self-hosted deployments have no signup flow: the first time you open the app
with an empty `users` table, it redirects to `/setup`, a one-time wizard
where you set the admin account's name, email, and password. Completing it
also creates the team "Engineering" and a default set of labels. Once an
admin user exists, `/setup` is permanently disabled — it can't be run again.
Everyone else is invited from inside the app.
