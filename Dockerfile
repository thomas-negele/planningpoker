# syntax=docker/dockerfile:1

# Build the frontend, embed it in the Go binary, then ship the binary alone.
# Image tags are mutable; use digest pins if reproducible rebuilds are required.

# Frontend build.
FROM node:26-alpine AS frontend
WORKDIR /src/web

# Copy manifests first so dependency installation can use the build cache.
COPY web/package.json web/package-lock.json ./
# npm ci installs the dependency versions from the lockfile.
RUN npm ci

COPY web/ ./
# Vite writes into the Go package because go:embed cannot include parent paths.
RUN npm run build

# Go binary.
FROM golang:1.27-alpine AS backend
WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY --from=frontend /src/internal/webassets/dist ./internal/webassets/dist

# Disable CGO for the static runtime. embedassets includes the frontend and CSP;
# -s -w removes symbol and debug information.
RUN CGO_ENABLED=0 GOOS=linux go build \
      -tags embedassets \
      -ldflags="-s -w" \
      -o /out/planningpoker \
      ./cmd/planningpoker

# Minimal runtime without a shell or package manager.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=backend /out/planningpoker /planningpoker

# EXPOSE documents the port; it does not publish it on the host.
EXPOSE 8080

# Explicitly retain the base image's nonroot user.
USER nonroot:nonroot

ENTRYPOINT ["/planningpoker"]
