---
sidebar_position: 2
---

# Installation

Loopira ships as a single container image plus Postgres — there's nothing to
build. Create a `docker-compose.yml`:

```yaml
name: loopira

services:
  app:
    image: ghcr.io/deranjer/loopira:latest
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://postgres:postgres@db:5432/loopira?sslmode=disable
      ADDR: ":8080"
    volumes:
      - attachments:/data/attachments
    depends_on:
      db:
        condition: service_healthy
    restart: unless-stopped

  db:
    image: postgres:17-alpine
    environment:
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: postgres
      POSTGRES_DB: loopira
    volumes:
      - db_data:/var/lib/postgresql/data
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U postgres"]
      interval: 5s
      timeout: 5s
      retries: 10
    restart: unless-stopped

volumes:
  db_data:
  attachments:
```

Then:

```sh
docker compose up -d
```

Database migrations run automatically on boot — there's no separate
migration step. Once it's up:

- **App**: [http://localhost:8080](http://localhost:8080)
- **Health check**: [http://localhost:8080/api/v1/health](http://localhost:8080/api/v1/health)

Data persists in two named Docker volumes: `db_data` (Postgres) and
`attachments` (uploaded files).

The first time you open the app, it redirects to `/setup`, a one-time wizard
for creating the admin account — see [Configuration](./configuration.md) for
details. For pinning to a specific version instead of `latest`, see
[Deployment](./deployment.md).
