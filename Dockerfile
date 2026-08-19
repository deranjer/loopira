FROM node:22-alpine AS frontend
WORKDIR /app/web
COPY web/package.json web/pnpm-lock.yaml ./
RUN corepack enable && corepack prepare pnpm@latest --activate && pnpm install --frozen-lockfile
COPY web/ ./
RUN pnpm build

FROM golang:1.25-alpine AS backend
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=frontend /app/internal/webapp/dist ./internal/webapp/dist
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian12
COPY --from=backend /out/server /server
EXPOSE 8080
ENTRYPOINT ["/server"]
