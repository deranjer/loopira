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
| `ADMIN_EMAIL` | `admin@loopira.local` | Email for the seeded admin user. Only consulted on the very first boot, when the `users` table is empty. |
| `ADMIN_PASSWORD` | *(random)* | Password for the seeded admin user. If unset, a random password is generated and printed to the container logs **once**, on first boot — it is never shown again. Set this explicitly for unattended deployments, or watch the logs on first run. |

Self-hosted deployments have no signup flow: the admin user created on first
boot (with the team "Engineering" and a default set of labels) is the only
way in. Everyone else is invited from inside the app.
