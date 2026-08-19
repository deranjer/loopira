---
sidebar_position: 2
---

# Installation

The quickest way to run Loopira — and the closest to how it runs in
production — is with Docker Compose.

```sh
docker compose up --build
```

This builds the frontend, embeds it into the Go binary, builds the app
image, and starts both containers (`app` and `db`). Database migrations run
automatically on boot — there's no separate migration step for a fresh
deployment.

Once it's up:

- **App**: [http://localhost:8080](http://localhost:8080)
- **Health check**: [http://localhost:8080/api/v1/health](http://localhost:8080/api/v1/health)
- **Postgres**: exposed on host port `5433` (mapped from the container's
  `5432`, to avoid clashing with any Postgres you may already have running
  locally on `5432`)

Data persists in two named Docker volumes: `db_data` (Postgres) and
`attachments` (uploaded files).

On first boot, Loopira seeds an initial team and an admin user — see
[Configuration](./configuration.md) for how to control the admin email and
password, and [Deployment](./deployment.md) for running the published image
instead of building locally.
