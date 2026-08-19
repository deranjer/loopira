---
sidebar_position: 6
---

# Deployment

Tagged releases are built and published to the GitHub Container Registry as
a single multi-arch image (`linux/amd64` and `linux/arm64`):

```
ghcr.io/deranjer/loopira
```

## Tags

Releases are cut by pushing a git tag like `vX.Y.Z`; the published image tags
drop the `v` prefix:

| Tag | Meaning |
| --- | --- |
| `X.Y.Z` | An exact release, e.g. `1.2.3` (from git tag `v1.2.3`). |
| `X.Y` | Latest patch release within that minor version. |
| `X` | Latest release within that major version. |
| `latest` | Latest tagged release. |
| `sha-<short-sha>` | The exact commit the image was built from. |

## Running the published image

```sh
docker run -d \
  -p 8080:8080 \
  -e DATABASE_URL="postgres://postgres:postgres@<db-host>:5432/loopira?sslmode=disable" \
  -e ADMIN_EMAIL="you@example.com" \
  -e ADMIN_PASSWORD="<a strong password>" \
  -v loopira_attachments:/data/attachments \
  ghcr.io/deranjer/loopira:latest
```

See [Configuration](./configuration.md) for the full list of environment
variables.

## Using it with Docker Compose

See [Installation](./installation.md) for a full `docker-compose.yml` using
the published image. For reproducible production deployments, pin the
`app` service to a specific version tag (e.g. `1.2.3`) instead of `latest`.
